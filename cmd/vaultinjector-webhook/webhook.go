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
	"crypto/tls"
	"fmt"
	"net/http"
	"talend/vault-sidecar-injector/pkg/config"
	"talend/vault-sidecar-injector/pkg/k8s"
	"talend/vault-sidecar-injector/pkg/webhook"

	"k8s.io/klog"
)

func createVaultInjector() (*webhook.VaultInjector, error) {
	// Patch MutatingWebhookConfiguration resource with CA certificate from mounted secret (set 'caBundle' attribute from Webhook CA)
	err := k8s.New(
		&k8s.WebhookData{
			WebhookCfgName: webhookParameters.WebhookCfgName,
		}).PatchWebhookConfiguration(webhookParameters.CACertFile)
	if err != nil {
		return nil, err
	}

	// Load TLS cert and key from mounted secret
	tlsCert, err := tls.LoadX509KeyPair(webhookParameters.CertFile, webhookParameters.KeyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		return nil, err
	}

	// Load webhook admission server's config
	vsiCfg, err := config.Load(webhookParameters)
	if err != nil {
		return nil, err
	}

	vaultInjector := webhook.New(
		vsiCfg,
		&http.Server{
			Addr:      fmt.Sprintf(":%v", webhookParameters.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
		},
	)

	return vaultInjector, nil
}
