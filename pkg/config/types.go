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

package config

import (
	corev1 "k8s.io/api/core/v1"
)

// WhSvrParameters : Webhook Server parameters
type WhSvrParameters struct {
	Port                  int    // webhook server port
	MetricsPort           int    // metrics server port (Prometheus)
	CertGeneration        bool   // generates webhook certificates, private key and Kubernetes secret
	CertSecretName        string // name of generated or provided Kubernetes secret storing webhook certificates and private key
	CertHostnames         string // host names to register in webhook certificate (comma-separated list)
	CertLifetime          int    // lifetime in years for generated certificates
	CACertFile            string // PEM-encoded webhook CA certificate
	CertFile              string // PEM-encoded webhook certificate used for TLS
	KeyFile               string // PEM-encoded webhook private key used for TLS
	WebhookCfgName        string // name of MutatingWebhookConfiguration resource
	AnnotationKeyPrefix   string // annotations key prefix
	AppLabelKey           string // key for application label
	AppServiceLabelKey    string // key for application's service label
	InjectionCfgFile      string // path to injection configuration file
	ProxyCfgFile          string // path to Vault proxy configuration file
	TemplateBlockFile     string // path to template file
	TemplateDefaultFile   string // path to default template content file
	PodLifecycleHooksFile string // path to pod's lifecycle hooks file
}

// InjectionConfig : resources that will be injected (read from config file)
type InjectionConfig struct {
	InitContainers []corev1.Container `yaml:"initContainers" json:"initContainers"`
	Containers     []corev1.Container `yaml:"containers" json:"containers"`
	Volumes        []corev1.Volume    `yaml:"volumes" json:"volumes"`
}

// LifecycleHooks : lifecycle hooks to inject in requesting pod
type LifecycleHooks struct {
	PostStart *corev1.Handler `yaml:"postStart" json:"postStart"`
}

// VSIConfig : Vault Sidecar Injector configuration
type VSIConfig struct {
	VaultInjectorAnnotationKeyPrefix string            // annotations prefix
	VaultInjectorAnnotationsFQ       map[string]string // supported annotations (fully-qualified with prefix if any)
	ApplicationLabelKey              string            // key for application label
	ApplicationServiceLabelKey       string            // key for application's service label
	InjectionConfig                  *InjectionConfig  // injection configuration
	ProxyConfig                      string            // Vault proxy configuration
	TemplateBlock                    string            // template
	TemplateDefaultTmpl              string            // default template content
	PodslifecycleHooks               *LifecycleHooks   // pod's lifecycle hooks
}
