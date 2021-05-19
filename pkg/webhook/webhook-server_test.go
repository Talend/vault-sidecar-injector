// Copyright © 2019-2021 Talend - www.talend.com
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

package webhook

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog"
)

func TestWebhookServer(t *testing.T) {
	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	if verbose {
		// Set Klog verbosity level to have detailed logs from our webhook (where we use level 5+ to log such info)
		klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
		klog.InitFlags(klogFlags)
		klogFlags.Set("v", "5")
	}

	// Create webhook instance
	vaultInjector, err := createTestVaultInjector()
	if err != nil {
		t.Fatalf("Loading error: %s", err)
	}

	tables := []struct {
		name                   string
		admissionReviewVersion string
		vaultInjection         bool
		statusCode             int
	}{
		{
			name:                   "AdmissionReview v1, no injection",
			admissionReviewVersion: "v1",
			vaultInjection:         false,
			statusCode:             http.StatusOK,
		},
		{
			name:                   "AdmissionReview v1",
			admissionReviewVersion: "v1",
			vaultInjection:         true,
			statusCode:             http.StatusOK,
		},
		{
			name:                   "AdmissionReview v1beta1",
			admissionReviewVersion: "v1beta1",
			vaultInjection:         true,
			statusCode:             http.StatusOK,
		},
		{
			name:                   "AdmissionReview v1beta2",
			admissionReviewVersion: "v1beta2",
			vaultInjection:         true,
			statusCode:             http.StatusBadRequest,
		},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			uid := string(uuid.NewUUID())
			request := httptest.NewRequest(http.MethodPost, "/mutate", strings.NewReader(`{
				"kind":"AdmissionReview",
				"apiVersion":"admission.k8s.io/`+table.admissionReviewVersion+`",
				"request":{
				  "uid":"`+uid+`",
				  "kind":{
					"group":"",
					"version":"v1",
					"kind":"Pod"
				  },
				  "namespace":"default",
				  "operation":"CREATE",
				  "object":{
					"apiVersion":"v1",
					"kind":"Pod",
					"metadata":{
						"annotations":{
							"sidecar.vault.talend.org/inject": "`+strconv.FormatBool(table.vaultInjection)+`"
						},
						"labels":{
							"com.talend.application": "test",
							"com.talend.service": "test-app-svc"
						}
					},
					"spec":{
						"containers":[
							{
								"name": "testcontainer",
								"image": "myfakeimage:1.0.0",
								"volumeMounts":[
									{
										"name": "default-token-1234",
										"mountPath" : "/var/run/secrets/kubernetes.io/serviceaccount"
									}
								]
							}
						]
					}
				  }
				}
			  }`))
			request.Header.Add("Content-Type", "application/json")
			responseRecorder := httptest.NewRecorder()

			vaultInjector.Serve(responseRecorder, request)

			if klog.V(5) {
				klog.Infof("HTTP Response=%+v", responseRecorder)
			}

			assert.Equal(t, responseRecorder.Code, table.statusCode)
			assert.Condition(t, func() bool {
				if responseRecorder.Code == http.StatusOK {
					return strings.Contains(responseRecorder.Body.String(),
						`"kind":"AdmissionReview","apiVersion":"admission.k8s.io/`+table.admissionReviewVersion+`"`) &&
						strings.Contains(responseRecorder.Body.String(),
							`"response":{"uid":"`+uid+`"`)
				} else {
					return true // HTTP error: return true to skip this test
				}
			}, "AdmissionReview version must match received version and admission response UID must match admission request UID")
		})
	}
}
