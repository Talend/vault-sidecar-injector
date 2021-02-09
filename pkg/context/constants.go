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

package context

const (
	//--- Vault Sidecar Injector annotation keys (without prefix)
	// Input annotations (set on incoming manifest)
	VaultInjectorAnnotationInjectKey     = "inject"      // Mandatory
	VaultInjectorAnnotationVaultImageKey = "vault-image" // Optional. Image to inject
	VaultInjectorAnnotationAuthMethodKey = "auth"        // Optional. Vault Auth Method to use: kubernetes (default) or approle
	VaultInjectorAnnotationModeKey       = "mode"        // Optional. Comma-separated list of mode(s) to enable.
	VaultInjectorAnnotationRoleKey       = "role"        // Optional. To explicitly provide Vault role to use
	VaultInjectorAnnotationSATokenKey    = "sa-token"    // Optional. Full path to service account token used for Vault Kubernetes authentication
	VaultInjectorAnnotationWorkloadKey   = "workload"    // Optional and deprecated. If set to "job", supplementary container and signaling mechanism will also be injected to properly handle k8s job
	// Output annotation (set by VSI webhook)
	VaultInjectorAnnotationStatusKey = "status" // Not to be set by requesting pods: set by the Webhook Admission Controller if injection ok
)

const (
	//--- Vault Sidecar Injector status
	VaultInjectorStatusInjected = "injected"
)

const (
	VaultK8sAuthMethod     = "kubernetes" // Vault K8S auth method. Default for VSI.
	VaultAppRoleAuthMethod = "approle"    // Vault AppRole auth method
)

const (
	//--- JSON Patch operations
	JsonPatchOpAdd     = "add"
	JsonPatchOpReplace = "replace"
)

const (
	//--- JSON Path
	JsonPathAnnotations    = "/metadata/annotations"
	JsonPathInitContainers = "/spec/initContainers"
	JsonPathContainers     = "/spec/containers"
	JsonPathVolumes        = "/spec/volumes"
)
