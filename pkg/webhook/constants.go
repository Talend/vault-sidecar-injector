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
	vaultInjectorAnnotationRoleKey            = "role"                // Optional. To explicitly provide Vault role to use
	vaultInjectorAnnotationSecretsPathKey     = "secrets-path"        // Optional. Full path, e.g.: "secret/<some value>", "aws/creds/<some role>", ... Several values separated by ','.
	vaultInjectorAnnotationSecretsTemplateKey = "secrets-template"    // Optional. Allow to override default Consul Template's template. Ignore 'secrets-path' annotation. Several values separated by ','.
	vaultInjectorAnnotationTemplateDestKey    = "secrets-destination" // Optional. If not set, secrets will be stored in file "secrets.properties". Several values separated by ','.
	vaultInjectorAnnotationLifecycleHookKey   = "secrets-hook"        // Optional. If set, lifecycle hooks loaded from config will be added to pod's container(s)
	vaultInjectorAnnotationTemplateCmdKey     = "notify"              // Optional. Command to run after template is rendered by Consul Template. Several values separated by ','.
	vaultInjectorAnnotationWorkloadKey        = "workload"            // Optional. If set to "job", supplementary container and signaling mechanism will also be injected to properly handle k8s job
	vaultInjectorAnnotationStatusKey          = "status"              // Not to be set by requesting pods: set by the Webhook Admission Controller if injection ok

	//--- Vault Sidecar Injector annotation values
	vaultInjectorAnnotationStatusValue      = "injected"
	vaultInjectorAnnotationWorkloadJobValue = "job"

	//--- Vault Agent & Consul Template placeholders
	vaultRolePlaceholder                       = "<APP_VAULT_ROLE>"
	vaultAppSvcSecretsPathPlaceholder          = "<APPSVC_VAULT_SECRETS_PATH>"
	consulTemplateAppSvcDestinationPlaceholder = "<APPSVC_SECRETS_DESTINATION>"
	consulTemplateTemplateContentPlaceholder   = "<APPSVC_TEMPLATE_CONTENT>"
	consulTemplateCommandPlaceholder           = "<APPSVC_TEMPLATE_COMMAND_TO_RUN>"
	consulTemplateTemplatesPlaceholder         = "<APPSVC_TEMPLATES>"
	appSvcSecretsVolMountPathPlaceholder       = "<APPSVC_SECRETS_VOL_MOUNTPATH>"

	appSvcSecretsVolName                   = "secrets"            // Name of the volume shared between containers to store secrets file(s)
	consulTemplateAppSvcDefaultDestination = "secrets.properties" // Default secrets destination
	k8sServiceAccountTokenVolMountPath     = "/var/run/secrets/kubernetes.io/serviceaccount"
	vaultDefaultSecretsEnginePath          = "secret" // Default path for Vault K/V Secrets Engine if no 'secrets-path' annotation

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
