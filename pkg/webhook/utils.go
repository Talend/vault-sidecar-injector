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
	"fmt"
	"strings"

	ctx "talend/vault-sidecar-injector/pkg/context"

	admv1 "k8s.io/api/admission/v1"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	admregv1 "k8s.io/api/admissionregistration/v1beta1"
	admregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func init() {
	must(corev1.AddToScheme(runtimeScheme))

	// admission v1
	must(admv1.AddToScheme(runtimeScheme))
	must(admregv1.AddToScheme(runtimeScheme))

	// admission v1beta1
	must(admv1beta1.AddToScheme(runtimeScheme))
	must(admregv1beta1.AddToScheme(runtimeScheme))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Check whether the target resource need to be mutated
func mutationRequired(ignoredList []string, vaultInjectorAnnotations map[string]string, podMetadata *metav1.ObjectMeta) bool {
	var entityName string
	var entityNamespace string

	if podMetadata.Name == "" {
		entityName = podMetadata.GenerateName
	} else {
		entityName = podMetadata.Name
	}

	if podMetadata.Namespace == "" {
		entityNamespace = metav1.NamespaceDefault
	} else {
		entityNamespace = podMetadata.Namespace
	}

	// skip special Kubernetes system namespaces
	for _, namespace := range ignoredList {
		if entityNamespace == namespace {
			klog.Infof("Skip mutation for %v for it's in special namespace:%v", entityName, entityNamespace)
			return false
		}
	}

	annotations := podMetadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status := annotations[vaultInjectorAnnotations[ctx.VaultInjectorAnnotationStatusKey]]

	// determine whether to perform mutation based on annotation for the target resource
	var required bool
	if strings.ToLower(status) == ctx.VaultInjectorStatusInjected {
		required = false
	} else {
		switch strings.ToLower(annotations[vaultInjectorAnnotations[ctx.VaultInjectorAnnotationInjectKey]]) {
		default:
			required = false
		case "y", "yes", "true", "on":
			required = true
		}
	}

	klog.Infof("Mutation policy for %v/%v: status: %q required:%v", entityNamespace, entityName, status, required)
	return required
}

func getServiceAccountTokenVolumeName(cnts []corev1.Container, saTokenPath string) (string, error) {
	var k8sSaSecretsVolName string

Loop:
	for _, sourceContainer := range cnts {
		for _, volMount := range sourceContainer.VolumeMounts {
			if volMount.MountPath == saTokenPath {
				k8sSaSecretsVolName = volMount.Name
				break Loop
			}
		}
	}

	if k8sSaSecretsVolName == "" {
		err := fmt.Errorf("Volume Mount for path %s not found in submitted pod", saTokenPath)
		klog.Error(err.Error())
		return "", err
	}

	return k8sSaSecretsVolName, nil
}

func updateAnnotation(target map[string]string, added map[string]string) (patch []ctx.PatchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, ctx.PatchOperation{
				Op:   ctx.JsonPatchOpAdd,
				Path: ctx.JsonPathAnnotations,
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, ctx.PatchOperation{
				Op:    ctx.JsonPatchOpReplace,
				Path:  ctx.JsonPathAnnotations + "/" + key,
				Value: value,
			})
		}
	}

	return patch
}
