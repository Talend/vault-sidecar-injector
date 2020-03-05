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

package mode

import (
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"

	corev1 "k8s.io/api/core/v1"
)

// VaultInjectorModes : modes
var VaultInjectorModes = make(map[string]VaultInjectorModeInfo)

// VaultInjectorModeInfo : mode info
type VaultInjectorModeInfo struct {
	Key                  string   // mode key (== mode's name or id)
	DefaultMode          bool     // mode to enable when no mode explicitly requested in incoming manifest
	EnableDefaultMode    bool     // should we also enable the default mode when this mode is the only one being requested?
	Annotations          []string // mode's annotations
	ComputeTemplatesFunc func(
		config *cfg.VSIConfig,
		labels,
		annotations map[string]string) (ctx.ModeConfig, error) // to compute templates used in injected container(s)
	PatchPodFunc func(
		config *cfg.VSIConfig,
		podSpec corev1.PodSpec,
		annotations map[string]string,
		context *ctx.InjectionContext) (patch []ctx.PatchOperation, err error) // to patch submitted pod's container(s)
	InjectContainerFunc func(
		containerBasePath string,
		podContainers []corev1.Container,
		containerName string,
		env []corev1.EnvVar,
		context *ctx.InjectionContext) (bool, error) // to test if container should be injected and, if so, resolve env vars
}
