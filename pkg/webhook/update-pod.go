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
	"errors"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func (vaultInjector *VaultInjector) updatePodSpec(pod *corev1.Pod) (patch []patchOperation, err error) {
	// Add Security Context (if provided)
	if vaultInjector.SidecarConfig.SecurityContext != nil {
		var op string

		if pod.Spec.SecurityContext == nil {
			op = jsonPatchOpAdd
		} else {
			op = jsonPatchOpReplace
		}

		patch = append(patch, patchOperation{Op: op, Path: jsonPathSecurityCtx, Value: *vaultInjector.SidecarConfig.SecurityContext})
	}

	// Extract labels and annotations to compute values for placeholders in sidecars' configuration
	context, err := vaultInjector.computeContext(pod.Spec.Containers, pod.Labels, pod.Annotations)
	if err == nil {
		// Add lifecycle hooks to requesting pod's container(s) if needed
		patchHooks, err := vaultInjector.addLifecycleHooks(pod.Spec.Containers, pod.Annotations, *context)
		if err == nil {
			patch = append(patch, patchHooks...)

			// Add sidecars' initcontainers
			patchInitContainers, err := vaultInjector.addContainer(pod.Spec.InitContainers, pod.Annotations, jsonPathInitContainers, *context)
			if err == nil {
				patch = append(patch, patchInitContainers...)

				// Add sidecars' containers
				patchContainers, err := vaultInjector.addContainer(pod.Spec.Containers, pod.Annotations, jsonPathContainers, *context)
				if err == nil {
					patch = append(patch, patchContainers...)

					// Add volume(s)
					patch = append(patch, vaultInjector.addVolume(pod.Spec.Volumes, jsonPathVolumes)...)
					return patch, nil
				}
			}
		}
	}

	return nil, err
}

func (vaultInjector *VaultInjector) computeContext(podContainers []corev1.Container, labels, annotations map[string]string) (*sidecarContext, error) {
	var k8sSaSecretsVolName, vaultInjectorSaSecretsVolName, proxyConfig, secretsTemplates string

	// Get enabled Vault Sidecar Injector modes
	modes := make(map[string]bool, len(vaultInjectorModes))
	getModes(strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationModeKey]], ","), modes)

	vaultAuthMethod := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationAuthMethodKey]]
	vaultRole := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationRoleKey]]
	vaultSATokenPath := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSATokenKey]]

	if vaultAuthMethod == "" { // Default Vault Auth Method is "kubernetes"
		vaultAuthMethod = vaultK8sAuthMethod
	}

	if vaultRole == "" && vaultAuthMethod == vaultK8sAuthMethod { // If role annotation not provided and "kubernetes" Vault Auth
		// Look after application label to set role
		vaultRole = labels[vaultInjector.ApplicationLabelKey]

		if vaultRole == "" {
			klog.Errorf("Submitted pod must contain label %s", vaultInjector.ApplicationLabelKey)
			return nil, fmt.Errorf("Submitted pod must contain label %s", vaultInjector.ApplicationLabelKey)
		}
	}

	// Look after volumeMounts' Names for Service Account's mountPath: '/var/run/secrets/kubernetes.io/serviceaccount' and
	// possible custom value provided with 'sa-token' annotation (get rid of ending '/token' if any to have mount path only).
	//
	// To be done since Service Account Admission Controller does not automatically add volumeSource for our injected containers.
	k8sSaSecretsVolName, err := getServiceAccountTokenVolumeName(podContainers, k8sDefaultSATokenVolMountPath)
	if err != nil {
		return nil, err
	}

	if vaultSATokenPath == "" { // Use default SA volume
		vaultInjectorSaSecretsVolName = k8sSaSecretsVolName
	} else {
		vaultInjectorSaSecretsVolName, err = getServiceAccountTokenVolumeName(podContainers, strings.TrimSuffix(vaultSATokenPath, "/token"))
		if err != nil {
			return nil, err
		}
	}

	// Proxy mode
	if modes[vaultInjectorModeProxy] {
		proxyConfig, err = vaultInjector.proxyMode(annotations)
		if err != nil {
			return nil, err
		}
	}

	// Secrets mode
	if modes[vaultInjectorModeSecrets] {
		secretsTemplates, err = vaultInjector.secretsMode(labels, annotations)
		if err != nil {
			return nil, err
		}
	}

	return &sidecarContext{modes, k8sSaSecretsVolName, vaultInjectorSaSecretsVolName, vaultAuthMethod, vaultRole, proxyConfig, secretsTemplates}, nil
}

func (vaultInjector *VaultInjector) addLifecycleHooks(podContainers []corev1.Container, annotations map[string]string, context sidecarContext) (patch []patchOperation, err error) {
	if context.modes[vaultInjectorModeSecrets] {
		switch strings.ToLower(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationLifecycleHookKey]]) {
		default:
			return patch, nil
		case "y", "yes", "true", "on":
			if vaultInjector.PodslifecycleHooks.PostStart != nil {
				// As we inject hooks, there should have existing containers so len(podContainers) shoud be > 0
				if len(podContainers) == 0 {
					klog.Error("Submitted pod must contain at least one container")
					return nil, errors.New("Submitted pod must contain at least one container")
				}

				secretsVolMountPath, err := getMountPathOfSecretsVolume(podContainers)
				if err != nil {
					return nil, err
				}

				if vaultInjector.PodslifecycleHooks.PostStart.Exec == nil {
					klog.Error("Unsupported lifecycle hook. Only support Exec type")
					return nil, errors.New("Unsupported lifecycle hook. Only support Exec type")
				}

				// We will modify some values here so make a copy to not change origin
				hookCommand := make([]string, len(vaultInjector.PodslifecycleHooks.PostStart.Exec.Command))
				copy(hookCommand, vaultInjector.PodslifecycleHooks.PostStart.Exec.Command)

				for commandIdx := range hookCommand {
					hookCommand[commandIdx] = strings.Replace(hookCommand[commandIdx], appSvcSecretsVolMountPathPlaceholder, secretsVolMountPath, -1)
				}

				postStartHook := &corev1.Handler{Exec: &corev1.ExecAction{Command: hookCommand}}

				// Add hooks to container(s) of requesting pod
				for podCntIdx, podCnt := range podContainers {
					if podCnt.Lifecycle != nil {
						if podCnt.Lifecycle.PostStart != nil {
							klog.Warningf("Replacing existing postStart hook for container %s", podCnt.Name)
						}

						podCnt.Lifecycle.PostStart = postStartHook
					} else {
						podCnt.Lifecycle = &corev1.Lifecycle{
							PostStart: postStartHook,
						}
					}

					// Here we have to use 'replace' JSON Patch operation
					patch = append(patch, patchOperation{
						Op:    jsonPatchOpReplace,
						Path:  jsonPathContainers + "/" + strconv.Itoa(podCntIdx),
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

// Deal with both InitContainers & Containers
func (vaultInjector *VaultInjector) addContainer(podContainers []corev1.Container, annotations map[string]string, basePath string, context sidecarContext) (patch []patchOperation, err error) {
	var first bool
	var value interface{}
	var sidecarCfgContainers []corev1.Container

	initContainer := (basePath == jsonPathInitContainers)
	jobWorkload := (annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationWorkloadKey]] == vaultInjectorAnnotationWorkloadJobValue)

	// As we inject containers to a pod, there should have existing containers so len(podContainers) shoud be > 0
	if !initContainer && len(podContainers) == 0 {
		klog.Error("Submitted pod must contain at least one container")
		return nil, errors.New("Submitted pod must contain at least one container")
	}

	if initContainer {
		// there may be no initContainers in the requesting pod
		first = len(podContainers) == 0
		sidecarCfgContainers = vaultInjector.SidecarConfig.InitContainers
	} else {
		first = false
		sidecarCfgContainers = vaultInjector.SidecarConfig.Containers

		// If workload is a job we expect only one container (good assumption given current limitations using job with sidecars)
		// Limitation to remove when KEP https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/sidecarcontainers.md is implemented and supported on our clusters
		if jobWorkload && len(podContainers) > 1 {
			klog.Error("Submitted pod contains more than one container: not supported for job workload")
			return nil, errors.New("Submitted pod contains more than one container: not supported for job workload")
		}
	}

	// Add our injected containers/initContainers to the submitted pod
	sidecarCntIdx := 0
	for _, sidecarCnt := range sidecarCfgContainers {
		if !jobWorkload && sidecarCnt.Name == jobMonitoringContainerName {
			// Workload is not a job so do not inject our job specific sidecar
			continue
		}
		container := sidecarCnt

		// We will modify some values here so make a copy to not change origin
		container.Command = make([]string, len(sidecarCnt.Command))
		copy(container.Command, sidecarCnt.Command)

		for commandIdx := range container.Command {
			if !initContainer && jobWorkload {
				container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], appJobContainerNamePlaceholder, podContainers[0].Name, -1)
			}
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], appJobVarPlaceholder, strconv.FormatBool(jobWorkload), -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], vaultAuthMethodPlaceholder, context.vaultAuthMethod, -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], vaultRolePlaceholder, context.vaultRole, -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], vaultProxyConfigPlaceholder, context.proxy, -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], templateTemplatesPlaceholder, context.templates, -1)
		}

		// We will modify some values here so make a copy to not change origin
		container.VolumeMounts = make([]corev1.VolumeMount, len(sidecarCnt.VolumeMounts))
		copy(container.VolumeMounts, sidecarCnt.VolumeMounts)

		// Loop to set proper volume names (extracted values from submitted pod)
		for volMountIdx := range container.VolumeMounts {
			switch container.VolumeMounts[volMountIdx].MountPath {
			case k8sDefaultSATokenVolMountPath:
				container.VolumeMounts[volMountIdx].Name = context.k8sDefaultSATokenVolumeName
			case vaultInjectorSATokenVolMountPath:
				container.VolumeMounts[volMountIdx].Name = context.vaultInjectorSATokenVolumeName
			}
		}

		value = container
		path := basePath

		// For initContainers:
		// add them at the beginning of the array to make sure they are run before any initContainers in the requesting pod: this way initContainers
		// belonging to the pod have a chance to process the secrets file(s) if needed.
		//
		// For containers:
		// let's add them also at the beginning of the array (even if no order constraint there as they are started in parallel by K8S)
		if first {
			first = false
			value = []corev1.Container{container}
		} else {
			// JSON Patch: use '/<index of sidecar>' to add our container/initContainer at the beginning of the array
			path = path + "/" + strconv.Itoa(sidecarCntIdx)
		}

		patch = append(patch, patchOperation{
			Op:    jsonPatchOpAdd,
			Path:  path,
			Value: value,
		})

		sidecarCntIdx++
	}

	return patch, nil
}

func (vaultInjector *VaultInjector) addVolume(podVolumes []corev1.Volume, basePath string) (patch []patchOperation) {
	first := len(podVolumes) == 0
	isSecretsVolumeInPod := false

	var value interface{}

	for _, sidecarVol := range vaultInjector.SidecarConfig.Volumes {
		// Do not inject the 'secrets' volume we define in our injector config if the pod we mutate already has a definition for such volume
		if sidecarVol.Name == appSvcSecretsVolName && len(podVolumes) > 0 {
			for _, podVol := range podVolumes {
				if podVol.Name == appSvcSecretsVolName {
					isSecretsVolumeInPod = true
					break
				}
			}

			if isSecretsVolumeInPod { // Volume 'secrets' exists in pod so do not add ours
				klog.Infof("Found existing '%s' volume in requesting pod: skip injector volume definition", appSvcSecretsVolName)
				continue
			}
		}

		value = sidecarVol
		path := basePath

		if first {
			first = false
			value = []corev1.Volume{sidecarVol}
		} else {
			// JSON Patch: use '-' to add our volumes at the end of the array (to not overwrite existing ones)
			path = path + "/-"
		}

		patch = append(patch, patchOperation{
			Op:    jsonPatchOpAdd,
			Path:  path,
			Value: value,
		})
	}

	return patch
}
