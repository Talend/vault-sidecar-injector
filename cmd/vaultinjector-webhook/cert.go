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
	"strings"
	"talend/vault-sidecar-injector/pkg/certs"
	"talend/vault-sidecar-injector/pkg/k8s"
)

func genCertificates() error {
	// Generate certificates and key
	cert := &certs.Cert{
		CN:       "Vault Sidecar Injector",
		Hosts:    strings.Split(certParameters.CertHostnames, ","),
		Lifetime: certParameters.CertLifetime,
	}

	bundle, err := cert.GenerateWebhookBundle()
	if err != nil {
		return err
	}

	// Create Kubernetes Secret
	return k8s.New(
		&k8s.WebhookData{
			WebhookSecretName: certParameters.CertSecretName,
			WebhookCACertName: certParameters.CACertFile,
			WebhookCertName:   certParameters.CertFile,
			WebhookKeyName:    certParameters.KeyFile,
		}).CreateCertSecret(bundle.CACert, bundle.Cert, bundle.PrivKey)
}

func deleteCertificates() error {
	// Delete Kubernetes Secret
	return k8s.New(
		&k8s.WebhookData{
			WebhookSecretName: certParameters.CertSecretName,
		}).DeleteCertSecret()
}
