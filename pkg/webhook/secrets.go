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
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog"
)

// Vault Sidecar Injector: Secrets Mode
func (vaultInjector *VaultInjector) secretsMode(labels, annotations map[string]string) (string, error) {
	secretsPath := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsPathKey]], ",")
	secretsTemplate := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationSecretsTemplateKey]], "---")
	templateDest := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateDestKey]], ",")
	templateCmd := strings.Split(annotations[vaultInjector.VaultInjectorAnnotationsFQ[vaultInjectorAnnotationTemplateCmdKey]], ",")

	secretsPathNum := len(secretsPath)
	secretsTemplateNum := len(secretsTemplate)
	templateDestNum := len(templateDest)

	if secretsPathNum == 1 && secretsPath[0] == "" { // Build default secrets path: "secret/<application label>/<service label>"
		applicationLabel := labels[vaultInjector.ApplicationLabelKey]
		applicationServiceLabel := labels[vaultInjector.ApplicationServiceLabelKey]

		if applicationLabel == "" || applicationServiceLabel == "" {
			klog.Errorf("Submitted pod must contain labels %s and %s", vaultInjector.ApplicationLabelKey, vaultInjector.ApplicationServiceLabelKey)
			return "", fmt.Errorf("Submitted pod must contain labels %s and %s", vaultInjector.ApplicationLabelKey, vaultInjector.ApplicationServiceLabelKey)
		}

		secretsPath[0] = vaultDefaultSecretsEnginePath + "/" + applicationLabel + "/" + applicationServiceLabel
	}

	if templateDestNum == 1 && templateDest[0] == "" { // Use default
		templateDest[0] = templateAppSvcDefaultDestination
	}

	if secretsTemplateNum == 1 && secretsTemplate[0] == "" {
		// We must have same numbers of secrets path & secrets destinations
		if templateDestNum != secretsPathNum {
			klog.Error("Submitted pod must contain same numbers of secrets path and secrets destinations")
			return "", errors.New("Submitted pod must contain same numbers of secrets path and secrets destinations")
		}

		// If no custom template(s), use default template
		secretsTemplate = make([]string, templateDestNum)
		for tmplIdx := 0; tmplIdx < templateDestNum; tmplIdx++ {
			secretsTemplate[tmplIdx] = vaultInjector.TemplateDefaultTmpl
		}
	} else {
		// We must have same numbers of custom templates & secrets destinations ...
		if templateDestNum != secretsTemplateNum {
			klog.Error("Submitted pod must contain same numbers of templates and secrets destinations")
			return "", errors.New("Submitted pod must contain same numbers of templates and secrets destinations")
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
		templateBlock = vaultInjector.TemplateBlock
		templateBlock = strings.Replace(templateBlock, templateAppSvcDestinationPlaceholder, templateDest[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, templateContentPlaceholder, secretsTemplate[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, vaultAppSvcSecretsPathPlaceholder, secretsPath[tmplIdx], -1)
		templateBlock = strings.Replace(templateBlock, templateCommandPlaceholder, templateCommands[tmplIdx], -1)

		templates.WriteString(templateBlock)
		templates.WriteString("\n")
	}

	return templates.String(), nil
}
