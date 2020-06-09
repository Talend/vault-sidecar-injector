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

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	defaultSecretsFilesPath = "/opt/talend/secrets"
)

func init() {
	jsonLog, _ := strconv.ParseBool(os.Getenv("VSI_ENV_LOG_JSON"))
	if jsonLog {
		// Log as JSON instead of the default ASCII formatter
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Only log the warning severity or above
	log.SetLevel(log.WarnLevel)
}

func main() {
	var entrypointCmd []string
	if len(os.Args) == 1 {
		log.Fatalln("failed to determine entrypoint")
	} else {
		entrypointCmd = os.Args[1:]
	}

	binary, err := exec.LookPath(entrypointCmd[0])
	if err != nil {
		log.Fatalln(err.Error())
	}

	env := os.Environ()

	// Enrich env vars with content of all secrets files
	secretsFilesPath := os.Getenv("VSI_ENV_SECRETS_FILES_PATH")
	if secretsFilesPath == "" {
		secretsFilesPath = defaultSecretsFilesPath
	}

	secretsFiles, err := ioutil.ReadDir(secretsFilesPath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, fsecrets := range secretsFiles {
		props, err := parsePropertiesFile(path.Join(secretsFilesPath, fsecrets.Name()))
		if err != nil {
			log.Fatalln(err.Error())
		}

		// Add to env vars. We do not check for collisions: make sure to not have same keys in secrets files (and do not use existing env keys either)
		env = append(env, props...)
	}

	// Replace current process with original one, providing env vars (including new ones from fetched secrets)
	err = syscall.Exec(binary, entrypointCmd, env)
	if err != nil {
		log.Panicln("failed to exec process", entrypointCmd, err.Error())
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
