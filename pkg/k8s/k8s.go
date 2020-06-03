// Copyright © 2019-2020 Talend - www.talend.com
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

package k8s

import (
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// New : init K8S client
func New() *K8SClient {
	// In-cluster K8S client: must exist obviously
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("Failed to load in-cluster K8S config: %s", err)
		panic(err)
	}

	k8sClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		klog.Errorf("Error creating K8S client: %s", err)
		panic(err)
	}

	return &K8SClient{k8sClientset}
}

// PatchWebhookConfiguration generates CA and certificate for webhook then patches MutatingWebhookConfiguration's caBundle
func (k8sctl *K8SClient) PatchWebhookConfiguration(webhookCfgName string, ca []byte) error {
	// Patch MutatingWebhookConfiguration resource with generated CA (should be base64-encoded)
	_, err := k8sctl.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Patch(
		webhookCfgName, types.JSONPatchType, []byte(fmt.Sprintf(
			`[{
				"op": "add",
				"path": "/webhooks/0/clientConfig/caBundle",
				"value": %q
			}]`, base64.StdEncoding.EncodeToString(ca))))
	if err != nil {
		klog.Errorf("Error patching MutatingWebhookConfiguration's caBundle: %s", err)
		return err
	}

	return nil
}
