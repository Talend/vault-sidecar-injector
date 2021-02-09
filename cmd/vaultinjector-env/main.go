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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func init() {
	jsonLog, _ := strconv.ParseBool(os.Getenv("VSI_ENV_LOG_JSON"))
	if jsonLog {
		// Log as JSON instead of the default ASCII formatter
		log.SetFormatter(&log.JSONFormatter{})
	}

	logLevel, err := strconv.Atoi(os.Getenv("VSI_ENV_LOG_LEVEL"))
	if err != nil {
		// Only log the warning severity or above
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(log.Level(logLevel))
	}
}

// Program accepts following env vars:
//
// VSI_ENV_LOG_JSON				true/false (default)	Log as JSON
// VSI_ENV_LOG_LEVEL			0 to 6 (default: 3)		Log level (3 for warning, 6 to trace everything)
//
func main() {
	var entrypointCmd []string
	if len(os.Args) == 1 {
		// a 'command' attribute must be set on images in pod manifest. If not we cannot start the expected process
		log.Fatalln("no command explicitly provided in submitted manifest, cannot determine image entrypoint")
	} else {
		entrypointCmd = os.Args[1:]
	}

	binary, err := exec.LookPath(entrypointCmd[0])
	log.Infof("Process to execute=%s", binary)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Our program has been copied in same location as secrets files
	vsienvProcess, err := os.Executable()
	if err != nil {
		log.Fatalln(err.Error())
	}

	secretsFilesPath := filepath.Dir(vsienvProcess)
	secretsFiles, err := ioutil.ReadDir(secretsFilesPath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Get currently defined env vars
	env := os.Environ()

	// Enrich env vars with content of all secrets files
	for _, fsecrets := range secretsFiles {
		secretsFile := path.Join(secretsFilesPath, fsecrets.Name())

		// Do not consider our program
		if secretsFile == vsienvProcess {
			continue
		}

		log.Infof("Secrets file=%s", secretsFile)

		props, err := parsePropertiesFile(secretsFile)
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Add to env vars. We do not check for collisions: make sure to not have same keys in secrets files (and do not use existing env keys either)
		env = append(env, props...)
	}

	// Replace current process with original one, providing env vars (including new ones from fetched secrets)
	err = syscall.Exec(binary, entrypointCmd, env)
	if err != nil {
		log.Fatalln("failed to exec process", entrypointCmd, err.Error())
	}
}

func parsePropertiesFile(filename string) ([]string, error) {
	var props []string

	if len(filename) == 0 {
		return props, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}

				props = append(props, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return props, nil
}
