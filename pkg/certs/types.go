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

package certs

import (
	"crypto"
	"crypto/x509"
)

// PEMBundle stores webhook certificates and private key
type PEMBundle struct {
	CACert  []byte // CA certificate (PEM-encoded)
	Cert    []byte // Webhook certificate (signed by CA) (PEM-encoded)
	PrivKey []byte // Webhook private key (PEM-encoded)
}

// Cert holds all useful data for certificate generation
type Cert struct {
	CN         string            // Common Name
	Hosts      []string          // Host names to register in webhook certificate
	Lifetime   int               // Certificate lifetime (years)
	caCert     []byte            // CA certificate (PEM-encoded)
	caTemplate *x509.Certificate // CA template (used to generate webhook certificate)
	caPrivKey  crypto.Signer     // CA private key
}
