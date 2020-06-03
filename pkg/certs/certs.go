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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"time"

	"k8s.io/klog"
)

// GenerateCerts generates a CA with a leaf certificate and key and returns the CA, cert and key as PEM encoded slices
func GenerateCerts(host string) (ca, cert, key []byte) {
	notBefore := time.Now().Add(time.Minute * -5)
	notAfter := notBefore.Add(5 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		klog.Errorf("Failed to generate serial number: %s", err)
	}
	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		klog.Errorf("Failed ECDSA Key Generation: %s", err)
	}

	rootTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		klog.Errorf("Failed createCertificate for CA: %s", err)
	}

	ca = encodeCert(derBytes)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		klog.Errorf("Failed createLeafKey for certificate: %s", err)
	}

	key = encodeKey(leafKey)

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		klog.Errorf("Failed to generate serial number: %s", err)
	}

	leafTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	hosts := strings.Split(host, ",")

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			leafTemplate.IPAddresses = append(leafTemplate.IPAddresses, ip)
		} else {
			leafTemplate.DNSNames = append(leafTemplate.DNSNames, h)
		}
	}

	derBytes, err = x509.CreateCertificate(rand.Reader, &leafTemplate, &rootTemplate, &leafKey.PublicKey, rootKey)
	if err != nil {
		klog.Errorf("Failed createLeaf certificate: %s", err)
	}

	cert = encodeCert(derBytes)
	return ca, cert, key
}

func encodeKey(key *ecdsa.PrivateKey) []byte {
	b, err := x509.MarshalECPrivateKey(key)

	if err != nil {
		klog.Errorf("Unable to marshal ECDSA private key: %s", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
}

func encodeCert(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}
