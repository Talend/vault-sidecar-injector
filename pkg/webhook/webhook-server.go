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
	m "talend/vault-sidecar-injector/pkg/mode"

	admv1 "k8s.io/api/admission/v1"
	admv1beta1 "k8s.io/api/admission/v1beta1"
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

// Serve method for webhook server
func (vaultInjector *VaultInjector) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte

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
	arInVersion := admv1.SchemeGroupVersion
	arIn, isV1 := rawAR.(*admv1.AdmissionReview)
	if !isV1 {
		arInVersion = admv1beta1.SchemeGroupVersion
		arInv1beta1, isv1beta1 := rawAR.(*admv1beta1.AdmissionReview)
		if !isv1beta1 {
			klog.Errorf("Unsupported AdmissionReview version %v", gvk.Version)
			http.Error(w, fmt.Sprintf("Unsupported AdmissionReview version %v", gvk.Version), http.StatusBadRequest)
			return
		}

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Received AdmissionReview '%v' Request=%+v", arInv1beta1.GroupVersionKind(), arInv1beta1.Request)
		}

		// Convert v1beta1 to v1
		arIn = &admv1.AdmissionReview{}
		arIn.SetGroupVersionKind(admv1.SchemeGroupVersion.WithKind("AdmissionReview"))
		Convert_v1beta1_AdmissionReview_To_admission_AdmissionReview(arInv1beta1, arIn)

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Converted AdmissionReview '%v' Request=%+v", arIn.GroupVersionKind(), arIn.Request)
		}
	} else {
		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Received AdmissionReview '%v' Request=%+v", arIn.GroupVersionKind(), arIn.Request)
		}
	}

	arOut := &admv1.AdmissionReview{}
	arOut.SetGroupVersionKind(admv1.SchemeGroupVersion.WithKind("AdmissionReview"))
	arOut.Response = vaultInjector.mutate(arIn)

	var returnedAR interface{}
	returnedAR = arOut

	// If v1 received
	if arInVersion.Version == admv1.SchemeGroupVersion.Version {
		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Returned AdmissionReview '%v' Response=%+v", arOut.GroupVersionKind(), arOut.Response)
		}
	} else { // If v1beta1 received: convert v1 to v1beta1
		arOutv1beta1 := &admv1beta1.AdmissionReview{}
		arOutv1beta1.SetGroupVersionKind(admv1beta1.SchemeGroupVersion.WithKind("AdmissionReview"))
		Convert_admission_AdmissionReview_To_v1beta1_AdmissionReview(arOut, arOutv1beta1)
		returnedAR = arOutv1beta1

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Returned AdmissionReview '%v' Response=%+v", arOutv1beta1.GroupVersionKind(), arOutv1beta1.Response)
		}
	}

	response, err := json.Marshal(returnedAR)
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
