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
	"strings"

	"k8s.io/klog"
)

// Vault Sidecar Injector: Proxy Mode
func (vaultInjector *VaultInjector) proxyMode(annotations map[string]string) (string, error) {
	proxyPort := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationProxyPortKey]]

	if proxyPort == "" { // Default port
		proxyPort = vaultProxyDefaultPort
	}

	proxyConfig := strings.Replace(vaultInjector.ProxyConfig, vaultProxyPortPlaceholder, proxyPort, -1)

	// TODO - to remove
	klog.Infof("vaultInjector.ProxyConfig: %s", vaultInjector.ProxyConfig)
	klog.Infof("proxyConfig: %s", proxyConfig)
	// TODO - to remove

	return proxyConfig, nil
}
