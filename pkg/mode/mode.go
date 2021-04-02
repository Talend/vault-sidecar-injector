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

package mode

import (
	"os"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog"
)

// RegisterMode : register mode
func RegisterMode(modeInfo VaultInjectorModeInfo) {
	//klog.Infof("Registering mode: %s", modeInfo.Key)

	VaultInjectorModes[modeInfo.Key] = modeInfo

	if modeInfo.InjectContainerFunc == nil {
		klog.Error("Mandatory InjectContainer function not implemented")
		os.Exit(1)
	}
}

// GetModesStatus : get modes' status
func GetModesStatus(requestedModes []string, modes map[string]bool) {
	var defaultModeKey string

	// Init modes for current injection context
	for key := range VaultInjectorModes {
		modes[key] = false

		if VaultInjectorModes[key].DefaultMode {
			defaultModeKey = key
		}
	}

	requestedModesNum := len(requestedModes)

	if requestedModesNum > 0 {
		if requestedModesNum == 1 && requestedModes[0] == "" { // If no mode(s) provided then only enable default mode
			modes[defaultModeKey] = true
		} else {
			// Look at requested modes, ignore and remove unknown values
			for _, requestedMode := range requestedModes {
				bModeFound := false

				for key := range VaultInjectorModes {
					if requestedMode == key {
						modes[requestedMode] = true
						bModeFound = true

						// Look if we must also enable default mode when this mode is on (**and no other mode(s) requested**)
						if requestedModesNum == 1 && VaultInjectorModes[key].EnableDefaultMode {
							modes[defaultModeKey] = true
						}

						break
					}
				}

				if !bModeFound {
					klog.Warningf("Ignore unknown requested Vault Sidecar Injector mode: %s", requestedMode)
				}
			}
		}
	} else { // If no mode(s) provided then only enable default mode
		modes[defaultModeKey] = true
	}
}

// Use assert.ElementsMatch for comparing slices, but with a bool result.
type dummyt struct{}

func (t dummyt) Errorf(string, ...interface{}) {}

func IsEnabledModes(modesStatus map[string]bool, modesToCheck []string) bool {
	var enabledModes []string

	for mode, enabled := range modesStatus {
		if enabled {
			enabledModes = append(enabledModes, mode)
		}
	}

	// Allow to check for equality on slices without order. See https://stackoverflow.com/a/66062073
	return assert.ElementsMatch(dummyt{}, enabledModes, modesToCheck)
}
