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

package secrets

const (
	//--- Vault Sidecar Injector modes annotation keys (without prefix)
	vaultInjectorAnnotationSecretsPathKey            = "secrets-path"             // Optional. Full path, e.g.: "secret/<some value>", "aws/creds/<some role>", ... Several values separated by ','.
	vaultInjectorAnnotationSecretsTemplateKey        = "secrets-template"         // Optional. Allow to override default template. Ignore 'secrets-path' annotation. Several values separated by ','.
	vaultInjectorAnnotationTemplateDestKey           = "secrets-destination"      // Optional. If not set, secrets will be stored in file "secrets.properties". Several values separated by ','.
	vaultInjectorAnnotationLifecycleHookKey          = "secrets-hook"             // Optional. If set, lifecycle hooks loaded from config will be added to pod's container(s)
	vaultInjectorAnnotationSecretsTypeKey            = "secrets-type"             // Optional. Type of secrets to handle: dynamic (default) or static
	vaultInjectorAnnotationSecretsInjectionMethodKey = "secrets-injection-method" // Optional. Method used to provide secrets to applications: file (default) or env
	vaultInjectorAnnotationTemplateCmdKey            = "notify"                   // Optional. Command to run after template is rendered. Several values separated by ','.
)

const (
	secretsContainerName               = "tvsi-vault-agent"      // Name of our secrets container to inject
	secretsInitContainerName           = "tvsi-vault-agent-init" // Name of our secrets init container to inject
	secretsEnvInitContainerName        = "tvsi-env-init"         // Name of our env process init container to inject
	templateAppSvcDefaultDestination   = "secrets.properties"    // Default secrets destination
	vaultDefaultSecretsEnginePath      = "secret"                // Default path for Vault K/V Secrets Engine if no 'secrets-path' annotation
	secretsAnnotationSeparator         = ","                     // Generic separator for secrets annotations' values
	secretsAnnotationTemplateSeparator = "---"                   // Separator for secrets templates annotation's values
)

const (
	//--- Volume & VolumeMount for secrets
	SecretsVolName          = "secrets"             // Name of the volume shared between containers to store secrets file(s)
	SecretsDefaultMountPath = "/opt/talend/secrets" // Default mount path for secretsVolName volume
)

const (
	//--- Vault Agent placeholders related to modes
	secretsVaultPathPlaceholder       = "<VSI_SECRETS_VAULT_SECRETS_PATH>"
	secretsDestinationPlaceholder     = "<VSI_SECRETS_DESTINATION>"
	secretsTemplateContentPlaceholder = "<VSI_SECRETS_TEMPLATE_CONTENT>"
	secretsTemplateCommandPlaceholder = "<VSI_SECRETS_TEMPLATE_COMMAND_TO_RUN>"
	secretsVolMountPathPlaceholder    = "<VSI_SECRETS_VOL_MOUNTPATH>"
)

const (
	//--- Vault Agent env vars related to modes
	secretsTemplatesPlaceholderEnv = "VSI_SECRETS_TEMPLATES_PLACEHOLDER"
)

const (
	//--- Vault Sidecar Injector supported modes
	VaultInjectorModeSecrets = "secrets" // Enable fetching of secrets from Vault store
)

const (
	//--- Vault Sidecar Injector secrets type
	vaultInjectorSecretsTypeDynamic = "dynamic"
	vaultInjectorSecretsTypeStatic  = "static"
)

const (
	//--- Vault Sidecar Injector secrets injection method
	vaultInjectorSecretsInjectionMethodFile = "file"
	vaultInjectorSecretsInjectionMethodEnv  = "env"
)

const (
	vaultInjectorEnvProcess = "vaultinjector-env"
)
