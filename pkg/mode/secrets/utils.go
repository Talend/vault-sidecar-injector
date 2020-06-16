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
	"errors"
	"fmt"
	ctx "talend/vault-sidecar-injector/pkg/context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog"
)

func getSecretsType(config ctx.ModeConfig) (string, error) {
	if config != nil {
		secModeCfg, ok := config.(*secretsModeConfig) // here we use type assertion (https://golang.org/ref/spec#Type_assertions)

		if ok {
			return secModeCfg.secretsType, nil
		}

		err := errors.New("Provided type cannot be casted to 'secretsModeConfig'")
		klog.Error(err.Error())
		return "", err
	}

	err := errors.New("Null mode config")
	klog.Error(err.Error())
	return "", err
}

func isSecretsStatic(context *ctx.InjectionContext) bool {
	if secretsType, err := getSecretsType(context.ModesConfig[VaultInjectorModeSecrets]); err == nil {
		return secretsType == vaultInjectorSecretsTypeStatic
	}

	return false
}

func getMountPathOfSecretsVolume(cnts []corev1.Container) (string, error) {
	var secretsVolMountPath string

Loop:
	for _, sourceContainer := range cnts {
		for _, volMount := range sourceContainer.VolumeMounts {
			if volMount.Name == secretsVolName {
				secretsVolMountPath = volMount.MountPath
				break Loop
			}
		}
	}

	if secretsVolMountPath == "" {
		err := fmt.Errorf("Volume Mount %s not found in submitted pod", secretsVolName)
		klog.Error(err.Error())
		return "", err
	}

	return secretsVolMountPath, nil
}
