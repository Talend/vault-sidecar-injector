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

package main

import (
	"flag"
	"fmt"
	"os"
	"talend/vault-sidecar-injector/pkg/config"

	"k8s.io/klog"
)

func parseFlags() {
	flag.StringVar(&parameters.Mode, "mode", config.WebhookMode, "operational mode (cert, inline, webhook)")
	flag.IntVar(&parameters.Port, "port", 8443, "webhook server port")
	flag.IntVar(&parameters.MetricsPort, "metricsport", 9000, "metrics server port (Prometheus)")
	flag.StringVar(&parameters.Manifest, "manifest", "", "manifest as input for inline injection")
	flag.StringVar(&parameters.CertOperation, "certop", config.CreateCert, "operation on webhook certificates (create, delete)")
	flag.StringVar(&parameters.CertSecretName, "certsecretname", "", "name of generated or provided Kubernetes secret storing webhook certificates and private key")
	flag.StringVar(&parameters.CertHostnames, "certhostnames", "", "host names to register in webhook certificate (comma-separated list)")
	flag.IntVar(&parameters.CertLifetime, "certlifetime", 10, "lifetime in years for generated certificates")
	flag.StringVar(&parameters.CACertFile, "cacertfile", "", "PEM-encoded webhook CA certificate")
	flag.StringVar(&parameters.CertFile, "certfile", "", "PEM-encoded webhook certificate used for TLS")
	flag.StringVar(&parameters.KeyFile, "keyfile", "", "PEM-encoded webhook private key used for TLS")
	flag.StringVar(&parameters.WebhookCfgName, "webhookcfgname", "", "name of MutatingWebhookConfiguration resource")
	flag.StringVar(&parameters.AnnotationKeyPrefix, "annotationkeyprefix", "sidecar.vault", "annotations key prefix")
	flag.StringVar(&parameters.AppLabelKey, "applabelkey", "application.name", "key for application label")
	flag.StringVar(&parameters.AppServiceLabelKey, "appservicelabelkey", "service.name", "key for application's service label")
	flag.StringVar(&parameters.InjectionCfgFile, "injectioncfgfile", "", "file containing the mutation configuration (initcontainers, sidecars, volumes, ...)")
	flag.StringVar(&parameters.ProxyCfgFile, "proxycfgfile", "", "file containing Vault proxy configuration")
	flag.StringVar(&parameters.TemplateBlockFile, "tmplblockfile", "", "file containing the template block")
	flag.StringVar(&parameters.TemplateDefaultFile, "tmpldefaultfile", "", "file containing the default template")
	flag.StringVar(&parameters.PodLifecycleHooksFile, "podlchooksfile", "", "file containing the lifecycle hooks to inject in the requesting pod")
	version := flag.Bool("version", false, "print current version")
	flag.Parse()

	if *version {
		fmt.Println("\nVSI (Vault Sidecar Injector) version " + VERSION)
		os.Exit(0)
	}

	// Beware as glog is here behind the scene and we use klog here
	// So logging command line parameters -v, -logtostderr, -alsologtostderr, -log_dir, -log_file, ... are already initialized (see glog init() func)
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})
}
