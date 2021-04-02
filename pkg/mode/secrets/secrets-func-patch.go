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

package secrets

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"
	m "talend/vault-sidecar-injector/pkg/mode"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog"
)

func secretsModePatch(config *cfg.VSIConfig, podSpec corev1.PodSpec, annotations map[string]string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	var patchHooks, patchCmd []ctx.PatchOperation

	if !IsSecretsStatic(context) { // Only for dynamic secrets
		// Add lifecycle hooks to requesting pod's container(s) if needed
		if patchHooks, err = addLifecycleHooks(config, podSpec.Containers, annotations, context); err == nil {
			patch = append(patch, patchHooks...)
		}
	} else {
		if IsSecretsInjectionEnv(context) { // Look if injection method is set to 'env'
			// Patch containers' commands to invoke 'vaultinjector-env' process first to add env vars from secrets
			if patchCmd, err = patchCommand(podSpec.Containers); err == nil {
				patch = append(patch, patchCmd...)
			}
		}
	}

	return
}

func patchCommand(podContainers []corev1.Container) (patch []ctx.PatchOperation, err error) {
	for podCntIdx, podCnt := range podContainers {
		secretsVolMountPath := getMountPathOfSecretsVolume(podCnt)

		if secretsVolMountPath == "" { // As we force volumeMount on 'secrets' volume if not defined on containers, pick default value
			secretsVolMountPath = SecretsDefaultMountPath
		}

		// We currently require an explicit command to determine what is the process to run in the end
		if podCnt.Command != nil {
			// Prepend existing command array with our specific env process
			command := append([]string{path.Join(secretsVolMountPath, vaultInjectorEnvProcess)}, podCnt.Command...)

			// Here we have to use 'replace' JSON Patch operation
			patch = append(patch, ctx.PatchOperation{
				Op:    ctx.JsonPatchOpReplace,
				Path:  ctx.JsonPathContainers + "/" + strconv.Itoa(podCntIdx) + "/command",
				Value: command,
			})
		} else {
			err = fmt.Errorf("No explicit command found for container %s", podCnt.Name)
			klog.Errorf("[%s] %s", m.VaultInjectorModeSecrets, err.Error())
			return
		}
	}

	return
}

func addLifecycleHooks(config *cfg.VSIConfig, podContainers []corev1.Container, annotations map[string]string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	switch strings.ToLower(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationLifecycleHookKey]]) {
	default:
		return
	case "y", "yes", "true", "on":
		// Check if job mode is enabled: this annotation should not be used then
		if context.ModesStatus[m.VaultInjectorModeJob] {
			err = fmt.Errorf("Submitted pod uses unsupported combination of '%s' annotation with '%s' mode", config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationLifecycleHookKey], m.VaultInjectorModeJob)
			klog.Errorf("[%s] %s", m.VaultInjectorModeSecrets, err.Error())
			return
		}

		if config.PodslifecycleHooks.PostStart != nil {
			if config.PodslifecycleHooks.PostStart.Exec == nil {
				err = errors.New("Unsupported lifecycle hook. Only support Exec type")
				klog.Errorf("[%s] %s", m.VaultInjectorModeSecrets, err.Error())
				return
			}

			// Add hooks to container(s) of requesting pod
			for podCntIdx, podCnt := range podContainers {
				secretsVolMountPath := getMountPathOfSecretsVolume(podCnt)

				if secretsVolMountPath == "" { // As we force volumeMount on 'secrets' volume if not defined on containers, pick default value
					secretsVolMountPath = SecretsDefaultMountPath
				}

				// We will modify some values here so make a copy to not change origin
				hookCommand := make([]string, len(config.PodslifecycleHooks.PostStart.Exec.Command))
				copy(hookCommand, config.PodslifecycleHooks.PostStart.Exec.Command)

				for commandIdx := range hookCommand {
					hookCommand[commandIdx] = strings.Replace(hookCommand[commandIdx], secretsVolMountPathPlaceholder, secretsVolMountPath, -1)
				}

				postStartHook := &corev1.Handler{Exec: &corev1.ExecAction{Command: hookCommand}}

				if podCnt.Lifecycle != nil {
					if podCnt.Lifecycle.PostStart != nil {
						klog.Warningf("[%s] Replacing existing postStart hook for container %s", m.VaultInjectorModeSecrets, podCnt.Name)
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

		return
	}
}
