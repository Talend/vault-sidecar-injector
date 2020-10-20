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

package config

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

const (
	proxyCfgFileResolved    = "cache {\n    use_auto_auth_token = true\n}\n\nlistener \"tcp\" {\n    address = \"127.0.0.1:<VSI_PROXY_PORT>\"\n    tls_disable = true\n}"
	templateBlockResolved   = "template {\n    destination = \"/opt/talend/secrets/<VSI_SECRETS_DESTINATION>\"\n    contents = <<EOH\n    <VSI_SECRETS_TEMPLATE_CONTENT>\n    EOH\n    command = \"<VSI_SECRETS_TEMPLATE_COMMAND_TO_RUN>\"\n    wait {\n    min = \"1s\"\n    max = \"2s\"\n    }\n}"
	templateDefaultResolved = "{{ with secret \"<VSI_SECRETS_VAULT_SECRETS_PATH>\" }}{{ range $k, $v := .Data }}\n{{ $k }}={{ $v }}\n{{ end }}{{ end }}"
)

type inputLoaded struct {
	injectionCfgFile      string
	proxyCfgFile          string
	templateBlockFile     string
	templateDefaultFile   string
	podLifecycleHooksFile string
}

type expectedLoad struct {
	injectionCfgFileResolved      string
	proxyCfgFileResolved          string
	templateBlockResolved         string
	templateDefaultResolved       string
	podLifecycleHooksFileResolved string
}

func TestLoadConfig(t *testing.T) {
	tables := []struct {
		inputLoaded
		expectedLoad
	}{
		{
			inputLoaded{
				"../../test/config/injectionconfig.yaml",
				"../../test/config/proxyconfig.hcl",
				"../../test/config/tmplblock.hcl",
				"../../test/config/tmpldefault.tmpl",
				"../../test/config/podlifecyclehooks.yaml",
			},
			expectedLoad{
				"../../test/config/injectionconfig.yaml.resolved",
				proxyCfgFileResolved,
				templateBlockResolved,
				templateDefaultResolved,
				"../../test/config/podlifecyclehooks.yaml.resolved",
			},
		},
	}

	for _, table := range tables {
		vsiCfg, err := Load(
			WhSvrParameters{
				Port: 0, MetricsPort: 0,
				CACertFile: "", CertFile: "", KeyFile: "",
				WebhookCfgName:      "",
				AnnotationKeyPrefix: "", AppLabelKey: "", AppServiceLabelKey: "",
				InjectionCfgFile:      table.injectionCfgFile,
				ProxyCfgFile:          table.proxyCfgFile,
				TemplateBlockFile:     table.templateBlockFile,
				TemplateDefaultFile:   table.templateDefaultFile,
				PodLifecycleHooksFile: table.podLifecycleHooksFile,
			},
		)
		if err != nil {
			t.Fatalf("Loading error \"%s\"", err)
		}

		// Verify strings
		assert.Equal(t, table.proxyCfgFileResolved, vsiCfg.ProxyConfig)
		assert.Equal(t, table.templateBlockResolved, vsiCfg.TemplateBlock)
		assert.Equal(t, table.templateDefaultResolved, vsiCfg.TemplateDefaultTmpl)

		// Verify yaml by marshalling the object into yaml again
		assert.Equal(t, stringFromYamlFile(t, table.injectionCfgFileResolved), stringFromYamlObj(t, vsiCfg.InjectionConfig))
		assert.Equal(t, stringFromYamlFile(t, table.podLifecycleHooksFileResolved), stringFromYamlObj(t, vsiCfg.PodslifecycleHooks))
	}
}

func stringFromYamlFile(t *testing.T, filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("ReadFile error \"%s\"", err)
	}

	return string(data)
}

func stringFromYamlObj(t *testing.T, o interface{}) string {
	data, err := yaml.Marshal(o)
	if err != nil {
		t.Errorf("Yaml marshal error \"%s\"", err)
	}

	return string(data)
}
