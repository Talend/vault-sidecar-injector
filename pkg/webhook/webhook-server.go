// Copyright © 2019-2021 Talend - www.talend.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"
	m "talend/vault-sidecar-injector/pkg/mode"

	admv1 "k8s.io/api/admission/v1"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

// New : init new VaultInjector type
func New(config *cfg.VSIConfig, server *http.Server) *VaultInjector {
	// Add mode annotations
	for _, mode := range m.VaultInjectorModes {
		vaultInjectorAnnotationKeys = append(vaultInjectorAnnotationKeys, mode.Annotations...)
	}

	// Compute FQ annotations
	config.VaultInjectorAnnotationsFQ = make(map[string]string, len(vaultInjectorAnnotationKeys))
	for _, vaultAnnotationKey := range vaultInjectorAnnotationKeys {
		if config.VaultInjectorAnnotationKeyPrefix != "" {
			config.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = config.VaultInjectorAnnotationKeyPrefix + "/" + vaultAnnotationKey
		} else {
			config.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = vaultAnnotationKey
		}
	}

	return &VaultInjector{
		VSIConfig: config,
		Server:    server,
	}
}

// Create mutation patch for resources
func (vaultInjector *VaultInjector) createPatch(pod *corev1.Pod, annotations map[string]string) ([]byte, error) {

	patchPodSpec, err := vaultInjector.updatePodSpec(pod)
	if err != nil {
		return nil, err
	}

	var patch []ctx.PatchOperation

	patch = append(patch, patchPodSpec...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}

// Main mutation process
func (vaultInjector *VaultInjector) mutate(ar *admv1.AdmissionReview) *admv1.AdmissionResponse {
	var pod corev1.Pod
	var podName, podNamespace string

	req := ar.Request

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return &admv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("Pod=%+v", pod)
	}

	if pod.Name == "" {
		podName = pod.GenerateName
	} else {
		podName = pod.Name
	}

	if pod.Namespace == "" {
		podNamespace = metav1.NamespaceDefault
	} else {
		podNamespace = pod.Namespace
	}

	klog.Infof("AdmissionReview '%v' for '%+v', Namespace=%v Name='%v (%s/%s)' UID=%v patchOperation=%v",
		ar.GroupVersionKind(), req.Kind, req.Namespace, req.Name, podNamespace, podName, req.UID, req.Operation)

	// Determine whether to perform mutation
	if !mutationRequired(ignoredNamespaces, vaultInjector.VaultInjectorAnnotationsFQ, &pod.ObjectMeta) {
		klog.Infof("Skipping mutation for %s/%s due to policy check", podNamespace, podName)
		return &admv1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{vaultInjector.VaultInjectorAnnotationsFQ[ctx.VaultInjectorAnnotationStatusKey]: ctx.VaultInjectorStatusInjected}
	patchBytes, err := vaultInjector.createPatch(&pod, annotations)
	if err != nil {
		return &admv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	klog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &admv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admv1.PatchType {
			pt := admv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook server
func (vaultInjector *VaultInjector) Serve(w http.ResponseWriter, r *http.Request) {
	var body, response []byte
	var returnedAR interface{}

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("HTTP Request=%+v", r)
	}

	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		klog.Error("Empty body")
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "Invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	rawAR, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		klog.Errorf("Can't decode body: %v", err)
		http.Error(w, fmt.Sprintf("Cannot decode body: %v", err), http.StatusBadRequest)
		return
	}

	// Webhook can currently receive either v1 or v1beta1 AdmissionReview objects.
	// v1 is the default, internal version in use by VSI (v1beta1 support will be removed).
	// If v1beta1 is received, it will be converted into v1 and back to v1beta1 in response
	// (as spec states response should use same version as request)
	arVersion := admv1.SchemeGroupVersion
	ar, isV1 := rawAR.(*admv1.AdmissionReview)
	if !isV1 {
		arVersion = admv1beta1.SchemeGroupVersion
		arv1beta1, isv1beta1 := rawAR.(*admv1beta1.AdmissionReview)
		if !isv1beta1 {
			klog.Errorf("Unsupported AdmissionReview version %v", gvk.Version)
			http.Error(w, fmt.Sprintf("Unsupported AdmissionReview version %v", gvk.Version), http.StatusBadRequest)
			return
		}

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Received AdmissionReview '%v' Request=%+v", arv1beta1.GroupVersionKind(), arv1beta1.Request)
		}

		// Convert v1beta1 to v1
		ar = &admv1.AdmissionReview{}
		Convert_v1beta1_AdmissionReview_To_admission_AdmissionReview(arv1beta1, ar)
	} else {
		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Received AdmissionReview '%v' Request=%+v", ar.GroupVersionKind(), ar.Request)
		}
	}

	ar.Response = vaultInjector.mutate(ar)
	returnedAR = ar

	// If v1 received
	if arVersion.Version == admv1.SchemeGroupVersion.Version {
		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Returned AdmissionReview '%v' Response=%+v", ar.GroupVersionKind(), ar.Response)
		}
	} else { // If v1beta1 received: convert v1 to v1beta1
		arv1beta1 := &admv1beta1.AdmissionReview{}
		Convert_admission_AdmissionReview_To_v1beta1_AdmissionReview(ar, arv1beta1)
		returnedAR = arv1beta1

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Returned AdmissionReview '%v' Response=%+v", arv1beta1.GroupVersionKind(), arv1beta1.Response)
		}
	}

	response, err = json.Marshal(returnedAR)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("Cannot encode response: %v", err), http.StatusInternalServerError)
	}

	klog.Infof("Write reponse ...")
	if _, err := w.Write(response); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
	}
}
