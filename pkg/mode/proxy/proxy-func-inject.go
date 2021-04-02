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

import (
	ctx "talend/vault-sidecar-injector/pkg/context"
	m "talend/vault-sidecar-injector/pkg/mode"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func proxyModeInject(containerBasePath string, podContainers []corev1.Container, containerName string, env []corev1.EnvVar, context *ctx.InjectionContext) (bool, error) {
	for _, cntName := range proxyContainerNames[containerBasePath] {
		if cntName == containerName {
			klog.Infof("[%s] Injecting container %s (path: %s)", m.VaultInjectorModeProxy, containerName, containerBasePath)

			// Resolve proxy env vars
			for envIdx := range env {
				if env[envIdx].Name == vaultProxyConfigPlaceholderEnv {
					env[envIdx].Value = context.ModesConfig[m.VaultInjectorModeProxy].GetTemplate()
				}
			}

			return true, nil
		}
	}

	return false, nil
}
