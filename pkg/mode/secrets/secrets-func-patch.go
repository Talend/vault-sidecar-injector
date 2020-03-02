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

import (
	"fmt"
	"strconv"
	"strings"
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog"
)

func secretsModePatch(config *cfg.VSIConfig, podSpec corev1.PodSpec, annotations map[string]string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	// Add lifecycle hooks to requesting pod's container(s) if needed
	if patchHooks, err := addLifecycleHooks(config, podSpec.Containers, annotations, context); err == nil {
		patch = append(patch, patchHooks...)
		return patch, nil
	}

	return nil, err
}

func addLifecycleHooks(config *cfg.VSIConfig, podContainers []corev1.Container, annotations map[string]string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	// Only for dynamic secrets
	if !isSecretsStatic(context) {
		switch strings.ToLower(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationLifecycleHookKey]]) {
		default:
			return patch, nil
		case "y", "yes", "true", "on":
			if config.PodslifecycleHooks.PostStart != nil {
				// As we inject hooks, there should have existing containers so len(podContainers) shoud be > 0
				if len(podContainers) == 0 {
					err = fmt.Errorf("[%s] Submitted pod must contain at least one container", vaultInjectorModeSecrets)
					klog.Error(err.Error())
					return nil, err
				}

				secretsVolMountPath, err := getMountPathOfSecretsVolume(podContainers)
				if err != nil {
					return nil, err
				}

				if config.PodslifecycleHooks.PostStart.Exec == nil {
					err = fmt.Errorf("[%s] Unsupported lifecycle hook. Only support Exec type", vaultInjectorModeSecrets)
					klog.Error(err.Error())
					return nil, err
				}

				// We will modify some values here so make a copy to not change origin
				hookCommand := make([]string, len(config.PodslifecycleHooks.PostStart.Exec.Command))
				copy(hookCommand, config.PodslifecycleHooks.PostStart.Exec.Command)

				for commandIdx := range hookCommand {
					hookCommand[commandIdx] = strings.Replace(hookCommand[commandIdx], secretsVolMountPathPlaceholder, secretsVolMountPath, -1)
				}

				postStartHook := &corev1.Handler{Exec: &corev1.ExecAction{Command: hookCommand}}

				// Add hooks to container(s) of requesting pod
				for podCntIdx, podCnt := range podContainers {
					if podCnt.Lifecycle != nil {
						if podCnt.Lifecycle.PostStart != nil {
							klog.Warningf("[%s] Replacing existing postStart hook for container %s", vaultInjectorModeSecrets, podCnt.Name)
						}

						podCnt.Lifecycle.PostStart = postStartHook
					} else {
						podCnt.Lifecycle = &corev1.Lifecycle{
							PostStart: postStartHook,
						}
					}

					// Here we have to use 'replace' JSON Patch operation
					patch = append(patch, ctx.PatchOperation{
						Op:    ctx.JsonPatchOpReplace,
						Path:  ctx.JsonPathContainers + "/" + strconv.Itoa(podCntIdx),
						Value: podCnt,
					})
				}
			}

			return patch, nil
		}
	} else {
		return patch, nil
	}
}
