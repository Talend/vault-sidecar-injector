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

package proxy

import "talend/vault-sidecar-injector/pkg/config"

const (
	//--- Vault Sidecar Injector modes annotation keys (without prefix)
	vaultInjectorAnnotationProxyPortKey = "proxy-port" // Optional. Port assigned to local Vault proxy.
)

const (
	proxyContainerName    = config.VaultAgentContainerName // Name of our proxy container to inject
	vaultProxyDefaultPort = "8200"                         // Default port to access local Vault proxy
)

const (
	//--- Vault Agent placeholders related to modes
	vaultProxyPortPlaceholder = "<VSI_PROXY_PORT>"
)

const (
	//--- Vault Agent env vars related to modes
	vaultProxyConfigPlaceholderEnv = "VSI_PROXY_CONFIG_PLACEHOLDER"
)
