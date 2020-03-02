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

const (
	//--- Job handling - Temporary mechanism until KEP https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/sidecarcontainers.md is implemented (and we migrate on appropriate version of k8s)
	jobMonitoringContainerName = "tvsi-job-babysitter" // Name of our specific sidecar container to inject in submitted jobs
	jobListenerContainerName   = "tvsi-vault-agent"    // Name of the container listening for signal from job monitoring container

	//--- Job handling env vars
	jobContainerNameEnv = "VSI_JOB_CNT_NAME" // Env var for name of the app job's container
	jobWorkloadEnv      = "VSI_JOB_WORKLOAD" // Env var set to "true" if submitted workload is a k8s job
)

const (
	//--- Vault Sidecar Injector supported modes
	vaultInjectorModeJob = "job" // Enable handling of Kubernetes Job
)
