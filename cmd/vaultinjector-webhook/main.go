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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"talend/vault-sidecar-injector/pkg/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

var (
	// VERSION stores current version. Set in Makefile (see build flag -ldflags "-X=main.VERSION=$(VERSION)")
	VERSION string

	parameters config.WhSvrParameters
)

func main() {
	parseFlags()

	switch parameters.Mode {
	case config.CertMode:
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
	case config.InlineMode:
		//TODO
	case config.WebhookMode:
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
