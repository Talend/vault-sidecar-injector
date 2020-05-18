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
	"syscall"
	"talend/vault-sidecar-injector/pkg/config"
	"talend/vault-sidecar-injector/pkg/webhook"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

var (
	// VERSION stores current version. Set in Makefile (see build flag -ldflags "-X=main.VERSION=$(VERSION)")
	VERSION string
)

func main() {
	var parameters config.WhSvrParameters

	// get command line parameters
	flag.IntVar(&parameters.Port, "port", 8443, "webhook server port")
	flag.IntVar(&parameters.MetricsPort, "metricsport", 9000, "metrics server port (Prometheus)")
	flag.StringVar(&parameters.CertFile, "tlscertfile", config.CertsPath+"/cert.pem", "file containing the x509 Certificate for HTTPS")
	flag.StringVar(&parameters.KeyFile, "tlskeyfile", config.CertsPath+"/key.pem", "file containing the x509 private key to tlscertfile")
	flag.StringVar(&parameters.AnnotationKeyPrefix, "annotationkeyprefix", "sidecar.vault", "annotations key prefix")
	flag.StringVar(&parameters.AppLabelKey, "applabelkey", "application.name", "key for application label")
	flag.StringVar(&parameters.AppServiceLabelKey, "appservicelabelkey", "service.name", "key for application's service label")
	flag.StringVar(&parameters.InjectionCfgFile, "injectioncfgfile", config.ConfigFilesPath+"/injectionconfig.yaml", "file containing the mutation configuration (initcontainers, sidecars, volumes, ...)")
	flag.StringVar(&parameters.ProxyCfgFile, "proxycfgfile", config.ConfigFilesPath+"/proxyconfig.hcl", "file containing Vault proxy configuration")
	flag.StringVar(&parameters.TemplateBlockFile, "tmplblockfile", config.ConfigFilesPath+"/templateblock.hcl", "file containing the template block")
	flag.StringVar(&parameters.TemplateDefaultFile, "tmpldefaultfile", config.ConfigFilesPath+"/templatedefault.tmpl", "file containing the default template")
	flag.StringVar(&parameters.PodLifecycleHooksFile, "podlchooksfile", config.ConfigFilesPath+"/podlifecyclehooks.yaml", "file containing the lifecycle hooks to inject in the requesting pod")
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

	// Load webhook admission server's config
	vsiCfg, err := config.Load(parameters)
	if err != nil {
		os.Exit(1)
	}

	pair, err := tls.LoadX509KeyPair(parameters.CertFile, parameters.KeyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		os.Exit(1)
	}

	vaultInjector := webhook.New(
		vsiCfg,
		&http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	)

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", vaultInjector.Serve)
	vaultInjector.Server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := vaultInjector.Server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	// define metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", parameters.MetricsPort),
		Handler: metricsMux,
	}

	// start metrics server in new routine
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			klog.Errorf("Failed to listen and serve metrics server: %v", err)
		}
	}()

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	vaultInjector.Server.Shutdown(context.Background())
	metricsServer.Shutdown(context.Background())
}
