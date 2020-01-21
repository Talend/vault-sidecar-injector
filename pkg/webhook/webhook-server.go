// Copyright © 2019 Talend
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
	"talend/vault-sidecar-injector/pkg/config"

	"k8s.io/api/admission/v1beta1"
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
func New(cfg *config.VSIConfig, server *http.Server) *VaultInjector {
	// Compute FQ annotations
	cfg.VaultInjectorAnnotationsFQ = make(map[string]string, len(vaultInjectorAnnotationKeys))
	for _, vaultAnnotationKey := range vaultInjectorAnnotationKeys {
		if cfg.VaultInjectorAnnotationKeyPrefix != "" {
			cfg.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = cfg.VaultInjectorAnnotationKeyPrefix + "/" + vaultAnnotationKey
		} else {
			cfg.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = vaultAnnotationKey
		}
	}

	return &VaultInjector{
		VSIConfig: cfg,
		Server:    server,
		// Register Modes functions (using trick to be able to use function pointers to methods, see https://stackoverflow.com/a/31561683)
		ModesFunc: map[string]func(vaultInjector *VaultInjector, labels, annotations map[string]string) (modeConfig, error){
			vaultInjectorModeSecrets: (*VaultInjector).secretsMode,
			vaultInjectorModeProxy:   (*VaultInjector).proxyMode,
		},
	}
}

// create mutation patch for resoures
func (vaultInjector *VaultInjector) createPatch(pod *corev1.Pod, annotations map[string]string) ([]byte, error) {

	patchPodSpec, err := vaultInjector.updatePodSpec(pod)
	if err != nil {
		return nil, err
	}

	var patch []patchOperation

	patch = append(patch, patchPodSpec...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}

// main mutation process
func (vaultInjector *VaultInjector) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	var podName string
	var podNamespace string

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("Request=%+v", req)
	}

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
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

	klog.Infof("AdmissionReview for GroupVersionKind=%+v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%+v",
		req.Kind, req.Namespace, req.Name, podName, req.UID, req.Operation, req.UserInfo)

	// determine whether to perform mutation
	if !mutationRequired(ignoredNamespaces, vaultInjector.VaultInjectorAnnotationsFQ, &pod.ObjectMeta) {
		klog.Infof("Skipping mutation for %s/%s due to policy check", podNamespace, podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationStatusKey]: vaultInjectorStatusInjected}
	patchBytes, err := vaultInjector.createPatch(&pod, annotations)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	klog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook server
func (vaultInjector *VaultInjector) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = vaultInjector.mutate(&ar)
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
