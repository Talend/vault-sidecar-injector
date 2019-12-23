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
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"talend/vault-sidecar-injector/pkg/config"
	"testing"

	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type testResource struct {
	manifest        string
	workloadType    string
	podTemplateSpec *corev1.PodTemplateSpec
}

func TestInlineInjector(t *testing.T) {
	injectionCfg, err := config.Load(
		config.WhSvrParameters{
			0, 0, "", "",
			"sidecar.vault.talend.org", "com.talend.application", "com.talend.service",
			"../../test/config/sidecarconfig.yaml",
			"../../test/config/proxyconfig.hcl",
			"../../test/config/tmplblock.hcl",
			"../../test/config/tmpldefault.tmpl",
			"../../test/config/podlifecyclehooks.yaml",
		},
	)
	if err != nil {
		t.Errorf("Loading error \"%s\"", err)
	}

	// Create webhook instance
	vaultInjector := New(injectionCfg, nil)

	// Get all test workloads
	workloads, err := filepath.Glob("../../test/workloads/*.yaml")
	if err != nil {
		t.Errorf("Fail listing files: %s", err)
	}

	// Loop on all test workloads: mutate and display JSON Patch structure
	for _, workloadManifest := range workloads {
		klog.Infof("Loading workload %s", workloadManifest)

		ar, err := (&testResource{
			manifest: workloadManifest,
			workloadType: func(w string) string {
				if strings.HasSuffix(w, "-deployment.yaml") {
					return "deployment"
				} else if strings.HasSuffix(w, "-job.yaml") {
					return "job"
				} else {
					return ""
				}
			}(workloadManifest),
		}).load()
		if err != nil {
			t.Errorf("Error creating AR \"%s\"", err)
		}

		// Mutate pod
		resp := vaultInjector.mutate(ar)

		assert.Equal(t, true, resp.Allowed)
		assert.Nil(t, resp.Result)
		assert.NotNil(t, resp.Patch)

		var patch []patchOperation
		if err = yaml.Unmarshal(resp.Patch, &patch); err != nil {
			t.Errorf("JSON Patch unmarshal error \"%s\"", err)
		}

		klog.Infof("JSON Patch=%+v", patch)
	}
}

func (tr *testResource) load() (*v1beta1.AdmissionReview, error) {
	// TODO: Beware, there may be several resources. Only keep and mutate the ones with Vault Sidecar Injector's `sidecar.vault.talend.org/inject: "true"` annotation
	data, err := ioutil.ReadFile(tr.manifest)
	if err != nil {
		return nil, err
	}

	if tr.workloadType == "deployment" {
		resource := appsv1.Deployment{}
		_, _, err = deserializer.Decode(data, nil, &resource)
		if err != nil {
			return nil, err
		}

		tr.podTemplateSpec = &resource.Spec.Template
	} else if tr.workloadType == "job" {
		resource := batchv1.Job{}
		_, _, err = deserializer.Decode(data, nil, &resource)
		if err != nil {
			return nil, err
		}

		tr.podTemplateSpec = &resource.Spec.Template
	} else {
		return nil, errors.New("Worload not supported")
	}

	tr.addSATokenVolume()
	return tr.createAdmissionReview()
}

func (tr *testResource) addSATokenVolume() {
	// We expect to find serviceaccount token volume. It is dynamically added to the pod by the Service Account Admission Controller.
	// Add it manually here to pass internal check.
	saTokenVolumeMount := corev1.VolumeMount{
		Name:      "default-token-1234",
		ReadOnly:  true,
		MountPath: k8sDefaultSATokenVolMountPath,
	}

	saTokenVolume := corev1.Volume{
		Name: "default-token-1234",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "default-token",
			},
		},
	}

	tr.podTemplateSpec.Spec.Containers[0].VolumeMounts = append(tr.podTemplateSpec.Spec.Containers[0].VolumeMounts, saTokenVolumeMount)
	tr.podTemplateSpec.Spec.Volumes = append(tr.podTemplateSpec.Spec.Volumes, saTokenVolume)
}

func (tr *testResource) createAdmissionReview() (*v1beta1.AdmissionReview, error) {
	rawPod, err := json.Marshal(tr.podTemplateSpec)
	if err != nil {
		return nil, err
	}

	return &v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{
				Version: "v1",
				Kind:    "Pod",
			},
			Namespace: tr.podTemplateSpec.GetNamespace(),
			Operation: v1beta1.Create,
			UserInfo: authenticationv1.UserInfo{
				Username: "vault-sidecar-injector",
			},
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}, nil
}
