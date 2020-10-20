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

	"k8s.io/klog"
)

func parseFlags() string {
	certCmd := flag.NewFlagSet(CertCmd, flag.ExitOnError)
	certCmd.StringVar(&certParameters.CertOperation, "certop", CreateCert, "operation on webhook certificates (create, delete)")
	certCmd.StringVar(&certParameters.CertSecretName, "certsecretname", "", "name of generated or provided Kubernetes secret storing webhook certificates and private key")
	certCmd.StringVar(&certParameters.CertHostnames, "certhostnames", "", "host names to register in webhook certificate (comma-separated list)")
	certCmd.IntVar(&certParameters.CertLifetime, "certlifetime", 10, "lifetime in years for generated certificates")

	inlineCmd := flag.NewFlagSet(InlineCmd, flag.ExitOnError)
	inlineCmd.StringVar(&inlineParameters.Manifest, "manifest", "", "manifest as input for inline injection")

	webhookCmd := flag.NewFlagSet(WebhookCmd, flag.ExitOnError)
	webhookCmd.IntVar(&webhookParameters.Port, "port", 8443, "webhook server port")
	webhookCmd.IntVar(&webhookParameters.MetricsPort, "metricsport", 9000, "metrics server port (Prometheus)")
	webhookCmd.StringVar(&webhookParameters.CACertFile, "cacertfile", "", "PEM-encoded webhook CA certificate")
	webhookCmd.StringVar(&webhookParameters.CertFile, "certfile", "", "PEM-encoded webhook certificate used for TLS")
	webhookCmd.StringVar(&webhookParameters.KeyFile, "keyfile", "", "PEM-encoded webhook private key used for TLS")
	webhookCmd.StringVar(&webhookParameters.WebhookCfgName, "webhookcfgname", "", "name of MutatingWebhookConfiguration resource")
	webhookCmd.StringVar(&webhookParameters.AnnotationKeyPrefix, "annotationkeyprefix", "sidecar.vault", "annotations key prefix")
	webhookCmd.StringVar(&webhookParameters.AppLabelKey, "applabelkey", "application.name", "key for application label")
	webhookCmd.StringVar(&webhookParameters.AppServiceLabelKey, "appservicelabelkey", "service.name", "key for application's service label")
	webhookCmd.StringVar(&webhookParameters.InjectionCfgFile, "injectioncfgfile", "", "file containing the mutation configuration (initcontainers, sidecars, volumes, ...)")
	webhookCmd.StringVar(&webhookParameters.ProxyCfgFile, "proxycfgfile", "", "file containing Vault proxy configuration")
	webhookCmd.StringVar(&webhookParameters.TemplateBlockFile, "tmplblockfile", "", "file containing the template block")
	webhookCmd.StringVar(&webhookParameters.TemplateDefaultFile, "tmpldefaultfile", "", "file containing the default template")
	webhookCmd.StringVar(&webhookParameters.PodLifecycleHooksFile, "podlchooksfile", "", "file containing the lifecycle hooks to inject in the requesting pod")

	versionCmd := flag.NewFlagSet(VersionCmd, flag.ExitOnError)
	//flag.Parse()

	// Beware as glog is here behind the scene and we use klog here
	// So logging command line parameters -v, -logtostderr, -alsologtostderr, -log_dir, -log_file, ... are already initialized (see klog init() func)
	//klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)

	// Sync the glog and klog flags
	/*flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			f2.Value.Set(value)
		}
	})*/

	if len(os.Args) == 1 {
		usage(os.Args[0])
		os.Exit(1)
	}

	vsiSubcommand := os.Args[1]

	switch vsiSubcommand {
	case CertCmd:
		klog.InitFlags(certCmd)
		certCmd.Parse(os.Args[2:])
	case InlineCmd:
		klog.InitFlags(inlineCmd)
		inlineCmd.Parse(os.Args[2:])
	case WebhookCmd:
		klog.InitFlags(webhookCmd)
		webhookCmd.Parse(os.Args[2:])
	case VersionCmd:
		versionCmd.Parse(os.Args[2:])
		fmt.Println("VSI (Vault Sidecar Injector) version " + VERSION)
		os.Exit(0)
	default:
		usage(os.Args[0])
		os.Exit(1)
	}

	return vsiSubcommand
}

func usage(program string) {
	fmt.Printf("Usage: %s <command> [<args>]\n\nCommands:\n", program)
	fmt.Printf("  %s\n", CertCmd)
	fmt.Printf("  %s\n", InlineCmd)
	fmt.Printf("  %s\n", WebhookCmd)
	fmt.Printf("  %s\n", VersionCmd)
}
