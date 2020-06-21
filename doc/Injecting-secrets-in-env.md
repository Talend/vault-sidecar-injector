# Injecting secrets in environment variables

*June 2020, [Post by Alain Saint-Sever, Senior Cloud Software Architect (@alstsever)](https://twitter.com/alstsever)*

New `Vault Sidecar Injector` **`7.0.0`** release comes with the ability to inject fetched secrets from Vault right into your application's environment variables.

As of today, this feature only supports **static** secrets (i.e. secrets whose values will not be updated). To leverage it, just add new annotation `sidecar.vault.talend.org/secrets-injection-method` with value `env` and you are good to go.

As an example, available in the [samples](samples) folder:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
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
        sidecar.vault.talend.org/secrets-injection-method: "env"
        sidecar.vault.talend.org/secrets-type: "static"
      labels:
        com.talend.application: test
        com.talend.service: test-app-svc
    spec:
      serviceAccountName: default
      containers:
        - name: app-container
          image: busybox:1.28
          command:
            - "sh"
            - "-c"
            - |
              set -e
              echo "My secrets are: SECRET1_FROM_ENV=${SECRET1}, SECRET2_FROM_ENV=${SECRET2}"
              while true;do
                sleep 5
              done
```

Environment variables holding your secrets are named after the keys you defined in Vault (in the example above: `SECRET1` and `SECRET2`).

Last important point for this feature to work is to always provide a `command` attribute for your container(s), even if you already define a ENTRYPOINT/CMD directive in your images: the webhook has no means to determine them so the container's process has to be explicitly set in the manifest.
