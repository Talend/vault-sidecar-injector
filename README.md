
# Vault Sidecar Injector

[![GitHub All Releases](https://img.shields.io/github/downloads/Talend/vault-sidecar-injector/total?style=for-the-badge)](https://github.com/Talend/vault-sidecar-injector/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/talend/vault-sidecar-injector?style=for-the-badge)](https://hub.docker.com/r/talend/vault-sidecar-injector)

- [Vault Sidecar Injector](#vault-sidecar-injector)
  - [Announcements](#announcements)
  - [Overview](#overview)
  - [How to invoke Vault Sidecar Injector](#how-to-invoke-vault-sidecar-injector)
    - [Modes](#modes)
    - [Requirements](#requirements)
    - [Annotations](#annotations)
    - [Secrets Mode](#secrets-mode)
      - [Default template](#default-template)
      - [Template's Syntax](#templates-syntax)
    - [Proxy Mode](#proxy-mode)
    - [Modes and Injection Config Overview](#modes-and-injection-config-overview)
    - [Examples](#examples)
      - [Using Vault Kubernetes Auth Method](#using-vault-kubernetes-auth-method)
        - [Secrets mode - Usage with a K8S Deployment workload](#secrets-mode---usage-with-a-k8s-deployment-workload)
        - [Secrets mode - Static secrets](#secrets-mode---static-secrets)
        - [Secrets mode - Usage with a K8S Job workload](#secrets-mode---usage-with-a-k8s-job-workload)
        - [Secrets and Proxy modes - K8S Job workload](#secrets-and-proxy-modes---k8s-job-workload)
        - [Secrets mode - Custom secrets path and notification command](#secrets-mode---custom-secrets-path-and-notification-command)
        - [Secrets mode - Ask for secrets hook injection, custom secrets file and template](#secrets-mode---ask-for-secrets-hook-injection-custom-secrets-file-and-template)
        - [Secrets mode - Ask for secrets hook injection, several custom secrets files and templates](#secrets-mode---ask-for-secrets-hook-injection-several-custom-secrets-files-and-templates)
      - [Using Vault AppRole Auth Method](#using-vault-approle-auth-method)
  - [How to deploy Vault Sidecar Injector](#how-to-deploy-vault-sidecar-injector)
    - [Prerequisites](#prerequisites)
    - [Vault Sidecar Injector image](#vault-sidecar-injector-image)
    - [Installing the Chart](#installing-the-chart)
    - [Uninstalling the chart](#uninstalling-the-chart)
  - [Configuration](#configuration)
  - [Metrics](#metrics)
  - [List of changes](#list-of-changes)

## Announcements

- 2020-03: [Vault Sidecar Injector vs HashiCorp Vault Agent Injector - Features Comparison](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/HashiCorp-Vault-Agent-Injector.md)
- 2020-03: [Static vs Dynamic secrets](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Static-vs-Dynamic-Secrets.md)
- 2019-12: [Discovering Vault Sidecar Injector's Proxy feature](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Discovering-Vault-Sidecar-Injector-Proxy.md)
- 2019-11: [Vault Sidecar Injector now leverages Vault Agent Template feature](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Leveraging-Vault-Agent-Template.md)
- 2019-10: [Open-sourcing Vault Sidecar Injector](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Open-sourcing-Vault-Sidecar-Injector.md)

## Overview

`Vault Sidecar Injector` consists in a **Webhook Admission Server**, registered in the Kubernetes Mutating Admission Webhook Controller, that will mutate resources depending on defined criteriae. See here for more details: <https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks>.

This component allows **to dynamically inject Vault Agent as either an init or a sidecar container** (along with configuration and volumes) in any matching pod manifest to seamlessly fetch secrets from Vault. Pods willing to benefit from this feature just have to add some custom annotations to ask for the injection **at deployment time**.

To ease deployment, a Helm chart is provided under [deploy/helm](https://github.com/Talend/vault-sidecar-injector/blob/master/deploy/helm) folder of this repository as well as instructions to [install it](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#installing-the-chart).

> ⚠️ **Important note** ⚠️: support for sidecars in Kubernetes **jobs** suffers from limitations and issues exposed here: <https://github.com/kubernetes/kubernetes/issues/25908>.
>
> A Kubernetes proposal tries to address those points: <https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/sidecarcontainers.md>, <https://github.com/kubernetes/enhancements/issues/753>. Implementation of the proposal has started and may be released in Kubernetes 1.19 (in Alpha stage).
>
> In the meantime however, `Vault Sidecar Injector` implements **specific sidecar and signaling mechanism** to properly stop all injected containers on job termination.

## How to invoke Vault Sidecar Injector

### Modes

`Vault Sidecar Injector` supports several high-level features or *modes*:

- [**secrets**](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#secrets-mode), the primary mode allowing to retrieve secrets from Vault server's stores, either once (for **static secrets**) or continuously (for **dynamic secrets**), coping with secrets rotations (ie any change will be propagated and updated values made available to consume by applications).
- [**proxy**](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#proxy-mode), to enable the injected Vault Agent as a local, authenticated gateway to the remote Vault server. As an example, with this mode on, applications can easily leverage Vault's Transit Engine to cipher/decipher payloads by just sending data to the local proxy without dealing themselves with Vault authentication and tokens.
- **job**, to use when a Kubernetes Job is submitted. This new mode comes in replacement of the now deprecated `sidecar.vault.talend.org/workload` annotation.

For details, refer to [Modes and Injection Config Overview](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#modes-and-injection-config-overview).

### Requirements

Invoking `Vault Sidecar Injector` is pretty straightforward. In your application manifest:

- Add annotation `sidecar.vault.talend.org/inject: "true"`. This is the only mandatory annotation (see list of supported annotations below).
- When using **secrets** mode:
  - Add volume `secrets`, setting field `emptyDir.medium` to *Memory*. Deciphered secrets will be made available in file `secrets.properties` (using format `<secret key>=<secret value>`) by default or in the secrets destination you provide with annotation `sidecar.vault.talend.org/secrets-destination`.
    > Note: as a fallback measure, if your manifest does not define a `secrets` volume, `Vault Sidecar Injector` will add one to the resulting pod.
  - For Kubernetes **Job workloads dealing with dynamic secrets only**:
    - **Use of `serviceAccountName` attribute**, with role allowing to perform GET on pods (needed to poll for job's pod status)
    - **Do not make use of annotation `sidecar.vault.talend.org/secrets-hook`** as it will immediately put the job in error state. This hook is meant to be used with regular workloads only (ie Kubernetes Deployments) as it forces a restart of the application container until secrets are available in application's context. With jobs, as we look after status of the job container, our special signaling mechanism will terminate all the sidecars upon job exit thus preventing use of the hook.

Refer to provided [sample files](https://github.com/Talend/vault-sidecar-injector/blob/master/samples) and [examples](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#examples) section.

### Annotations

Following annotations in requesting pods are supported:

| Annotation                            | (M)andatory / (O)ptional |  Apply to mode | Default Value        | Supported Values               | Description |
|---------------------------------------|--------------------------|-----------------|----------------------|--------------------------------|-------------|
| `sidecar.vault.talend.org/inject`     | M           |    N/A          |                      | "true" / "on" / "yes" / "y"  | Ask for sidecar injection to get secrets from Vault    |
| `sidecar.vault.talend.org/auth`       | O           |    N/A          | "kubernetes"   | "kubernetes" / "approle" | Vault Auth Method to use. **Static secrets only supports "kubernetes" authentication method** |
| `sidecar.vault.talend.org/mode`       | O           |    N/A          | "secrets"      | "secrets" / "proxy" / "job" / Comma-separated values (eg "secrets,proxy") | Enable provided mode(s). **Note: `secrets` mode will be enabled if you only set `job` mode**   |
| `sidecar.vault.talend.org/notify`     | O           |    secrets   | ""   | Comma-separated strings  | List of commands to notify application/service of secrets change, one per secrets path. **Usage context: dynamic secrets only** |
| `sidecar.vault.talend.org/proxy-port` | O           |    proxy        | "8200"    | Any allowed port value  | Port for local Vault proxy |
| `sidecar.vault.talend.org/role`       | O           |    N/A          | "\<`com.talend.application` label\>" | Any string    | **Only used with "kubernetes" Vault Auth Method**. Vault role associated to requesting pod. If annotation not used, role is read from label defined by `mutatingwebhook.annotations.appLabelKey` key (refer to [configuration](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#configuration)) which is `com.talend.application` by default |
| `sidecar.vault.talend.org/sa-token`   | O           |    N/A         | "/var/run/secrets/kubernetes.io/serviceaccount/token" | Any string | Full path to service account token used for Vault Kubernetes authentication |
| `sidecar.vault.talend.org/secrets-destination` | O     | secrets | "secrets.properties" | Comma-separated strings  | List of secrets filenames (without path), one per secrets path |
| `sidecar.vault.talend.org/secrets-hook`        | O     | secrets | *nil* | "true" / "on" / "yes" / "y" | If set, lifecycle hooks will be added to pod's container(s) to wait for secrets files. **Usage context: dynamic secrets only** |
| `sidecar.vault.talend.org/secrets-path`        | O     | secrets | "secret/<`com.talend.application` label>/<`com.talend.service` label>" | Comma-separated strings | List of secrets engines and path. If annotation not used, path is set from labels defined by `mutatingwebhook.annotations.appLabelKey`  and `mutatingwebhook.annotations.appServiceLabelKey` keys (refer to [configuration](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#configuration))      |
| `sidecar.vault.talend.org/secrets-template`    | O     | secrets  | [Default template](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#default-template) | templates separated with `---` | Allow to override default template. Ignore `sidecar.vault.talend.org/secrets-path` annotation if set |
| `sidecar.vault.talend.org/secrets-type` | O  | secrets | "dynamic" | "static" / "dynamic" | Type of secrets to handle (see details [here](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Static-vs-Dynamic-Secrets.md)) |
| `sidecar.vault.talend.org/workload`   | O      | N/A | *nil*  | "job" | Type of submitted workload. **⚠️ Deprecated: use `sidecar.vault.talend.org/mode` instead. Using this annotation will enable `job` mode ⚠️** |

Upon successful injection, Vault Sidecar Injector will add annotation(s) to the requesting pods:

| Annotation                        | Value      | Description                                 |
|-----------------------------------|------------|---------------------------------------------|
| `sidecar.vault.talend.org/status` | "injected" | Status set by Vault Sidecar Injector        |

> **Note:** you can change the annotation prefix (set by default to `sidecar.vault.talend.org`) thanks to `mutatingwebhook.annotations.keyPrefix` key in [configuration](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#configuration).

### Secrets Mode

#### Default template

Template below is used by default to fetch all secrets and create corresponding key/value pairs. It is generic enough and should be fine for most use cases:

<!-- {% raw %} -->
```yaml
{{ with secret "<Path to secrets (defaut value or `sidecar.vault.talend.org/secrets-path` annotation)>" }}{{ range $k, $v := .Data }}
{{ $k }}={{ $v }}
{{ end }}{{ end }}
```
<!-- {% endraw %}) -->

Using annotation `sidecar.vault.talend.org/secrets-template` it is nevertheless possible to provide your own list of templates. For some examples have a look at the next section ([here](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#secrets-mode---ask-for-secrets-hook-injection-custom-secrets-file-and-template) and [there](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#secrets-mode---ask-for-secrets-hook-injection-several-custom-secrets-files-and-templates)).

#### Template's Syntax

Details on template syntax can be found in Consul Template's documentation (same syntax is supported by Vault Agent Template):

- <https://github.com/hashicorp/consul-template#secret>
- <https://github.com/hashicorp/consul-template#secrets>
- <https://github.com/hashicorp/consul-template#helper-functions>

### Proxy Mode

This mode opens the gate to virtually any Vault features for requesting applications. A [blog entry](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Discovering-Vault-Sidecar-Injector-Proxy.md) introduces this mode and examples are provided.

### Modes and Injection Config Overview

Depending on the modes you decide to enable and whether you opt for static or dynamic secrets (when **secrets** mode is selected), the configuration injected into your pod varies. The following table provides a quick glance at the different configurations.

<table>
  <colgroup span="4"></colgroup>
  <colgroup span="2"></colgroup>
  <tr>
    <th colspan="4" scope="colgroup">Enabled Mode(s)</th>
    <th colspan="2" scope="colgroup">Injected Configuration</th>
  </tr>
  <tr>
    <th scope="col"><b>secrets</b> <i>(static)</i></th>
    <th scope="col"><b>secrets</b> <i>(dynamic)</i></th>
    <th scope="col"><b>proxy</b></th>
    <th scope="col"><b>job</b></th>
    <th scope="col">Init Container</th>
    <th scope="col">Sidecar(s)<i>²</i></th>
  </tr>
  <tr>
    <td align="center">X</td><td/><td/><td/><td align="center" bgcolor="grey">X</td><td bgcolor="grey"/>
  </tr>
  <tr>
    <td align="center">X</td><td/><td/><td align="center">X<b><i>¹</i></b></td><td align="center" bgcolor="grey">X (secrets)</td><td align="center" bgcolor="grey">X (job)</td>
  </tr>
  <tr>
    <td/><td align="center">X</td><td/><td/><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
  <tr>
    <td/><td align="center">X</td><td/><td align="center">X<b><i>¹</i></b></td><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
  <tr>
    <td/><td/><td align="center">X</td><td/><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
  <tr>
    <td/><td/><td align="center">X</td><td align="center">X</td><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
  <tr>
    <td align="center">X</td><td/><td align="center">X</td><td/><td align="center" bgcolor="grey">X (secrets)</td><td align="center" bgcolor="grey">X (proxy)</td>
  </tr>
  <tr>
    <td align="center">X</td><td/><td align="center">X</td><td align="center">X</td><td align="center" bgcolor="grey">X (secrets)</td><td align="center" bgcolor="grey">X (proxy & job)</td>
  </tr>
  <tr>
    <td/><td align="center">X</td><td align="center">X</td><td/><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
  <tr>
    <td/><td align="center">X</td><td align="center">X</td><td align="center">X</td><td bgcolor="grey"/><td align="center" bgcolor="grey">X</td>
  </tr>
</table>

> **[1]** *on job mode:* if you only set mode annotation's value to "job", `secrets` mode will be enabled automatically and configured to handle dynamic secrets (unless you set `sidecar.vault.talend.org/secrets-type` to "static" but note that in this situation, there is no need, although we do not prevent it, to enable job mode as no Vault Agent will be injected as sidecar).

> **[2]** *on number of injected sidecars:* for Kubernetes **Deployment** workloads, **only one sidecar container** is added to your pod to handle dynamic secrets and/or proxy. For Kubernetes **Job** workloads, **two sidecars** are injected to achieve the same tasks.

### Examples

**Ready to use sample manifests are provided under [samples](https://github.com/Talend/vault-sidecar-injector/blob/master/samples) folder**. Just deploy them using `kubectl apply -f <sample file>`.

Examples hereafter go further and highlight all the features of `Vault Sidecar Injector` through the supported annotations.

#### Using Vault Kubernetes Auth Method

##### Secrets mode - Usage with a K8S Deployment workload

<details>
<summary>
Show example
</summary>

Only mandatory annotation to ask for Vault Agent injection is `sidecar.vault.talend.org/inject`.

It means that, with the provided manifest below:

- Vault authentication done using role `test-app-1` (value of `com.talend.application` label)
- secrets fetched from Vault's path `secret/test-app-1/test-app-1-svc` using default template
- secrets to be stored into `/opt/talend/secrets/secrets.properties`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-1
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test-app-1
      com.talend.service: test-app-1-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
      labels:
        com.talend.application: test-app-1
        com.talend.service: test-app-1-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          ...
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

##### Secrets mode - Static secrets

<details>
<summary>
Show example
</summary>

With the provided manifest below:

- Vault authentication done using role `test-app-1` (value of `com.talend.application` label)
- secrets fetched from Vault's path `secret/test-app-1/test-app-1-svc` using default template
- secrets to be stored into `/opt/talend/secrets/secrets.properties`
- as we specified *static* for `sidecar.vault.talend.org/secrets-type`, the secrets **will not be refreshed** but fetched only once

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-1
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test-app-1
      com.talend.service: test-app-1-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/secrets-type: "static" # static secrets
      labels:
        com.talend.application: test-app-1
        com.talend.service: test-app-1-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          ...
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

##### Secrets mode - Usage with a K8S Job workload

<details>
<summary>
Show example
</summary>

When submitting a job, use annotation `sidecar.vault.talend.org/mode` and set value to `job`.

The service account used to run the job should at least have the following permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: job-sa
  namespace: default
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: job-pod-status
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: job-pod-status
subjects:
  - kind: ServiceAccount
    name: job-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: job-pod-status
```

Last point is to make sure your job is waiting for availability of secrets file(s) before starting as it can take a couple of seconds before secrets are fetched from the Vault server. A simple polling loop is provided in the sample below.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: example-2
  namespace: default
spec:
  backoffLimit: 1
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/mode: "job"
      labels:
        com.talend.application: test
        com.talend.service: test-app-svc
    spec:
      restartPolicy: Never
      serviceAccountName: job-sa
      containers:
        - name: test-app-1-job-container
          image: busybox:1.28
          command:
            - "sh"
            - "-c"
            - |
              while true; do
                echo "Wait for secrets file before running job..."
                if [ -f "/opt/talend/secrets/secrets.properties" ]; then
                  echo "Secrets available"
                  break
                fi
                sleep 2
              done
              echo "Job started"
              (...)
              exit_code=$?
              echo "Job stopped"
              exit $exit_code
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

##### Secrets and Proxy modes - K8S Job workload

<details>
<summary>
Show example
</summary>

This sample demonstrates how to enable the proxy mode in addition to the secrets mode used in previous samples. We are here again using a Kubernetes job so we reuse the dedicated service account.

Key annotation to use is `sidecar.vault.talend.org/mode` to let `Vault Sidecar Injector` knows that proxy mode must be enabled. Optional `sidecar.vault.talend.org/proxy-port` annotation can be handy if default proxy port has to be customized.

Once enabled, your application container can directly interact with the Vault server by sending requests to the injected Vault Agent sidecar that now also acts as a proxy handling authentication with the server. The proxy is available at `http://127.0.0.1:<proxy port>`.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: job-sa
  namespace: default
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: job-pod-status
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: job-pod-status
subjects:
  - kind: ServiceAccount
    name: job-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: job-pod-status
---
apiVersion: batch/v1
kind: Job
metadata:
  name: example-3
  namespace: default
spec:
  backoffLimit: 1
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/mode: "secrets,proxy,job"  # Enable 'secrets', 'proxy' and 'job' modes
        sidecar.vault.talend.org/proxy-port: "9999"     # Optional: override default proxy port value (8200)
      labels:
        com.talend.application: test
        com.talend.service: test-app-svc
    spec:
      restartPolicy: Never
      serviceAccountName: job-sa
      containers:
        - name: ...
          image: ...
          command:
            - "sh"
            - "-c"
            - |
              set -e
              while true; do
                echo "Wait for secrets file before running job..."
                if [ -f "/opt/talend/secrets/secrets.properties" ]; then
                  echo "Secrets available"
                  break
                fi
                sleep 2
              done
              echo "Job started"
              (...)

              # Vault server can be reached through the local Vault Agent sidecar acting as a proxy
              # You don't need to deal with authentication or the Vault token: Vault Agent proxy is in charge.
              curl [...] http://127.0.0.1:9999/<Vault API endpoint to use>

              (...)
              echo "Job stopped"
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

##### Secrets mode - Custom secrets path and notification command

<details>
<summary>
Show example
</summary>

Several optional annotations to end up with:

- Vault authentication using role `test-app-4` (value of `com.talend.application` label)
- secrets fetched from Vault's path `secret/test-app-4-svc` using default template
- secrets to be stored into `/opt/app/secrets.properties`
- secrets changes trigger notification command `curl localhost:8888/actuator/refresh -d {} -H 'Content-Type: application/json'`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-4
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test-app-4
      com.talend.service: test-app-4-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/secrets-path: "secret/test-app-4-svc"
        sidecar.vault.talend.org/notify: "curl localhost:8888/actuator/refresh -d {} -H 'Content-Type: application/json'"
      labels:
        com.talend.application: test-app-4
        com.talend.service: test-app-4-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          ...
          volumeMounts:
            - name: secrets
              mountPath: /opt/app
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

##### Secrets mode - Ask for secrets hook injection, custom secrets file and template

<details>
<summary>
Show example
</summary>

Several optional annotations to end up with:

- Vault authentication using role `test-app-6` (value of `com.talend.application` label)
- hook injected in application's container(s) to wait for secrets file availability
- secrets fetched from Vault's path `aws/creds/test-app-6` using **one custom template**
- secrets to be stored into `/opt/talend/secrets/creds.properties`

<!-- {% raw %} -->
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-5
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test-app-6
      com.talend.service: test-app-6-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/secrets-hook: "true"
        sidecar.vault.talend.org/secrets-destination: "creds.properties"
        sidecar.vault.talend.org/secrets-template: |
          {{ with secret "aws/creds/test-app-6" }}
          {{ if .Data.access_key }}
          test-app-6.s3.credentials.accessKey={{ .Data.access_key }}
          {{ end }}
          {{ if .Data.secret_key }}
          test-app-6.s3.credentials.secret={{ .Data.secret_key }}
          {{ end }}
          {{ end }}
      labels:
        com.talend.application: test-app-6
        com.talend.service: test-app-6-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          ...
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
<!-- {% endraw %}) -->
</details>

##### Secrets mode - Ask for secrets hook injection, several custom secrets files and templates

<details>
<summary>
Show example
</summary>

Several optional annotations to end up with:

- Vault authentication using role `test-app-2` (value of `com.talend.application` label)
- hook injected in application's container(s) to wait for secrets file availability
- secrets fetched from Vault's path `secret/test-app-2/test-app-2-svc` using **several custom templates** (use `---` as separation between them)
- secrets to be stored into `/opt/talend/secrets/secrets.properties` (using first template) and `/opt/talend/secrets/secrets2.properties` (using second template)

<!-- {% raw %} -->
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-6
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test-app-2
      com.talend.service: test-app-2-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/secrets-hook: "true"
        sidecar.vault.talend.org/secrets-destination: "secrets.properties,secrets2.properties"
        sidecar.vault.talend.org/secrets-template: |
          {{ with secret "secret/test-app-2/test-app-2-svc" }}
          {{ if .Data.SECRET1 }}
          my_custom_key_name1={{ .Data.SECRET1 }}
          {{ end }}
          {{ end }}
          ---
          {{ with secret "secret/test-app-2/test-app-2-svc" }}
          {{ if .Data.SECRET2 }}
          my_custom_key_name2={{ .Data.SECRET2 }}
          {{ end }}
          {{ end }}
      labels:
        com.talend.application: test-app-2
        com.talend.service: test-app-2-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          ...
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
<!-- {% endraw %}) -->
</details>

#### Using Vault AppRole Auth Method

> ⚠️ AppRole Auth Method can only be used with **dynamic** secrets.

<details>
<summary>
Show example
</summary>

- Ask Vault Sidecar Injector to use Vault authentication method `approle` (instead of the default which is `kubernetes`)
- Vault AppRole requires info stored into 2 files (containing role id and secret id) that have to be provided to Vault Sidecar Injector: in the sample below this is done via an init container (replace placeholders \<ROLE ID from Vault\> and \<SECRET ID from Vault\> by values generated by Vault server using commands `vault read auth/approle/role/test/role-id` and `vault write -f auth/approle/role/test/secret-id`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-7
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      com.talend.application: test
      com.talend.service: test-app-svc
  template:
    metadata:
      annotations:
        sidecar.vault.talend.org/inject: "true"
        sidecar.vault.talend.org/auth: "approle"
      labels:
        com.talend.application: test
        com.talend.service: test-app-svc
    spec:
      serviceAccountName: default
      initContainers:
        - name: test-app-8-container-init
          image: busybox:1.28
          command:
            - "sh"
            - "-c"
            - |
              echo "<ROLE ID from Vault>" > /opt/talend/secrets/approle_roleid
              echo "<SECRET ID from Vault>" > /opt/talend/secrets/approle_secretid
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      containers:
        - name: test-app-8-container
          image: busybox:1.28
          command:
            - "sh"
            - "-c"
            - >
              while true;do echo "My secrets are: $(cat /opt/talend/secrets/secrets.properties)"; sleep 5; done
          volumeMounts:
            - name: secrets
              mountPath: /opt/talend/secrets
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
</details>

## How to deploy Vault Sidecar Injector

The provided [chart](https://github.com/Talend/vault-sidecar-injector/blob/master/deploy/helm) is intended to be deployed in a "system" namespace and only once as it handles all injection requests from any pods deployed in any namespaces. **It *shall not* be deployed in every namespaces**.

>**Note**: it is possible to deploy an instance in a given namespace **and to restrict injection to this same namespace** if necessary, **in particular in a dev environment where each team wants its own instance of `Vault Sidecar Injector` for testing purpose** with its dedicated configuration (including a dedicated Vault server). Refer to section [Installing the chart in a dev environment](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#installing-the-chart-in-a-dev-environment) below.

### Prerequisites

Installation:

- Kubernetes v1.10+
- Helm 2 or 3

Runtime:

- Vault server deployed (either *in cluster* with official chart <https://github.com/hashicorp/vault-helm> or *out of cluster*), started and reachable through Kubernetes service & endpoint deployed into cluster

<details>
<summary>
<b>Helm 2: Tiller installation</b>
</summary>

> Note: this step does not apply if you are using Helm 3.

Install Tiller using a service account:

```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: tiller
  namespace: kube-system
EOF

$ helm init --service-account tiller
```

> On Kubernetes `1.16+`: make sure to use Helm **`2.16.0` or higher** as previous Helm 2 versions rely on deprecated APIs no longer served (see <https://kubernetes.io/blog/2019/09/18/kubernetes-1-16-release-announcement/>).

For details on using Tiller with RBAC:

- <https://v2.helm.sh/docs/using_helm/#tiller-and-user-permissions>
- <https://v2.helm.sh/docs/using_helm/#tiller-and-role-based-access-control>
</details>

<details>
<summary>
<b>Vault Server installation</b>
</summary>

> **Note:** this step is optional if you already have a running Vault server. This section helps you setup a test Vault server with ready to use configuration.

We will install a test Vault server in Kubernetes cluster but an external, out of cluster, Vault server can also be used. Note that we will install Vault server in *dev mode* below, do not use this setup in production.

Using HashiCorp's Vault Helm chart:

```bash
$ git clone https://github.com/hashicorp/vault-helm.git
$ cd vault-helm
$ git checkout v0.5.0
$ helm install . --name=vault --set injector.enabled=false --set server.dev.enabled=true --set ui.enabled=true --set ui.serviceType="NodePort"
```

Then init Vault server with our test config:

```bash
# Check status
$ kubectl exec -it vault-0 -- vault status
$ kubectl logs vault-0

# Set up needed auth methods, secrets engines, policies, roles and secrets
$ cd vault-sidecar-injector/deploy/vault
$ ./init-dev-vault-server.sh
```
</details>

### Vault Sidecar Injector image

> Note: if you don't intend to perform some tests with the image you can skip this section and jump to [Installing the Chart](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#installing-the-chart).

<details>
<summary>
<b>Pulling the image from Docker Hub</b>
</summary>

Official Docker images are published on [Talend's public Docker Hub](https://hub.docker.com/r/talend/vault-sidecar-injector) repository for each `Vault Sidecar Injector` release. Provided Helm chart will pull the image automatically if needed.  

For manual pull of a specific tag:

```bash
$ docker pull talend/vault-sidecar-injector:<tag>
```
</details>

<details>
<summary>
<b>Building the image</b>
</summary>

A [Dockerfile](https://github.com/Talend/vault-sidecar-injector/blob/master/Dockerfile) is also provided to both compile `Vault Sidecar Injector` and build the image locally if you prefer.

Just run following command:

```bash
$ make image
```

> Note: if you have Go installed on your machine, you can use `make image-from-build` instead.

</details>

### Installing the Chart

Several options to install the chart:

- from [Helm Hub](https://hub.helm.sh/charts/talend/vault-sidecar-injector) leveraging [Talend's public Helm charts registry](https://talend.github.io/helm-charts-public)
- by downloading the chart archive (`.tgz` file) from GitHub [releases](https://github.com/Talend/vault-sidecar-injector/releases)
- or cloning `Vault Sidecar Injector` GitHub repo and cd into `deploy/helm` directory

Depending on what you chose, define a `CHART_LOCATION` env var as follows:

- if you use [Helm Hub](https://hub.helm.sh/charts/talend/vault-sidecar-injector) / [Talend's public Helm charts registry](https://talend.github.io/helm-charts-public):

```bash
$ helm repo add talend https://talend.github.io/helm-charts-public/stable
$ helm repo update
$ export CHART_LOCATION=talend/vault-sidecar-injector
```

- if you use the downloaded chart archive:

```bash
$ export CHART_LOCATION=./vault-sidecar-injector-<x.y.z>.tgz
```

- if you install from the chart's folder:

```bash
$ export CHART_LOCATION=.
```

To see Chart content before installing it, perform a dry run first:

```bash
$ cd deploy/helm

# If using Helm 2
$ helm install $CHART_LOCATION --name vault-sidecar-injector --namespace <namespace for deployment> --set vault.addr=<Vault server address> --debug --dry-run

# If using Helm 3
$ helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address> --debug --dry-run
```

To install the chart on the cluster:

```bash
$ cd deploy/helm

# If using Helm 2
$ helm install $CHART_LOCATION --name vault-sidecar-injector --namespace <namespace for deployment> --set vault.addr=<Vault server address>

# If using Helm 3
$ helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address>
```

> **Note:** `Vault Sidecar Injector` should be deployed only once (except for testing purpose, see below). It will mutate any "vault-sidecar annotated" pod from any namespace.

As an example, to install `Vault Sidecar Injector` on our test cluster:

```bash
$ cd deploy/helm

# If using Helm 2
$ helm install $CHART_LOCATION --name vault-sidecar-injector --namespace kube-system --set vault.addr=http://vault:8200 --set vault.ssl.verify=false

# If using Helm 3
$ helm install vault-sidecar-injector $CHART_LOCATION --namespace kube-system --set vault.addr=http://vault:8200 --set vault.ssl.verify=false
```

This command deploys the component on the Kubernetes cluster with modified configuration to target our Vault server in-cluster test instance (no verification of certificates): such settings *are no fit for production*.

The [configuration](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#configuration) section lists all the parameters that can be configured during installation.

<details>
<summary>
<b>Installing the chart in a dev environment</b>
</summary>

In a dev environment, you may want to install your own test instance of `Vault Sidecar Injector`, connected to your own Vault server and limiting injection to a given namespace. To do so, use following options:

```bash
$ cd deploy/helm

# If using Helm 2
$ helm install $CHART_LOCATION --name vault-sidecar-injector --namespace <your dev namespace> --set vault.addr=<your dev Vault server address> --set mutatingwebhook.namespaceSelector.namespaced=true

# If using Helm 3
$ helm install vault-sidecar-injector $CHART_LOCATION --namespace <your dev namespace> --set vault.addr=<your dev Vault server address> --set mutatingwebhook.namespaceSelector.namespaced=true
```

And then **add a label on your namespace** as follows (if not done, no injection will be performed):

```bash
$ kubectl label namespace <your dev namespace> vault-injection=<your dev namespace> --overwrite

# check label on namespace
$ kubectl get namespace -L vault-injection
```
</details>

<details>
<summary>
<b>Restrict injection to specific namespaces</b>
</summary>

By default `Vault Sidecar Injector` monitors all namespaces (except `kube-system` and `kube-public`) and looks after annotations in submitted pods.

If you want to strictly control the list of namespaces where injection is allowed, set value `mutatingwebhook.namespaceSelector.boolean=true` when installing the chart as follows:

```bash
$ cd deploy/helm

# If using Helm 2
$ helm install $CHART_LOCATION --name vault-sidecar-injector --namespace <namespace for deployment> --set vault.addr=<Vault server address> --set mutatingwebhook.namespaceSelector.boolean=true

# If using Helm 3
$ helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address> --set mutatingwebhook.namespaceSelector.boolean=true
```

Then apply label `vault-injection=enabled` on **all** required namespaces:

```bash
$ kubectl label namespace <namespace> vault-injection=enabled

# check label on namespace
$ kubectl get namespace -L vault-injection
```
</details>

### Uninstalling the chart

To uninstall/delete the `Vault Sidecar Injector` deployment:

```bash
# If using Helm 2
$ helm delete --purge vault-sidecar-injector

# If using Helm 3
$ helm delete vault-sidecar-injector -n kube-system
```

> Note If you encounter issues trying to uninstall the chart, try option `--no-hooks` then remove remaining parts with kubectl cli.

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `Vault Sidecar Injector` chart and their default values.

| Parameter    | Description          | Default                                                         |
|:-------------|:---------------------|:----------------------------------------------------------------|
| hook.image.path            | Image path | bitnami/kubectl |
| hook.image.pullPolicy      | Pull policy for image: IfNotPresent or Always | Always |
| hook.image.tag             | Image tag | latest |
| image.applicationNameLabel   | Application Name. Must match label com.talend.application | talend-vault-sidecar-injector   |
| image.metricsPort                | Port exposed for metrics collection | 9000 |
| image.path       | Image path   | talend/vault-sidecar-injector |
| image.port       | Service main port    | 8443            |
| image.pullPolicy   | Pull policy for image: IfNotPresent or Always       | IfNotPresent           |
| image.serviceNameLabel   | Service Name. Must match label com.talend.service     | talend-vault-sidecar-injector      |
| image.tag  | Image tag     | latest *(local testing)*, [VERSION_VSI](VERSION_VSI) *(release)* |
| imageRegistry  | Image registry |   |
| injectconfig.jobbabysitter.image.path   | Image path | everpeace/curl-jq |
| injectconfig.jobbabysitter.image.pullPolicy | Pull policy for image: IfNotPresent or Always | Always |
| injectconfig.jobbabysitter.image.tag   | Image tag  | latest |
| injectconfig.jobbabysitter.resources.limits.cpu | Job babysitter sidecar CPU resource limits | 20m |
| injectconfig.jobbabysitter.resources.limits.memory | Job babysitter sidecar memory resource limits | 25Mi |
| injectconfig.jobbabysitter.resources.requests.cpu | Job babysitter sidecar CPU resource requests | 15m |
| injectconfig.jobbabysitter.resources.requests.memory | Job babysitter sidecar memory resource requests | 20Mi |
| injectconfig.vault.image.path  | Image path  | vault |
| injectconfig.vault.image.pullPolicy    | Pull policy for image: IfNotPresent or Always  | Always   |
| injectconfig.vault.image.tag  | Image tag | 1.4.2 |
| injectconfig.vault.loglevel                    | Vault log level: trace, debug, info, warn, err    | info    |
| injectconfig.vault.resources.limits.cpu | Vault sidecar CPU resource limits | 50m |
| injectconfig.vault.resources.limits.memory | Vault sidecar memory resource limits | 50Mi |
| injectconfig.vault.resources.requests.cpu | Vault sidecar CPU resource requests | 40m |
| injectconfig.vault.resources.requests.memory | Vault sidecar memory resource requests | 35Mi |
| mutatingwebhook.annotations.appLabelKey | Annotation for application's name. Annotation's value used as Vault role by default. | com.talend.application  |
| mutatingwebhook.annotations.appServiceLabelKey | Annotation for service's name | com.talend.service  |
| mutatingwebhook.annotations.keyPrefix | Prefix used for all vault sidecar injector annotations | sidecar.vault.talend.org  |
| mutatingwebhook.failurePolicy | Defines how unrecognized errors and timeout errors from the admission webhook are handled. Allowed values are Ignore or Fail | Ignore |
| mutatingwebhook.namespaceSelector.boolean    | Enable to control, with label "vault-injection=enabled", the namespaces where injection is allowed (if false: all namespaces except _kube-system_ and _kube-public_) | false                                                           |
| mutatingwebhook.namespaceSelector.namespaced | Enable to control, with label "vault-injection={{ .Release.Namespace }}", the specific namespace where injection is allowed (ie, restrict to namespace where injector is installed) | false |
| probes.liveness.failureThreshold                | Number of probe failure before restarting the probe                                 | 3  |
| probes.liveness.initialDelaySeconds             | Number of seconds after the container has started before the probe is initiated     | 2  |
| probes.liveness.periodSeconds                   | How often (in seconds) to perform the probe                                         | 20 |
| probes.liveness.timeoutSeconds                  | Number of seconds after which the probe times out                                   | 5  |
| probes.readiness.failureThreshold               | Number of probe failure before setting the probe to Unready                         | 3  |
| probes.readiness.initialDelaySeconds            | Number of seconds after the container has started before the probe is initiated     | 2  |
| probes.readiness.periodSeconds                  | How often (in seconds) to perform the probe       | 20   |
| probes.readiness.successThreshold      | Minimum consecutive successes for the probe to be considered successful after having failed  | 1  |
| probes.readiness.timeoutSecon          | Number of seconds after which the probe times out  | 5   |
| registryKey         | Name of Kubernetes secret for image registry                        |  |
| replicaCount                        | Number of replicas | 3    |
| resources.limits.cpu                | CPU resource limits                             | 250m         |
| resources.limits.memory             | Memory resource limits                          | 256Mi        |
| resources.requests.cpu              | CPU resource requests                           | 100m         |
| resources.requests.memory           | Memory resource requests                        | 128Mi        |
| revisionHistoryLimit                | Revision history limit in tiller / helm / k8s   | 3            |
| service.exposedServicePort   | Port exposed by the K8s service (Kubernetes always assumes port 443 for webhooks) | 443 |
| service.name                                    | Service name            | talend-vault-sidecar-injector                                   |
| service.prefixWithHelmRelease                   | Service name to be prefixed with Helm release name                                       | false                                                           |
| service.type                        | Kubernetes service type: ClusterIP, NodePort, LoadBalancer, ExternalName  | ClusterIP    |
| vault.addr                          | Address of Vault server    | `null` - To be provided at deployment time (e.g.: https://vault:8200)   |
| vault.authMethods.approle.path      | Path defined for AppRole Auth Method            | approle |
| vault.authMethods.approle.roleid_filename    | Filename for role id    | approle_roleid   |
| vault.authMethods.approle.secretid_filename  | Filename for secret id  | approle_secretid |
| vault.authMethods.kubernetes.path      | Path defined for Kubernetes Auth Method            | kubernetes |
| vault.ssl.verify               | Enable or disable verification of certificates               | true |

You can override these values at runtime using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
# If using Helm 2
$ helm install <chart_folder_location> \
               --name vault-sidecar-injector \
               --namespace <your namespace> \
               --set <parameter1>=<value1>,<parameter2>=<value2>

# If using Helm 3
$ helm install vault-sidecar-injector \
               <chart_folder_location> \
               --namespace <your namespace> \
               --set <parameter1>=<value1>,<parameter2>=<value2>
```

## Metrics

Vault Sidecar Injector exposes a Prometheus endpoint at `/metrics` on port `metricsPort` (default: 9000).

Following collectors are available:

<details>
<summary>
Process Collector
</summary>

- process_cpu_seconds_total
- process_virtual_memory_bytes
- process_start_time_seconds
- process_open_fds
- process_max_fds
</details>

<details>
<summary>
Go Collector
</summary>

- go_goroutines
- go_threads
- go_gc_duration_seconds
- go_info
- go_memstats_alloc_bytes
- go_memstats_heap_alloc_bytes
- go_memstats_alloc_bytes_total
- go_memstats_sys_bytes
- go_memstats_lookups_total
- ... [full list here](https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go)
</details>

![Grafana dashboard](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/grafana-vault-sidecar-injector.png)

## List of changes

Look at changes for Vault Sidecar Injector releases in [CHANGELOG](https://github.com/Talend/vault-sidecar-injector/blob/master/CHANGELOG.md) file.
