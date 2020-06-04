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
	"crypto/tls"
	"flag"
	"os"
	"strconv"
	"strings"
	"testing"

	"k8s.io/klog"
)

func TestCert(t *testing.T) {
	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	if verbose {
		// Set Klog verbosity level to have detailed logs from our webhook (where we use level 5+ to log such info)
		klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
		klog.InitFlags(klogFlags)
		klogFlags.Set("v", "5")
	}

	cert := &Cert{
		CN:       "Vault Sidecar Injector",
		Hosts:    strings.Split("vault-sidecar-injector,vault-sidecar-injector.default,vault-sidecar-injector.default.svc", ","),
		Lifetime: 10,
	}

	bundle, err := cert.GenerateWebhookBundle()
	if err != nil {
		t.Fatalf("Failed to generate webhook bundle: %v", err)
	}

	tlsCert, err := tls.X509KeyPair(bundle.Cert, bundle.PrivKey)
	if err != nil {
		t.Fatalf("Failed to load key pair: %v", err)
	}

	klog.Infof("tlsCert=%+v", tlsCert)
}
