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

package webhook

const (
	//--- Vault Sidecar Injector annotation keys (without prefix)
	// Input annotations (set on incoming manifest)
	vaultInjectorAnnotationInjectKey     = "inject"   // Mandatory
	vaultInjectorAnnotationAuthMethodKey = "auth"     // Optional. Vault Auth Method to use: kubernetes (default) or approle
	vaultInjectorAnnotationModeKey       = "mode"     // Optional. Comma-separated list of mode(s) to enable.
	vaultInjectorAnnotationRoleKey       = "role"     // Optional. To explicitly provide Vault role to use
	vaultInjectorAnnotationSATokenKey    = "sa-token" // Optional. Full path to service account token used for Vault Kubernetes authentication
	vaultInjectorAnnotationWorkloadKey   = "workload" // Optional. If set to "job", supplementary container and signaling mechanism will also be injected to properly handle k8s job
	// Output annotation (set by VSI webhook)
	vaultInjectorAnnotationStatusKey = "status" // Not to be set by requesting pods: set by the Webhook Admission Controller if injection ok
)

const (
	//--- Vault Sidecar Injector status
	vaultInjectorStatusInjected = "injected"
)

const (
	//--- Vault Sidecar Injector workloads
	vaultInjectorWorkloadJob = "job"
)

const (
	//--- Vault Sidecar Injector mount path for service accounts
	vaultInjectorSATokenVolMountPath = "/var/run/secrets/talend/vault-sidecar-injector/serviceaccount"
	k8sDefaultSATokenVolMountPath    = "/var/run/secrets/kubernetes.io/serviceaccount"
	//-- Volumes
	secretsVolName = "secrets" // Name of the volume shared between containers to store secrets file(s)
)

const (
	vaultK8sAuthMethod = "kubernetes" // Default auth method used by Vault Agent
)

const (
	//--- Vault Agent env vars
	vaultRoleEnv       = "VSI_VAULT_ROLE"
	vaultAuthMethodEnv = "VSI_VAULT_AUTH_METHOD"
)
