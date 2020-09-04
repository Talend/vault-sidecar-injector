# Examples

- [Examples](#examples)
  - [Using Vault Kubernetes Auth Method](#using-vault-kubernetes-auth-method)
    - [Secrets mode - Usage with a K8S Deployment workload](#secrets-mode---usage-with-a-k8s-deployment-workload)
    - [Secrets mode - Static secrets](#secrets-mode---static-secrets)
    - [Secrets mode - Injection into environment variables](#secrets-mode---injection-into-environment-variables)
    - [Secrets mode - Usage with a K8S Job workload](#secrets-mode---usage-with-a-k8s-job-workload)
    - [Secrets and Proxy modes - K8S Job workload](#secrets-and-proxy-modes---k8s-job-workload)
    - [Secrets mode - Custom secrets path and notification command](#secrets-mode---custom-secrets-path-and-notification-command)
    - [Secrets mode - Ask for secrets hook injection, custom secrets file, location and template](#secrets-mode---ask-for-secrets-hook-injection-custom-secrets-file-location-and-template)
    - [Secrets mode - Ask for secrets hook injection, several custom secrets files and templates](#secrets-mode---ask-for-secrets-hook-injection-several-custom-secrets-files-and-templates)
  - [Using Vault AppRole Auth Method](#using-vault-approle-auth-method)

**Ready to use sample manifests are provided under [samples](../samples) folder**. Just deploy them using `kubectl apply -f <sample file>`.

Examples hereafter highlight all the features of `Vault Sidecar Injector` through the supported [annotations](Usage.md#annotations).

## Using Vault Kubernetes Auth Method

### Secrets mode - Usage with a K8S Deployment workload

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
```
</details>

### Secrets mode - Static secrets

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
```
</details>

### Secrets mode - Injection into environment variables

<details>
<summary>
Show example
</summary>

With the provided manifest below:

- Vault authentication done using role `test-app-1` (value of `com.talend.application` label)
- secrets fetched from Vault's path `secret/test-app-1/test-app-1-svc` using default template
- secrets to be injected into environment values named after secret keys

> Note: you **must** set an explicit `command` attribute on containers (even if your image already defines a default ENTRYPOINT/CMD directive) to let the injector determine the process to be run once secrets are mounted as environment variables.

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
        sidecar.vault.talend.org/secrets-injection-method: "env"  # Secrets from env vars
        sidecar.vault.talend.org/secrets-type: "static" # env vars injection only support static secrets
      labels:
        com.talend.application: test-app-1
        com.talend.service: test-app-1-svc
    spec:
      serviceAccountName: ...
      containers:
        - name: ...
          image: ...
          command: ...
          ...
```

Refer to [samples](samples) folder for complete examples.

</details>

### Secrets mode - Usage with a K8S Job workload

When using Kubernetes Jobs, there are specific additional requirements:

- **Use of `serviceAccountName` attribute** in manifest, with role allowing to perform GET action on pods (needed to poll for job's pod status)
- **Do not make use of annotation `sidecar.vault.talend.org/secrets-hook`** as it will immediately put the job in error state. An error will be raised if this annotation is set on jobs.

> Note: This hook is meant to be used with regular workloads only (ie Kubernetes Deployments) as it forces a restart of the application container until secrets are available in application's context. With jobs, as `Vault Sidecar Injector` looks after status of the job container, the injected signaling mechanism will terminate all the sidecars upon job exit thus preventing use of the hook.

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
```
</details>

### Secrets and Proxy modes - K8S Job workload

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
```
</details>

### Secrets mode - Custom secrets path and notification command

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

### Secrets mode - Ask for secrets hook injection, custom secrets file, location and template

<details>
<summary>
Show example
</summary>

Several optional annotations to end up with:

- Vault authentication using role `test-app-6` (value of `com.talend.application` label)
- hook injected in application's container(s) to wait for secrets file availability
- secrets fetched from Vault's path `aws/creds/test-app-6` using **one custom template**
- secrets to be stored into `/custom-folder/creds.properties`

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
              mountPath: /custom-folder
      volumes:
        - name: secrets
          emptyDir:
            medium: Memory
```
<!-- {% endraw %}) -->
</details>

### Secrets mode - Ask for secrets hook injection, several custom secrets files and templates

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
```
<!-- {% endraw %}) -->
</details>

## Using Vault AppRole Auth Method

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
```
</details>
