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

package webhook

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	ctx "talend/vault-sidecar-injector/pkg/context"
	m "talend/vault-sidecar-injector/pkg/mode"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func (vaultInjector *VaultInjector) updatePodSpec(pod *corev1.Pod) (patch []ctx.PatchOperation, err error) {
	var context *ctx.InjectionContext
	var patchPod, patchInitContainers, patchContainers []ctx.PatchOperation

	// We expect at least one container in submitted pod
	if len(pod.Spec.Containers) == 0 {
		err = errors.New("Submitted pod must contain at least one container")
		klog.Error(err.Error())
		return
	}

	// 1) Extract labels and annotations to compute values for placeholders in injection configuration
	if context, err = vaultInjector.computeContext(pod.Spec.Containers, pod.Labels, pod.Annotations); err == nil {
		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("context=%+v", context)
		}

		// 2) Patch submitted pod
		if patchPod, err = vaultInjector.patchPod(pod.Spec, pod.Annotations, context); err == nil {
			patch = append(patch, patchPod...)

			// 3) Add init container(s) to submitted pod
			if patchInitContainers, err = vaultInjector.addContainer(pod.Spec.InitContainers, ctx.JsonPathInitContainers, context); err == nil {
				patch = append(patch, patchInitContainers...)

				// 4) Add sidecar(s) to submitted pod
				if patchContainers, err = vaultInjector.addContainer(pod.Spec.Containers, ctx.JsonPathContainers, context); err == nil {
					patch = append(patch, patchContainers...)

					// 5) Add volume(s)
					patch = append(patch, vaultInjector.addVolume(pod.Spec.Volumes, ctx.JsonPathVolumes)...)
				}
			}
		}
	}

	return
}

func (vaultInjector *VaultInjector) computeContext(podContainers []corev1.Container, labels, annotations map[string]string) (*ctx.InjectionContext, error) {
	var k8sSaSecretsVolName, vaultInjectorSaSecretsVolName string

	// Get status for Vault Sidecar Injector modes
	modesStatus := make(map[string]bool, len(m.VaultInjectorModes))
	m.GetModesStatus(strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationModeKey]], ","), modesStatus)

	// !!! This annotation is deprecated !!! Enable job mode if used
	if annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationWorkloadKey]] == vaultInjectorWorkloadJob {
		modesStatus[vaultInjectorWorkloadJob] = true
	}

	klog.Infof("Modes status: %+v", modesStatus)

	vaultAuthMethod := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationAuthMethodKey]]
	vaultRole := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationRoleKey]]
	vaultSATokenPath := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSATokenKey]]

	if vaultAuthMethod == "" { // Default Vault Auth Method is "kubernetes"
		vaultAuthMethod = vaultK8sAuthMethod
	}

	if (vaultRole == "") && (vaultAuthMethod == vaultK8sAuthMethod) { // If role annotation not provided and "kubernetes" Vault Auth
		// Look after application label to set role
		vaultRole = labels[vaultInjector.ApplicationLabelKey]

		if vaultRole == "" {
			err := fmt.Errorf("Submitted pod must contain label %s", vaultInjector.ApplicationLabelKey)
			klog.Error(err.Error())
			return nil, err
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

	// Loop through enabled modes and call associated compute functions to compute configs
	modesConfig := make(map[string]ctx.ModeConfig, len(m.VaultInjectorModes))

	for mode, enabled := range modesStatus {
		if enabled && m.VaultInjectorModes[mode].ComputeTemplatesFunc != nil {
			modesConfig[mode], err = m.VaultInjectorModes[mode].ComputeTemplatesFunc(vaultInjector.VSIConfig, labels, annotations)
			if err != nil {
				return nil, err
			}
		}
	}

	return &ctx.InjectionContext{
		K8sDefaultSATokenVolumeName:    k8sSaSecretsVolName,
		VaultInjectorSATokenVolumeName: vaultInjectorSaSecretsVolName,
		VaultAuthMethod:                vaultAuthMethod,
		VaultRole:                      vaultRole,
		ModesStatus:                    modesStatus,
		ModesConfig:                    modesConfig}, nil
}

func (vaultInjector *VaultInjector) patchPod(podSpec corev1.PodSpec, annotations map[string]string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	for mode, enabled := range context.ModesStatus {
		if enabled && m.VaultInjectorModes[mode].PatchPodFunc != nil {
			patchPod, err := m.VaultInjectorModes[mode].PatchPodFunc(vaultInjector.VSIConfig, podSpec, annotations, context)
			if err != nil {
				return nil, err
			}

			patch = append(patch, patchPod...)
		}
	}

	return patch, nil
}

// Deal with both InitContainers & Containers
func (vaultInjector *VaultInjector) addContainer(podContainers []corev1.Container, basePath string, context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) {
	var value interface{}

	first := false
	injectionCfgContainers := vaultInjector.InjectionConfig.Containers
	initContainer := (basePath == ctx.JsonPathInitContainers)

	if initContainer {
		// there may be no init container in the requesting pod
		first = (len(podContainers) == 0)
		injectionCfgContainers = vaultInjector.InjectionConfig.InitContainers
	}

	// Add our injected containers/initContainers to the submitted pod
	injectionCntIdx := 0
	for _, injectionCnt := range injectionCfgContainers {
		container := injectionCnt

		// We will modify env vars so make a copy to not change origin
		container.Env = make([]corev1.EnvVar, len(injectionCnt.Env))
		copy(container.Env, injectionCnt.Env)

		// Iterate over enabled mode(s) to check if we inject this container and resolve env vars if needed
		inject := false
		for mode, enabled := range context.ModesStatus {
			if enabled {
				modeInject, err := m.VaultInjectorModes[mode].InjectContainerFunc(basePath, podContainers, container.Name, container.Env, context)
				if err != nil {
					return nil, err
				}

				if modeInject {
					inject = true
				}
			}
		}

		// If no enabled mode(s) want this container to be injected: skip it
		if !inject {
			continue
		}

		// Set Vault role and Auth Method env vars
		for envIdx := range container.Env {
			if container.Env[envIdx].Name == vaultRoleEnv {
				container.Env[envIdx].Value = context.VaultRole
			}

			if container.Env[envIdx].Name == vaultAuthMethodEnv {
				container.Env[envIdx].Value = context.VaultAuthMethod
			}
		}

		if klog.V(5) { // enabled by providing '-v=5' at least
			klog.Infof("Env vars: %+v", container.Env)
		}

		// We will modify some values here so make a copy to not change origin
		container.VolumeMounts = make([]corev1.VolumeMount, len(injectionCnt.VolumeMounts))
		copy(container.VolumeMounts, injectionCnt.VolumeMounts)

		// Loop to set proper volume names (extracted values from submitted pod)
		for volMountIdx := range container.VolumeMounts {
			switch container.VolumeMounts[volMountIdx].MountPath {
			case k8sDefaultSATokenVolMountPath:
				container.VolumeMounts[volMountIdx].Name = context.K8sDefaultSATokenVolumeName
			case vaultInjectorSATokenVolMountPath:
				container.VolumeMounts[volMountIdx].Name = context.VaultInjectorSATokenVolumeName
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
			// JSON Patch: use '/<index of container>' to add our container/initContainer at the beginning of the array
			path = path + "/" + strconv.Itoa(injectionCntIdx)
		}

		patch = append(patch, ctx.PatchOperation{
			Op:    ctx.JsonPatchOpAdd,
			Path:  path,
			Value: value,
		})

		injectionCntIdx++
	}

	return patch, nil
}

func (vaultInjector *VaultInjector) addVolume(podVolumes []corev1.Volume, basePath string) (patch []ctx.PatchOperation) {
	first := len(podVolumes) == 0

	var value interface{}

	for _, sidecarVol := range vaultInjector.InjectionConfig.Volumes {
		// Do not inject the 'secrets' volume we define in our injector config if the pod we mutate already has a definition for such volume
		isSecretsVolumeInPod := false
		if sidecarVol.Name == secretsVolName && len(podVolumes) > 0 {
			for _, podVol := range podVolumes {
				if podVol.Name == secretsVolName {
					isSecretsVolumeInPod = true
					break
				}
			}

			if isSecretsVolumeInPod { // Volume 'secrets' exists in pod so do not add ours
				klog.Infof("Found existing '%s' volume in requesting pod: skip injector volume definition", secretsVolName)
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

		patch = append(patch, ctx.PatchOperation{
			Op:    ctx.JsonPatchOpAdd,
			Path:  path,
			Value: value,
		})
	}

	return patch
}
