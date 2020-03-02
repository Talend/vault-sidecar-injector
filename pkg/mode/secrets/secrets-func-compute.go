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

package secrets

import (
	"fmt"
	"strings"
	cfg "talend/vault-sidecar-injector/pkg/config"
	ctx "talend/vault-sidecar-injector/pkg/context"

	"k8s.io/klog"
)

func secretsModeCompute(config *cfg.VSIConfig, labels, annotations map[string]string) (ctx.ModeConfig, error) {
	secretsType := annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsTypeKey]]
	secretsPath := strings.Split(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsPathKey]], ",")
	secretsTemplate := strings.Split(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsTemplateKey]], "---")
	templateDest := strings.Split(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateDestKey]], ",")
	templateCmd := strings.Split(annotations[config.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateCmdKey]], ",")

	secretsPathNum := len(secretsPath)
	secretsTemplateNum := len(secretsTemplate)
	templateDestNum := len(templateDest)

	if secretsType == "" {
		secretsType = vaultInjectorSecretsTypeDynamic // Dynamic secrets by default
	} else {
		secretsTypeSupported := false
		for _, supportedSecretType := range vaultInjectorSecretsTypes {
			if secretsType == supportedSecretType {
				secretsTypeSupported = true
				break
			}
		}

		if !secretsTypeSupported {
			err := fmt.Errorf("[%s] Submitted pod makes use of unsupported secrets type %s", vaultInjectorModeSecrets, secretsType)
			klog.Error(err.Error())
			return nil, err
		}
	}

	if secretsPathNum == 1 && secretsPath[0] == "" { // Build default secrets path: "secret/<application label>/<service label>"
		applicationLabel := labels[config.ApplicationLabelKey]
		applicationServiceLabel := labels[config.ApplicationServiceLabelKey]

		if applicationLabel == "" || applicationServiceLabel == "" {
			err := fmt.Errorf("[%s] Submitted pod must contain labels %s and %s", vaultInjectorModeSecrets, config.ApplicationLabelKey, config.ApplicationServiceLabelKey)
			klog.Error(err.Error())
			return nil, err
		}

		secretsPath[0] = vaultDefaultSecretsEnginePath + "/" + applicationLabel + "/" + applicationServiceLabel
	}

	if templateDestNum == 1 && templateDest[0] == "" { // Use default
		templateDest[0] = templateAppSvcDefaultDestination
	}

	if secretsTemplateNum == 1 && secretsTemplate[0] == "" {
		// We must have same numbers of secrets path & secrets destinations
		if templateDestNum != secretsPathNum {
			err := fmt.Errorf("[%s] Submitted pod must contain same numbers of secrets path and secrets destinations", vaultInjectorModeSecrets)
			klog.Error(err.Error())
			return nil, err
		}

		// If no custom template(s), use default template
		secretsTemplate = make([]string, templateDestNum)
		for tmplIdx := 0; tmplIdx < templateDestNum; tmplIdx++ {
			secretsTemplate[tmplIdx] = config.TemplateDefaultTmpl
		}
	} else {
		// We must have same numbers of custom templates & secrets destinations ...
		if templateDestNum != secretsTemplateNum {
			err := fmt.Errorf("[%s] Submitted pod must contain same numbers of templates and secrets destinations", vaultInjectorModeSecrets)
			klog.Error(err.Error())
			return nil, err
		}

		// ... and we ignore content of 'secrets-path' annotation ('cause we provide full template), but we need to init an empty array
		// to not end up with errors in the replace loop to come
		secretsPath = make([]string, templateDestNum)
	}

	// Copy provided template commands, if less commands than secrets destinations: remaining commands set to ""
	templateCommands := make([]string, templateDestNum)
	copy(templateCommands, templateCmd)

	var templateBlock string
	var templates strings.Builder

	for tmplIdx := 0; tmplIdx < templateDestNum; tmplIdx++ {
		templateBlock = config.TemplateBlock
		templateBlock = strings.Replace(templateBlock, secretsDestinationPlaceholder, templateDest[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, secretsTemplateContentPlaceholder, secretsTemplate[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, secretsVaultPathPlaceholder, secretsPath[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, secretsTemplateCommandPlaceholder, templateCommands[tmplIdx], -1)

		templates.WriteString(templateBlock)
		templates.WriteString("\n")
	}

	return &secretsModeConfig{secretsType, templates.String()}, nil
}
