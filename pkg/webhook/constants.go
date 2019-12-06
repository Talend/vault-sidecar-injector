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

const (
	//--- Vault Sidecar Injector annotation keys (without prefix)
	vaultInjectorAnnotationInjectKey          = "inject"              // Mandatory
	vaultInjectorAnnotationAuthMethodKey      = "auth"                // Optional. Vault Auth Method to use: kubernetes (default) or approle
	vaultInjectorAnnotationModeKey            = "mode"                // Optional. Comma-separated list of mode(s) to enable.
	vaultInjectorAnnotationProxyPortKey       = "proxy-port"          // Optional. Port assigned to local Vault proxy.
	vaultInjectorAnnotationRoleKey            = "role"                // Optional. To explicitly provide Vault role to use
	vaultInjectorAnnotationSATokenKey         = "sa-token"            // Optional. Full path to service account token used for Vault Kubernetes authentication
	vaultInjectorAnnotationSecretsPathKey     = "secrets-path"        // Optional. Full path, e.g.: "secret/<some value>", "aws/creds/<some role>", ... Several values separated by ','.
	vaultInjectorAnnotationSecretsTemplateKey = "secrets-template"    // Optional. Allow to override default template. Ignore 'secrets-path' annotation. Several values separated by ','.
	vaultInjectorAnnotationTemplateDestKey    = "secrets-destination" // Optional. If not set, secrets will be stored in file "secrets.properties". Several values separated by ','.
	vaultInjectorAnnotationLifecycleHookKey   = "secrets-hook"        // Optional. If set, lifecycle hooks loaded from config will be added to pod's container(s)
	vaultInjectorAnnotationTemplateCmdKey     = "notify"              // Optional. Command to run after template is rendered. Several values separated by ','.
	vaultInjectorAnnotationWorkloadKey        = "workload"            // Optional. If set to "job", supplementary container and signaling mechanism will also be injected to properly handle k8s job
	vaultInjectorAnnotationStatusKey          = "status"              // Not to be set by requesting pods: set by the Webhook Admission Controller if injection ok

	//--- Vault Sidecar Injector annotation values
	vaultInjectorAnnotationStatusValue      = "injected"
	vaultInjectorAnnotationWorkloadJobValue = "job"

	//--- Vault Sidecar Injector supported modes
	vaultInjectorModeSecrets = "secrets" // Enable dynamic fetching of secrets from Vault store
	vaultInjectorModeProxy   = "proxy"   // Enable local Vault proxy

	//--- Vault Sidecar Injector mount path for service accounts
	vaultInjectorSATokenVolMountPath = "/var/run/secrets/talend/vault-sidecar-injector/serviceaccount"
	k8sDefaultSATokenVolMountPath    = "/var/run/secrets/kubernetes.io/serviceaccount"

	//--- Vault Agent placeholders
	vaultRolePlaceholder                 = "<APP_VAULT_ROLE>"
	vaultAuthMethodPlaceholder           = "<APPSVC_VAULT_AUTH_METHOD>"
	vaultProxyConfigPlaceholder          = "<APPSVC_PROXY_CONFIG>"
	vaultProxyPortPlaceholder            = "<APPSVC_PROXY_PORT>"
	vaultAppSvcSecretsPathPlaceholder    = "<APPSVC_VAULT_SECRETS_PATH>"
	templateAppSvcDestinationPlaceholder = "<APPSVC_SECRETS_DESTINATION>"
	templateContentPlaceholder           = "<APPSVC_TEMPLATE_CONTENT>"
	templateCommandPlaceholder           = "<APPSVC_TEMPLATE_COMMAND_TO_RUN>"
	templateTemplatesPlaceholder         = "<APPSVC_TEMPLATES>"
	appSvcSecretsVolMountPathPlaceholder = "<APPSVC_SECRETS_VOL_MOUNTPATH>"

	vaultK8sAuthMethod                = "kubernetes"         // Default auth method used by Vault Agent
	appSvcSecretsVolName              = "secrets"            // Name of the volume shared between containers to store secrets file(s)
	templateAppSvcDefaultDestination  = "secrets.properties" // Default secrets destination
	k8sDefaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	vaultDefaultSecretsEnginePath     = "secret" // Default path for Vault K/V Secrets Engine if no 'secrets-path' annotation
	vaultProxyDefaultPort             = "8200"   // Default port to access local Vault proxy

	//--- Job handling - Temporary mechanism until KEP https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/sidecarcontainers.md is implemented (and we migrate on appropriate version of k8s)
	jobMonitoringContainerName     = "tvsi-job-babysitter" // Name of our specific sidecar container to inject in submitted jobs
	appJobContainerNamePlaceholder = "<APP_JOB_CNT_NAME>"  // Name of the app job's container
	appJobVarPlaceholder           = "<APP_JOB>"           // Special script var set to "true" if submitted workload is a k8s job

	//--- JSON Patch operations
	jsonPatchOpAdd     = "add"
	jsonPatchOpReplace = "replace"

	//--- JSON Path
	jsonPathAnnotations    = "/metadata/annotations"
	jsonPathSecurityCtx    = "/spec/securityContext"
	jsonPathInitContainers = "/spec/initContainers"
	jsonPathContainers     = "/spec/containers"
	jsonPathVolumes        = "/spec/volumes"
)
