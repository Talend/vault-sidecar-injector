// Copyright © 2019 Talend
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
	"talend/vault-sidecar-injector/pkg/config"
	"testing"

	"k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

type inputMutate struct {
	ar *v1beta1.AdmissionReview
}

type expectedMutate struct {
	patchExpected []patchOperation
}

func TestMutate(t *testing.T) {
	// Need to load some conf to know mutations to apply (load function has its own dedicated test see 'load_test.go')
	injectionCfg, err := config.Load(
		config.WhSvrParameters{
			0, 0, "", "",
			"sidecar.vault.talend.org", "com.talend.application", "com.talend.service",
			"../../test/sidecarconfig.yaml",
			"../../test/tmplblock.hcl",
			"../../test/tmpldefault.tmpl",
			"../../test/podlifecyclehooks.yaml",
		},
	)
	if err != nil {
		t.Errorf("Loading error \"%s\"", err)
	}

	vaultInjector := New(injectionCfg, nil)

	// Our test struct
	tables := []struct {
		inputMutate
		expectedMutate
	}{
		{
			inputMutate{
				// Inject captured AdmissionReview struct (init useful fields only. Mandatory one is 'Object.Raw', others only used in log outputs)
				&v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						UID: "d6d378c8-6a84-11e9-bfcc-e8b9f80c46b6",
						Kind: metav1.GroupVersionKind{
							Version: "v1",
							Kind:    "Pod",
						},
						Namespace: "default",
						Operation: "CREATE",
						UserInfo: authenticationv1.UserInfo{
							Username: "system:serviceaccount:kube-system:replicaset-controller",
							UID:      "d13bb3f4-59d5-11e9-939e-e8b9f80c46b6",
							Groups:   []string{"system:serviceaccounts", "system:serviceaccounts:kube-system", "system:authenticated"},
						},
						Object: runtime.RawExtension{
							/*
								Encoding below is for following manifest:
								=========================================
									apiVersion: apps/v1
									kind: Deployment
									metadata:
										name: test-app-1
										namespace: default
									spec:
										replicas: 1
										selector:
											matchLabels:
												com.talend.application: test-app-1
												com.talend.service: test-app-1-svc
										template:
											metadata:
												annotations:
													sidecar.vault.talend.org/inject: "true"
												labels:
													com.talend.application: test-app-1
													com.talend.service: test-app-1-svc
											spec:
												serviceAccountName: app-1-sa
												containers:
													- name: test-app-1-container
													  image: busybox:1.28
													  command:
														- "sh"
														- "-c"
														- >
														while true;do echo "My secrets are: $(cat /opt/talend/secrets/secrets.properties)"; sleep 5; done
													  volumeMounts:
														- name: secrets
														  mountPath: /opt/talend/secrets
												volumes:
													- name: secrets
													  emptyDir:
														medium: Memory
							*/
							Raw: []byte{123, 34, 107, 105, 110, 100, 34, 58, 34, 80, 111, 100, 34, 44, 34, 97, 112, 105, 86, 101, 114, 115, 105, 111, 110, 34, 58, 34, 118, 49, 34, 44, 34, 109, 101, 116, 97, 100, 97, 116, 97, 34, 58, 123, 34, 103, 101, 110, 101, 114, 97, 116, 101, 78, 97, 109, 101, 34, 58, 34, 116, 101, 115, 116, 45, 97, 112, 112, 45, 49, 45, 55, 100, 56, 102, 56, 52, 98, 52, 52, 52, 45, 34, 44, 34, 99, 114, 101, 97, 116, 105, 111, 110, 84, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 110, 117, 108, 108, 44, 34, 108, 97, 98, 101, 108, 115, 34, 58, 123, 34, 99, 111, 109, 46, 116, 97, 108, 101, 110, 100, 46, 97, 112, 112, 108, 105, 99, 97, 116, 105, 111, 110, 34, 58, 34, 116, 101, 115, 116, 45, 97, 112, 112, 45, 49, 34, 44, 34, 99, 111, 109, 46, 116, 97, 108, 101, 110, 100, 46, 115, 101, 114, 118, 105, 99, 101, 34, 58, 34, 116, 101, 115, 116, 45, 97, 112, 112, 45, 49, 45, 115, 118, 99, 34, 44, 34, 112, 111, 100, 45, 116, 101, 109, 112, 108, 97, 116, 101, 45, 104, 97, 115, 104, 34, 58, 34, 55, 100, 56, 102, 56, 52, 98, 52, 52, 52, 34, 125, 44, 34, 97, 110, 110, 111, 116, 97, 116, 105, 111, 110, 115, 34, 58, 123, 34, 115, 105, 100, 101, 99, 97, 114, 46, 118, 97, 117, 108, 116, 46, 116, 97, 108, 101, 110, 100, 46, 111, 114, 103, 47, 105, 110, 106, 101, 99, 116, 34, 58, 34, 116, 114, 117, 101, 34, 125, 44, 34, 111, 119, 110, 101, 114, 82, 101, 102, 101, 114, 101, 110, 99, 101, 115, 34, 58, 91, 123, 34, 97, 112, 105, 86, 101, 114, 115, 105, 111, 110, 34, 58, 34, 97, 112, 112, 115, 47, 118, 49, 34, 44, 34, 107, 105, 110, 100, 34, 58, 34, 82, 101, 112, 108, 105, 99, 97, 83, 101, 116, 34, 44, 34, 110, 97, 109, 101, 34, 58, 34, 116, 101, 115, 116, 45, 97, 112, 112, 45, 49, 45, 55, 100, 56, 102, 56, 52, 98, 52, 52, 52, 34, 44, 34, 117, 105, 100, 34, 58, 34, 100, 54, 100, 50, 97, 55, 101, 100, 45, 54, 97, 56, 52, 45, 49, 49, 101, 57, 45, 98, 102, 99, 99, 45, 101, 56, 98, 57, 102, 56, 48, 99, 52, 54, 98, 54, 34, 44, 34, 99, 111, 110, 116, 114, 111, 108, 108, 101, 114, 34, 58, 116, 114, 117, 101, 44, 34, 98, 108, 111, 99, 107, 79, 119, 110, 101, 114, 68, 101, 108, 101, 116, 105, 111, 110, 34, 58, 116, 114, 117, 101, 125, 93, 125, 44, 34, 115, 112, 101, 99, 34, 58, 123, 34, 118, 111, 108, 117, 109, 101, 115, 34, 58, 91, 123, 34, 110, 97, 109, 101, 34, 58, 34, 115, 101, 99, 114, 101, 116, 115, 34, 44, 34, 101, 109, 112, 116, 121, 68, 105, 114, 34, 58, 123, 34, 109, 101, 100, 105, 117, 109, 34, 58, 34, 77, 101, 109, 111, 114, 121, 34, 125, 125, 44, 123, 34, 110, 97, 109, 101, 34, 58, 34, 97, 112, 112, 45, 49, 45, 115, 97, 45, 116, 111, 107, 101, 110, 45, 109, 55, 118, 102, 52, 34, 44, 34, 115, 101, 99, 114, 101, 116, 34, 58, 123, 34, 115, 101, 99, 114, 101, 116, 78, 97, 109, 101, 34, 58, 34, 97, 112, 112, 45, 49, 45, 115, 97, 45, 116, 111, 107, 101, 110, 45, 109, 55, 118, 102, 52, 34, 125, 125, 93, 44, 34, 99, 111, 110, 116, 97, 105, 110, 101, 114, 115, 34, 58, 91, 123, 34, 110, 97, 109, 101, 34, 58, 34, 116, 101, 115, 116, 45, 97, 112, 112, 45, 49, 45, 99, 111, 110, 116, 97, 105, 110, 101, 114, 34, 44, 34, 105, 109, 97, 103, 101, 34, 58, 34, 98, 117, 115, 121, 98, 111, 120, 58, 49, 46, 50, 56, 34, 44, 34, 99, 111, 109, 109, 97, 110, 100, 34, 58, 91, 34, 115, 104, 34, 44, 34, 45, 99, 34, 44, 34, 119, 104, 105, 108, 101, 32, 116, 114, 117, 101, 59, 100, 111, 32, 101, 99, 104, 111, 32, 92, 34, 77, 121, 32, 115, 101, 99, 114, 101, 116, 115, 32, 97, 114, 101, 58, 32, 36, 40, 99, 97, 116, 32, 47, 111, 112, 116, 47, 116, 97, 108, 101, 110, 100, 47, 115, 101, 99, 114, 101, 116, 115, 47, 115, 101, 99, 114, 101, 116, 115, 46, 112, 114, 111, 112, 101, 114, 116, 105, 101, 115, 41, 92, 34, 59, 32, 115, 108, 101, 101, 112, 32, 53, 59, 32, 100, 111, 110, 101, 92, 110, 34, 93, 44, 34, 114, 101, 115, 111, 117, 114, 99, 101, 115, 34, 58, 123, 125, 44, 34, 118, 111, 108, 117, 109, 101, 77, 111, 117, 110, 116, 115, 34, 58, 91, 123, 34, 110, 97, 109, 101, 34, 58, 34, 115, 101, 99, 114, 101, 116, 115, 34, 44, 34, 109, 111, 117, 110, 116, 80, 97, 116, 104, 34, 58, 34, 47, 111, 112, 116, 47, 116, 97, 108, 101, 110, 100, 47, 115, 101, 99, 114, 101, 116, 115, 34, 125, 44, 123, 34, 110, 97, 109, 101, 34, 58, 34, 97, 112, 112, 45, 49, 45, 115, 97, 45, 116, 111, 107, 101, 110, 45, 109, 55, 118, 102, 52, 34, 44, 34, 114, 101, 97, 100, 79, 110, 108, 121, 34, 58, 116, 114, 117, 101, 44, 34, 109, 111, 117, 110, 116, 80, 97, 116, 104, 34, 58, 34, 47, 118, 97, 114, 47, 114, 117, 110, 47, 115, 101, 99, 114, 101, 116, 115, 47, 107, 117, 98, 101, 114, 110, 101, 116, 101, 115, 46, 105, 111, 47, 115, 101, 114, 118, 105, 99, 101, 97, 99, 99, 111, 117, 110, 116, 34, 125, 93, 44, 34, 116, 101, 114, 109, 105, 110, 97, 116, 105, 111, 110, 77, 101, 115, 115, 97, 103, 101, 80, 97, 116, 104, 34, 58, 34, 47, 100, 101, 118, 47, 116, 101, 114, 109, 105, 110, 97, 116, 105, 111, 110, 45, 108, 111, 103, 34, 44, 34, 116, 101, 114, 109, 105, 110, 97, 116, 105, 111, 110, 77, 101, 115, 115, 97, 103, 101, 80, 111, 108, 105, 99, 121, 34, 58, 34, 70, 105, 108, 101, 34, 44, 34, 105, 109, 97, 103, 101, 80, 117, 108, 108, 80, 111, 108, 105, 99, 121, 34, 58, 34, 73, 102, 78, 111, 116, 80, 114, 101, 115, 101, 110, 116, 34, 125, 93, 44, 34, 114, 101, 115, 116, 97, 114, 116, 80, 111, 108, 105, 99, 121, 34, 58, 34, 65, 108, 119, 97, 121, 115, 34, 44, 34, 116, 101, 114, 109, 105, 110, 97, 116, 105, 111, 110, 71, 114, 97, 99, 101, 80, 101, 114, 105, 111, 100, 83, 101, 99, 111, 110, 100, 115, 34, 58, 51, 48, 44, 34, 100, 110, 115, 80, 111, 108, 105, 99, 121, 34, 58, 34, 67, 108, 117, 115, 116, 101, 114, 70, 105, 114, 115, 116, 34, 44, 34, 115, 101, 114, 118, 105, 99, 101, 65, 99, 99, 111, 117, 110, 116, 78, 97, 109, 101, 34, 58, 34, 97, 112, 112, 45, 49, 45, 115, 97, 34, 44, 34, 115, 101, 114, 118, 105, 99, 101, 65, 99, 99, 111, 117, 110, 116, 34, 58, 34, 97, 112, 112, 45, 49, 45, 115, 97, 34, 44, 34, 115, 101, 99, 117, 114, 105, 116, 121, 67, 111, 110, 116, 101, 120, 116, 34, 58, 123, 125, 44, 34, 115, 99, 104, 101, 100, 117, 108, 101, 114, 78, 97, 109, 101, 34, 58, 34, 100, 101, 102, 97, 117, 108, 116, 45, 115, 99, 104, 101, 100, 117, 108, 101, 114, 34, 44, 34, 116, 111, 108, 101, 114, 97, 116, 105, 111, 110, 115, 34, 58, 91, 123, 34, 107, 101, 121, 34, 58, 34, 110, 111, 100, 101, 46, 107, 117, 98, 101, 114, 110, 101, 116, 101, 115, 46, 105, 111, 47, 110, 111, 116, 45, 114, 101, 97, 100, 121, 34, 44, 34, 111, 112, 101, 114, 97, 116, 111, 114, 34, 58, 34, 69, 120, 105, 115, 116, 115, 34, 44, 34, 101, 102, 102, 101, 99, 116, 34, 58, 34, 78, 111, 69, 120, 101, 99, 117, 116, 101, 34, 44, 34, 116, 111, 108, 101, 114, 97, 116, 105, 111, 110, 83, 101, 99, 111, 110, 100, 115, 34, 58, 51, 48, 48, 125, 44, 123, 34, 107, 101, 121, 34, 58, 34, 110, 111, 100, 101, 46, 107, 117, 98, 101, 114, 110, 101, 116, 101, 115, 46, 105, 111, 47, 117, 110, 114, 101, 97, 99, 104, 97, 98, 108, 101, 34, 44, 34, 111, 112, 101, 114, 97, 116, 111, 114, 34, 58, 34, 69, 120, 105, 115, 116, 115, 34, 44, 34, 101, 102, 102, 101, 99, 116, 34, 58, 34, 78, 111, 69, 120, 101, 99, 117, 116, 101, 34, 44, 34, 116, 111, 108, 101, 114, 97, 116, 105, 111, 110, 83, 101, 99, 111, 110, 100, 115, 34, 58, 51, 48, 48, 125, 93, 44, 34, 112, 114, 105, 111, 114, 105, 116, 121, 34, 58, 48, 44, 34, 101, 110, 97, 98, 108, 101, 83, 101, 114, 118, 105, 99, 101, 76, 105, 110, 107, 115, 34, 58, 116, 114, 117, 101, 125, 44, 34, 115, 116, 97, 116, 117, 115, 34, 58, 123, 125, 125},
						},
					},
				},
			},
			expectedMutate{
				[]patchOperation{
					{
						Op: jsonPatchOpAdd, Path: jsonPathContainers + "/0",
						Value: map[string]interface{}{
							"name":  "tvsi-vault-agent",
							"image": "vault:1.3.0",
							"command": []interface{}{
								"sh",
								"-c",
								"cat \u003c\u003cEOF \u003e vault-agent-config.hcl\npid_file = \"/home/vault/pidfile\"\n\nauto_auth {\n  method \"kubernetes\" {\n    mount_path = \"auth/kubernetes\"\n    config = {\n      role = \"test-app-1\"\n      token_path = \"/var/run/secrets/talend/vault-sidecar-injector/serviceaccount/token\"\n    }\n  }\n\n  sink \"file\" {\n    config = {\n      path = \"/home/vault/.vault-token\"\n    }\n  }\n}\n\ntemplate {\n    destination = \"/opt/talend/secrets/secrets.properties\"\n    contents = <<EOH\n    {{ with secret \"secret/test-app-1/test-app-1-svc\" }}{{ range \\$k, \\$v := .Data }}\n{{ \\$k }}={{ \\$v }}\n{{ end }}{{ end }}\n    EOH\n    command = \"\"\n    wait {\n    min = \"1s\"\n    max = \"2s\"\n    }\n}\n\nEOF\nworkload_is_job=\"false\"\nif [ $workload_is_job = \"true\" ]; then\n  docker-entrypoint.sh agent -config=vault-agent-config.hcl -log-level=info &\n  while true; do\n    if [ -f \"/home/vault/vault-sidecars-signal-terminate\" ]; then\n      echo \"=> exit (signal received)\"\n      export VAULT_TOKEN=$(cat /home/vault/.vault-token);\n      vault token revoke -self;\n      exit 0\n    fi\n    sleep 2\n  done\nelse\n  docker-entrypoint.sh agent -config=vault-agent-config.hcl -log-level=info\nfi\n",
							},
							"env": []interface{}{
								map[string]interface{}{"name": "SKIP_SETCAP", "value": "true"},
								map[string]interface{}{"name": "VAULT_ADDR", "value": "https://vault:8200"},
							},
							"resources": map[string]interface{}{},
							"volumeMounts": []interface{}{
								map[string]interface{}{"name": "vault-token", "mountPath": "/home/vault"},
								map[string]interface{}{"name": "secrets", "mountPath": "/opt/talend/secrets"},
								map[string]interface{}{"name": "app-1-sa-token-m7vf4", "readOnly": true, "mountPath": "/var/run/secrets/talend/vault-sidecar-injector/serviceaccount"},
							},
							"lifecycle": map[string]interface{}{
								"preStop": map[string]interface{}{
									"exec": map[string]interface{}{
										"command": []interface{}{
											"sh",
											"-c",
											"export VAULT_TOKEN=$(cat /home/vault/.vault-token); vault token revoke -self;\n",
										},
									},
								},
							},
							"imagePullPolicy": "IfNotPresent",
						},
					},
					{
						Op: jsonPatchOpAdd, Path: jsonPathVolumes + "/-",
						Value: map[string]interface{}{"emptyDir": map[string]interface{}{"medium": "Memory"}, "name": "vault-token"},
					},
					{
						Op: jsonPatchOpAdd, Path: jsonPathAnnotations,
						Value: map[string]interface{}{vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationStatusKey]: vaultInjectorAnnotationStatusValue},
					},
				},
			},
		},
	}

	for _, table := range tables {
		// Mutate pod
		resp := vaultInjector.mutate(table.ar)

		assert.Equal(t, true, resp.Allowed)
		assert.Nil(t, resp.Result)
		assert.NotNil(t, resp.Patch)

		var patch []patchOperation
		if err = yaml.Unmarshal(resp.Patch, &patch); err != nil {
			t.Errorf("JSON Patch unmarshal error \"%s\"", err)
		}

		assert.Equal(t, table.patchExpected, patch)
	}
}
