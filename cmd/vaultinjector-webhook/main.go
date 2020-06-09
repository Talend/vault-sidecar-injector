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
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"talend/vault-sidecar-injector/pkg/certs"
	"talend/vault-sidecar-injector/pkg/config"
	"talend/vault-sidecar-injector/pkg/k8s"
	"talend/vault-sidecar-injector/pkg/webhook"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

var (
	// VERSION stores current version. Set in Makefile (see build flag -ldflags "-X=main.VERSION=$(VERSION)")
	VERSION string

	parameters config.WhSvrParameters
)

func parseFlags() {
	flag.IntVar(&parameters.Port, "port", 8443, "webhook server port")
	flag.IntVar(&parameters.MetricsPort, "metricsport", 9000, "metrics server port (Prometheus)")
	flag.StringVar(&parameters.CertOperation, "certop", "", "operation on webhook certificates (create, delete)")
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

func main() {
	parseFlags()

	if parameters.CertOperation != "" {
		switch parameters.CertOperation {
		case config.CreateCert: // Generate certificates, private key, K8S secret
			if err := genCertificates(); err != nil {
				os.Exit(1)
			}
		case config.DeleteCert: // Delete K8S secret used to store certificates and private key
			if err := deleteCertificates(); err != nil {
				os.Exit(1)
			}
		}
	} else {
		// Init and load config
		vaultInjector, err := createVaultInjector()
		if err != nil {
			os.Exit(1)
		}

		// Define http server and server handler
		mux := http.NewServeMux()
		mux.HandleFunc("/mutate", vaultInjector.Serve)
		vaultInjector.Server.Handler = mux

		// Start webhook server in new routine
		go func() {
			if err := vaultInjector.Server.ListenAndServeTLS("", ""); err != nil {
				klog.Errorf("Failed to listen and serve webhook server: %v", err)
			}
		}()

		// Define metrics server
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())

		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%v", parameters.MetricsPort),
			Handler: metricsMux,
		}

		// Start metrics server in new routine
		go func() {
			if err := metricsServer.ListenAndServe(); err != nil {
				klog.Errorf("Failed to listen and serve metrics server: %v", err)
			}
		}()

		// Listening OS shutdown singal
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan

		klog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
		vaultInjector.Server.Shutdown(context.Background())
		metricsServer.Shutdown(context.Background())
	}
}

func createVaultInjector() (*webhook.VaultInjector, error) {
	// Patch MutatingWebhookConfiguration (set 'caBundle' attribute from Webhook CA)
	err := patchWebhookConfig()
	if err != nil {
		return nil, err
	}

	// Load TLS cert and key from mounted secret
	tlsCert, err := tls.LoadX509KeyPair(parameters.CertFile, parameters.KeyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		return nil, err
	}

	// Load webhook admission server's config
	vsiCfg, err := config.Load(parameters)
	if err != nil {
		return nil, err
	}

	vaultInjector := webhook.New(
		vsiCfg,
		&http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
		},
	)

	return vaultInjector, nil
}

func genCertificates() error {
	// Generate certificates and key
	cert := &certs.Cert{
		CN:       "Vault Sidecar Injector",
		Hosts:    strings.Split(parameters.CertHostnames, ","),
		Lifetime: parameters.CertLifetime,
	}

	bundle, err := cert.GenerateWebhookBundle()
	if err != nil {
		return err
	}

	k8sClient := k8s.New(&k8s.WebhookData{
		WebhookSecretName: parameters.CertSecretName,
		WebhookCACertName: parameters.CACertFile,
		WebhookCertName:   parameters.CertFile,
		WebhookKeyName:    parameters.KeyFile,
	})

	// Create Kubernetes Secret
	return k8sClient.CreateCertSecret(bundle.CACert, bundle.Cert, bundle.PrivKey)
}

func deleteCertificates() error {
	k8sClient := k8s.New(&k8s.WebhookData{
		WebhookSecretName: parameters.CertSecretName,
	})

	// Delete Kubernetes Secret
	return k8sClient.DeleteCertSecret()
}

func patchWebhookConfig() error {
	k8sClient := k8s.New(&k8s.WebhookData{
		WebhookCfgName: parameters.WebhookCfgName,
	})

	// Patch MutatingWebhookConfiguration resource with CA certificate from mounted secret
	return k8sClient.PatchWebhookConfiguration(parameters.CACertFile)
}
