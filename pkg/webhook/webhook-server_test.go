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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"
	"testing"

	"k8s.io/apimachinery/pkg/util/uuid"

	admv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

var clientDeserializer = scheme.Codecs.UniversalDeserializer()

type testResource struct {
	manifest        string
	name            string
	namespace       string
	podTemplateSpec *corev1.PodTemplateSpec
}

type assertFunc func(*admv1.AdmissionResponse)

func TestWebhookServerOK(t *testing.T) {
	err := mutateWorkloads("../../test/workloads/ok/*.yaml",
		func(resp *admv1.AdmissionResponse) {
			assert.Condition(t, func() bool {
				// Handle injection cases *and* also pod submitted without `inject: "true"` annotation
				if (resp.Allowed && resp.Patch != nil && resp.Result == nil) || (resp.Allowed && resp.Patch == nil && resp.Result == nil) {
					return true
				}

				return false
			}, "Inconsistent AdmissionResponse")

			if resp.Patch != nil {
				var patch []ctx.PatchOperation
				if err := yaml.Unmarshal(resp.Patch, &patch); err != nil {
					t.Errorf("JSON Patch unmarshal error \"%s\"", err)
				}

				klog.Infof("JSON Patch=%+v", patch)
			}
		})

	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestWebhookServerKO(t *testing.T) {
	err := mutateWorkloads("../../test/workloads/ko/*.yaml",
		func(resp *admv1.AdmissionResponse) {
			assert.Condition(t, func() bool {
				// Handle error cases
				if !resp.Allowed && resp.Patch == nil && resp.Result != nil {
					return true
				}

				return false
			}, "Inconsistent AdmissionResponse")

			klog.Infof("Result=%+v", resp.Result)
		})

	if err != nil {
		t.Fatalf("%s", err)
	}
}

func mutateWorkloads(manifestsPattern string, test assertFunc) error {
	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	if verbose {
		// Set Klog verbosity level to have detailed logs from our webhook (where we use level 5+ to log such info)
		klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
		klog.InitFlags(klogFlags)
		klogFlags.Set("v", "5")
	}

	// Create webhook instance
	vaultInjector, err := createVaultInjector()
	if err != nil {
		return fmt.Errorf("Loading error: %s", err)
	}

	// Get all test workloads
	workloads, err := filepath.Glob(manifestsPattern)
	if err != nil {
		return fmt.Errorf("Fail listing files: %s", err)
	}

	// Loop on all test workloads: mutate and display JSON Patch structure
	for _, workloadManifest := range workloads {
		klog.Info("================================================================================================")
		klog.Infof("Loading workload %s", workloadManifest)

		ar, err := (&testResource{manifest: workloadManifest}).load()
		if err != nil {
			return fmt.Errorf("Error creating AR: %s", err)
		}

		// Mutate pod and test result
		test(vaultInjector.mutate(ar))
	}

	return nil
}

func createVaultInjector() (*VaultInjector, error) {
	vsiCfg, err := cfg.Load(
		cfg.WhSvrParameters{
			Port: 0, MetricsPort: 0,
			CACertFile: "", CertFile: "", KeyFile: "",
			WebhookCfgName:      "",
			AnnotationKeyPrefix: "sidecar.vault.talend.org", AppLabelKey: "com.talend.application", AppServiceLabelKey: "com.talend.service",
			InjectionCfgFile:      "../../test/config/injectionconfig.yaml",
			ProxyCfgFile:          "../../test/config/proxyconfig.hcl",
			TemplateBlockFile:     "../../test/config/tmplblock.hcl",
			TemplateDefaultFile:   "../../test/config/tmpldefault.tmpl",
			PodLifecycleHooksFile: "../../test/config/podlifecyclehooks.yaml",
		},
	)
	if err != nil {
		return nil, err
	}

	// Create webhook instance
	return New(vsiCfg, nil), nil
}

func (tr *testResource) load() (*admv1.AdmissionReview, error) {
	data, err := ioutil.ReadFile(tr.manifest)
	if err != nil {
		return nil, err
	}

	obj, _, err := clientDeserializer.Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}

	switch resource := obj.(type) {
	// Beware: despite content being the same, golang does not support 'fallthrough' keyword in type switch (see https://stackoverflow.com/questions/11531264/why-isnt-fallthrough-allowed-in-a-type-switch)
	case *appsv1.Deployment: // here 'resource' type is now *appsv1.Deployment
		tr.name = resource.Name
		tr.namespace = resource.Namespace
		tr.podTemplateSpec = &resource.Spec.Template
	case *batchv1.Job: // here 'resource' type is now *batchv1.Job
		tr.name = resource.Name
		tr.namespace = resource.Namespace
		tr.podTemplateSpec = &resource.Spec.Template
	default:
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

	if len(tr.podTemplateSpec.Spec.Containers) > 0 {
		tr.podTemplateSpec.Spec.Containers[0].VolumeMounts = append(tr.podTemplateSpec.Spec.Containers[0].VolumeMounts, saTokenVolumeMount)
	}

	tr.podTemplateSpec.Spec.Volumes = append(tr.podTemplateSpec.Spec.Volumes, saTokenVolume)
}

func (tr *testResource) createAdmissionReview() (*admv1.AdmissionReview, error) {
	rawPod, err := json.Marshal(tr.podTemplateSpec)
	if err != nil {
		return nil, err
	}

	ar := admv1.AdmissionReview{}
	ar.SetGroupVersionKind(admv1.SchemeGroupVersion.WithKind("AdmissionReview"))
	ar.Request = &admv1.AdmissionRequest{
		UID: uuid.NewUUID(),
		Kind: metav1.GroupVersionKind{
			Version: "v1",
			Kind:    "Pod",
		},
		Name:      tr.name,
		Namespace: tr.namespace,
		Operation: admv1.Create,
		UserInfo: authenticationv1.UserInfo{
			Username: "vault-sidecar-injector",
		},
		Object: runtime.RawExtension{
			Raw: rawPod,
		},
	}

	return &ar, nil
}
