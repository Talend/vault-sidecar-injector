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
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"k8s.io/klog"
)

// GenerateWebhookBundle generates and returns CA, webhook certificate and private key
func (c *Cert) GenerateWebhookBundle() (*PEMBundle, error) {
	// CA
	_, err := c.genCertificate(true)
	if err != nil {
		klog.Errorf("Failed to generate CA certificate: %s", err)
		return nil, err
	}

	// Webhook certificate
	webhookBundle, err := c.genCertificate(false)
	if err != nil {
		klog.Errorf("Failed to generate webhook certificate: %s", err)
		return nil, err
	}

	if klog.V(5) { // enabled by providing '-v=5' at least
		klog.Infof("Generated Webhook CA Certificate: %s", string(webhookBundle.CACert))
		klog.Infof("Generated Webhook Certificate: %s", string(webhookBundle.Cert))
	}

	return webhookBundle, nil
}

func (c *Cert) genCertificate(isCA bool) (*PEMBundle, error) {
	var derBytes []byte
	var certBnd PEMBundle
	var keyUsage x509.KeyUsage

	cn := c.CN
	notBefore := time.Now().Add(time.Minute * -5)
	notAfter := notBefore.Add(time.Duration(c.Lifetime) * 365 * 24 * time.Hour)

	sn, err := serialNumber()
	if err != nil {
		klog.Errorf("Failed to generate serial number: %s", err)
		return nil, err
	}

	if isCA {
		cn = cn + " CA"
		keyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	} else {
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	}

	template := x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	//-- Generate key pair
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		klog.Errorf("Failed ECDSA Key Generation: %s", err)
		return nil, err
	}

	//-- Generate certificate
	if isCA { // Self-signed
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
		c.caTemplate = &template
		c.caPrivKey = key
	} else {
		for _, h := range c.Hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}

		derBytes, err = x509.CreateCertificate(rand.Reader, &template, c.caTemplate, &key.PublicKey, c.caPrivKey)
	}

	if err != nil {
		klog.Errorf("Failed to create certificate: %s", err)
		return nil, err
	}

	//-- PEM-encode certificate and private key
	pemCert, err := pemEncodeCert(derBytes)
	if err != nil {
		klog.Errorf("Failed to encode certificate: %s", err)
		return nil, err
	}

	pemPrivkey, err := pemEncodeKey(key)
	if err != nil {
		klog.Errorf("Failed to encode private key: %s", err)
		return nil, err
	}

	if isCA { // Self-signed
		certBnd = PEMBundle{CACert: pemCert, Cert: pemCert, PrivKey: pemPrivkey}
		c.caCert = pemCert
	} else {
		certBnd = PEMBundle{CACert: c.caCert, Cert: pemCert, PrivKey: pemPrivkey}
	}

	return &certBnd, nil
}

func serialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
}

func pemEncodeKey(key *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func pemEncodeCert(derBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
