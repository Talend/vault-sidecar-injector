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

package config

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

const (
	proxyCfgFileResolved    = "cache {\n    use_auto_auth_token = true\n}\n\nlistener \"tcp\" {\n    address = \"127.0.0.1:<APPSVC_PROXY_PORT>\"\n    tls_disable = true\n}"
	templateBlockResolved   = "template {\n    destination = \"/opt/talend/secrets/<APPSVC_SECRETS_DESTINATION>\"\n    contents = <<EOH\n    <APPSVC_TEMPLATE_CONTENT>\n    EOH\n    command = \"<APPSVC_TEMPLATE_COMMAND_TO_RUN>\"\n    wait {\n    min = \"1s\"\n    max = \"2s\"\n    }\n}"
	templateDefaultResolved = "{{ with secret \"<APPSVC_VAULT_SECRETS_PATH>\" }}{{ range \\$k, \\$v := .Data }}\n{{ \\$k }}={{ \\$v }}\n{{ end }}{{ end }}"
)

type inputLoaded struct {
	sidecarCfgFile        string
	proxyCfgFile          string
	templateBlockFile     string
	templateDefaultFile   string
	podLifecycleHooksFile string
}

type expectedLoad struct {
	sidecarCfgFileResolved        string
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
				"../../test/config/sidecarconfig.yaml",
				"../../test/config/proxyconfig.hcl",
				"../../test/config/tmplblock.hcl",
				"../../test/config/tmpldefault.tmpl",
				"../../test/config/podlifecyclehooks.yaml",
			},
			expectedLoad{
				"../../test/config/sidecarconfig.yaml.resolved",
				proxyCfgFileResolved,
				templateBlockResolved,
				templateDefaultResolved,
				"../../test/config/podlifecyclehooks.yaml.resolved",
			},
		},
	}

	for _, table := range tables {
		injectionCfg, err := Load(
			WhSvrParameters{
				0, 0, "", "",
				"", "", "",
				table.sidecarCfgFile,
				table.proxyCfgFile,
				table.templateBlockFile,
				table.templateDefaultFile,
				table.podLifecycleHooksFile,
			},
		)
		if err != nil {
			t.Errorf("Loading error \"%s\"", err)
		}

		// Verify strings
		assert.Equal(t, table.proxyCfgFileResolved, injectionCfg.ProxyConfig)
		assert.Equal(t, table.templateBlockResolved, injectionCfg.TemplateBlock)
		assert.Equal(t, table.templateDefaultResolved, injectionCfg.TemplateDefaultTmpl)

		// Verify yaml by marshalling the object into yaml again
		assert.Equal(t, stringFromYamlFile(t, table.sidecarCfgFileResolved), stringFromYamlObj(t, injectionCfg.SidecarConfig))
		assert.Equal(t, stringFromYamlFile(t, table.podLifecycleHooksFileResolved), stringFromYamlObj(t, injectionCfg.PodslifecycleHooks))
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
