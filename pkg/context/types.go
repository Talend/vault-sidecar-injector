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

package context

// InjectionContext : struct to carry computed placeholders' values and context info for current injection
type InjectionContext struct {
	K8sDefaultSATokenVolumeName    string
	VaultImage                     string
	VaultInjectorSATokenVolumeName string
	VaultAuthMethod                string
	VaultRole                      string
	ModesStatus                    map[string]bool
	ModesConfig                    map[string]ModeConfig
}

// ModeConfig : interface for mode's config
type ModeConfig interface {
	GetTemplate() string
}

// PatchOperation : this struct represents a JSON Patch operation (see http://jsonpatch.com/)
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
