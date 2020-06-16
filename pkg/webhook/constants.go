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
	//--- Vault Sidecar Injector mount path for service accounts
	vaultInjectorSATokenVolMountPath = "/var/run/secrets/talend/vault-sidecar-injector/serviceaccount"
	k8sDefaultSATokenVolMountPath    = "/var/run/secrets/kubernetes.io/serviceaccount"

	//-- Volumes
	secretsVolName = "secrets" // Name of the volume shared between containers to store secrets file(s)

	//-- VolumeMounts
	secretsDefaultMountPath = "/opt/talend/secrets" // Default mount path for secretsVolName volume
)

const (
	//--- Vault Agent env vars
	vaultRoleEnv       = "VSI_VAULT_ROLE"
	vaultAuthMethodEnv = "VSI_VAULT_AUTH_METHOD"
)
