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

package job

import (
	m "talend/vault-sidecar-injector/pkg/mode"
)

func init() {
	// Register mode
	m.RegisterMode(
		m.VaultInjectorModeInfo{
			Key:                 vaultInjectorModeJob,
			DefaultMode:         false,
			EnableDefaultMode:   true, // Default mode will also be enabled if job is **the only mode on** (as it does not make sense to have only this mode)
			InjectContainerFunc: jobModeInject,
		},
	)
}

func (jobModeCfg *jobModeConfig) GetTemplate() string {
	return jobModeCfg.template
}
