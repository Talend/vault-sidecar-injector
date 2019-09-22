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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"talend/vault-sidecar-injector/pkg/config"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

// New : init new VaultInjector type
func New(cfg *config.InjectionConfig, server *http.Server) *VaultInjector {
	// Compute FQ annotations
	cfg.VaultInjectorAnnotationsFQ = make(map[string]string, len(vaultInjectorAnnotationKeys))
	for _, vaultAnnotationKey := range vaultInjectorAnnotationKeys {
		if cfg.VaultInjectorAnnotationKeyPrefix != "" {
			cfg.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = cfg.VaultInjectorAnnotationKeyPrefix + "/" + vaultAnnotationKey
		} else {
			cfg.VaultInjectorAnnotationsFQ[vaultAnnotationKey] = vaultAnnotationKey
		}
	}

	return &VaultInjector{
		InjectionConfig: cfg,
		Server:          server,
	}
}

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
	placeholders, err := vaultInjector.computeSidecarsPlaceholders(pod.Spec.Containers, pod.Labels, pod.Annotations)
	if err == nil {
		// Add lifecycle hooks to requesting pod's container(s) if needed
		patchHooks, err := vaultInjector.addLifecycleHooks(pod.Spec.Containers, pod.Annotations)
		if err == nil {
			patch = append(patch, patchHooks...)

			// Add sidecars' initcontainers
			patchInitContainers, err := vaultInjector.addContainer(pod.Spec.InitContainers, pod.Annotations, jsonPathInitContainers, *placeholders)
			if err == nil {
				patch = append(patch, patchInitContainers...)

				// Add sidecars' containers
				patchContainers, err := vaultInjector.addContainer(pod.Spec.Containers, pod.Annotations, jsonPathContainers, *placeholders)
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

func (vaultInjector *VaultInjector) addLifecycleHooks(podContainers []corev1.Container, annotations map[string]string) (patch []patchOperation, err error) {
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
}

func (vaultInjector *VaultInjector) computeSidecarsPlaceholders(podContainers []corev1.Container, labels, annotations map[string]string) (*sidecarPlaceholders, error) {
	// Look for mandatory labels in submitted pod
	applicationLabel := labels[vaultInjector.ApplicationLabelKey]
	applicationServiceLabel := labels[vaultInjector.ApplicationServiceLabelKey]

	if applicationLabel == "" || applicationServiceLabel == "" {
		klog.Errorf("Submitted pod must contain labels %s and %s", vaultInjector.ApplicationLabelKey, vaultInjector.ApplicationServiceLabelKey)
		return nil, fmt.Errorf("Submitted pod must contain labels %s and %s", vaultInjector.ApplicationLabelKey, vaultInjector.ApplicationServiceLabelKey)
	}

	// Look after volumeMounts' Name for mountPath '/var/run/secrets/kubernetes.io/serviceaccount'
	// To be done since Service Account Admission Controller does not automatically add volumeSource mounted at
	// '/var/run/secrets/kubernetes.io/serviceaccount' for our injected containers.
	// Vault Agent needs to retrieve service account's JWT token there to perform Vault K8S Auth.
	k8sSaSecretsVolName, err := getServiceAccountTokenVolume(podContainers)
	if err != nil {
		return nil, err
	}

	annotationVaultRole := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationRoleKey]]
	annotationVaultSATokenPath := annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSATokenKey]]
	annotationVaultSecretsPath := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsPathKey]], ",")
	annotationConsulTemplateTemplate := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsTemplateKey]], ",")
	annotationConsulTemplateDest := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateDestKey]], ",")
	annotationConsulTemplateCmd := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateCmdKey]], ",")

	annotationVaultSecretsPathNum := len(annotationVaultSecretsPath)
	annotationConsulTemplateTemplateNum := len(annotationConsulTemplateTemplate)
	annotationConsulTemplateDestNum := len(annotationConsulTemplateDest)

	if annotationVaultRole == "" { // If annotation not provided, Vault role set to application label
		annotationVaultRole = applicationLabel
	}

	if annotationVaultSATokenPath == "" { // Use default
		annotationVaultSATokenPath = k8sServiceAccountTokenPath
	}

	if annotationVaultSecretsPathNum == 1 && annotationVaultSecretsPath[0] == "" { // Build default secrets path: "secret/<application label>/<service label>"
		annotationVaultSecretsPath[0] = vaultDefaultSecretsEnginePath + "/" + applicationLabel + "/" + applicationServiceLabel
	}

	if annotationConsulTemplateDestNum == 1 && annotationConsulTemplateDest[0] == "" { // Use default
		annotationConsulTemplateDest[0] = consulTemplateAppSvcDefaultDestination
	}

	if annotationConsulTemplateTemplateNum == 1 && annotationConsulTemplateTemplate[0] == "" {
		// We must have same numbers of secrets path & secrets destinations
		if annotationConsulTemplateDestNum != annotationVaultSecretsPathNum {
			klog.Error("Submitted pod must contain same numbers of secrets path and secrets destinations")
			return nil, errors.New("Submitted pod must contain same numbers of secrets path and secrets destinations")
		}

		// If no custom template(s), use default Consul Template's template
		annotationConsulTemplateTemplate = make([]string, annotationConsulTemplateDestNum)
		for tmplIdx := 0; tmplIdx < annotationConsulTemplateDestNum; tmplIdx++ {
			annotationConsulTemplateTemplate[tmplIdx] = vaultInjector.CtTemplateDefaultTmpl
		}
	} else {
		// We must have same numbers of custom templates & secrets destinations ...
		if annotationConsulTemplateDestNum != annotationConsulTemplateTemplateNum {
			klog.Error("Submitted pod must contain same numbers of templates and secrets destinations")
			return nil, errors.New("Submitted pod must contain same numbers of templates and secrets destinations")
		}

		// ... and we ignore content of 'secrets-path' annotation ('cause we provide full template), but we need to init an empty array
		// to not end up with errors in the replace loop to come
		annotationVaultSecretsPath = make([]string, annotationConsulTemplateDestNum)
	}

	// Copy provided CT commands, if less commands than secrets destinations: remaining commands set to ""
	consulTemplateCommands := make([]string, annotationConsulTemplateDestNum)
	copy(consulTemplateCommands, annotationConsulTemplateCmd)

	var ctTemplateBlock string
	var ctTemplates strings.Builder

	for tmplIdx := 0; tmplIdx < annotationConsulTemplateDestNum; tmplIdx++ {
		ctTemplateBlock = vaultInjector.CtTemplateBlock
		ctTemplateBlock = strings.Replace(ctTemplateBlock, consulTemplateAppSvcDestinationPlaceholder, annotationConsulTemplateDest[tmplIdx], -1)
		ctTemplateBlock = strings.Replace(ctTemplateBlock, consulTemplateTemplateContentPlaceholder, annotationConsulTemplateTemplate[tmplIdx], -1)
		ctTemplateBlock = strings.Replace(ctTemplateBlock, vaultAppSvcSecretsPathPlaceholder, annotationVaultSecretsPath[tmplIdx], -1)
		ctTemplateBlock = strings.Replace(ctTemplateBlock, consulTemplateCommandPlaceholder, consulTemplateCommands[tmplIdx], -1)

		ctTemplates.WriteString(ctTemplateBlock)
		ctTemplates.WriteString("\n")
	}

	return &sidecarPlaceholders{k8sSaSecretsVolName, annotationVaultRole, annotationVaultSATokenPath, ctTemplates.String()}, nil
}

// Deal with both InitContainers & Containers
func (vaultInjector *VaultInjector) addContainer(podContainers []corev1.Container, annotations map[string]string, basePath string, placeholders sidecarPlaceholders) (patch []patchOperation, err error) {
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
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], vaultRolePlaceholder, placeholders.vaultRole, -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], vaultAppSvcSATokenPathPlaceholder, placeholders.vaultSATokenPath, -1)
			container.Command[commandIdx] = strings.Replace(container.Command[commandIdx], consulTemplateTemplatesPlaceholder, placeholders.consulTemplateTemplates, -1)
		}

		// We will modify some values here so make a copy to not change origin
		container.VolumeMounts = make([]corev1.VolumeMount, len(sidecarCnt.VolumeMounts))
		copy(container.VolumeMounts, sidecarCnt.VolumeMounts)

		// Loop to see if we have a '/var/run/secrets/kubernetes.io/serviceaccount' mountPath
		// then overwrite its name with extracted value from submitted pod
		for volMountIdx := range container.VolumeMounts {
			if container.VolumeMounts[volMountIdx].MountPath == k8sServiceAccountTokenVolMountPath {
				container.VolumeMounts[volMountIdx].Name = placeholders.serviceAccountTokenVolumeName
				break
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

	var value interface{}

	for _, sidecarVol := range vaultInjector.SidecarConfig.Volumes {
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

// create mutation patch for resoures
func (vaultInjector *VaultInjector) createPatch(pod *corev1.Pod, annotations map[string]string) ([]byte, error) {

	patchPodSpec, err := vaultInjector.updatePodSpec(pod)
	if err != nil {
		return nil, err
	}

	var patch []patchOperation

	patch = append(patch, patchPodSpec...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}

// main mutation process
func (vaultInjector *VaultInjector) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	var podName string
	var podNamespace string

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("Request=%+v", req)
	}

	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("Pod=%+v", pod)
	}

	if pod.Name == "" {
		podName = pod.GenerateName
	} else {
		podName = pod.Name
	}

	if pod.Namespace == "" {
		podNamespace = metav1.NamespaceDefault
	} else {
		podNamespace = pod.Namespace
	}

	klog.Infof("AdmissionReview for GroupVersionKind=%+v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%+v",
		req.Kind, req.Namespace, req.Name, podName, req.UID, req.Operation, req.UserInfo)

	// determine whether to perform mutation
	if !mutationRequired(ignoredNamespaces, vaultInjector.VaultInjectorAnnotationsFQ, &pod.ObjectMeta) {
		klog.Infof("Skipping mutation for %s/%s due to policy check", podNamespace, podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	annotations := map[string]string{vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationStatusKey]: vaultInjectorAnnotationStatusValue}
	patchBytes, err := vaultInjector.createPatch(&pod, annotations)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	klog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook server
func (vaultInjector *VaultInjector) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = vaultInjector.mutate(&ar)
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
