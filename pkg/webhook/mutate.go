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
	ctx "talend/vault-sidecar-injector/pkg/context"

	admv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func (vaultInjector *VaultInjector) mutate(ar *admv1.AdmissionReview) *admv1.AdmissionResponse {
	var pod corev1.Pod
	var podName, podNamespace string

	req := ar.Request

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return &admv1.AdmissionResponse{
			UID: req.UID,
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
			UID:     req.UID,
			Allowed: true,
		}
	}

	annotations := map[string]string{vaultInjector.VaultInjectorAnnotationsFQ[ctx.VaultInjectorAnnotationStatusKey]: ctx.VaultInjectorStatusInjected}
	patchBytes, err := vaultInjector.createPatch(&pod, annotations)
	if err != nil {
		return &admv1.AdmissionResponse{
			UID:     req.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	klog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &admv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admv1.PatchType {
			pt := admv1.PatchTypeJSONPatch
			return &pt
		}(),
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
