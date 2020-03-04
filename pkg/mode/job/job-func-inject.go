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

package job

import (
	"errors"
	ctx "talend/vault-sidecar-injector/pkg/context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func jobModeInject(containerBasePath string, podContainers []corev1.Container, containerName string, env []corev1.EnvVar, context *ctx.InjectionContext) (bool, error) {
	if (containerBasePath == ctx.JsonPathContainers) && (len(podContainers) != 1) {
		err := errors.New("Submitted pod should contain only one container")
		klog.Errorf("[%s] %s", vaultInjectorModeJob, err.Error())
		return false, err
	}

	for _, cntName := range jobContainerNames[containerBasePath] {
		if cntName == containerName {
			klog.Infof("[%s] Injecting container %s (path: %s)", vaultInjectorModeJob, containerName, containerBasePath)

			// Resolve job env vars
			for envIdx := range env {
				if env[envIdx].Name == jobContainerNameEnv {
					env[envIdx].Value = podContainers[0].Name
				}

				if env[envIdx].Name == jobWorkloadEnv {
					env[envIdx].Value = "true"
				}
			}

			return true, nil
		}
	}

	return false, nil
}
