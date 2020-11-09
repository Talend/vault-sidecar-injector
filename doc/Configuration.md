# Configuration

The following table lists the configurable parameters of the `Vault Sidecar Injector` chart and their default values.

| Parameter    | Description          | Default                                                         |
|:-------------|:---------------------|:----------------------------------------------------------------|
| image.applicationNameLabel   | Application Name. Must match label com.talend.application | talend-vault-sidecar-injector   |
| image.metricsPort                | Port exposed for metrics collection | 9000 |
| image.path       | Image path   | talend/vault-sidecar-injector |
| image.port       | Service main port    | 8443            |
| image.pullPolicy   | Pull policy for image: IfNotPresent or Always       | IfNotPresent           |
| image.serviceNameLabel   | Service Name. Must match label com.talend.service     | talend-vault-sidecar-injector      |
| image.tag  | Image tag     | latest *(local testing)*, [VERSION_VSI](../VERSION_VSI) *(release)* |
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
| injectconfig.vault.image.tag  | Image tag | 1.5.4 |
| injectconfig.vault.log.format                    | Vault log format: standard, json    | json    |
| injectconfig.vault.log.level                    | Vault log level: trace, debug, info, warn, err    | info    |
| injectconfig.vault.resources.limits.cpu | Vault sidecar CPU resource limits | 50m |
| injectconfig.vault.resources.limits.memory | Vault sidecar memory resource limits | 50Mi |
| injectconfig.vault.resources.requests.cpu | Vault sidecar CPU resource requests | 40m |
| injectconfig.vault.resources.requests.memory | Vault sidecar memory resource requests | 35Mi |
| mutatingwebhook.annotations.appLabelKey | Annotation for application's name. Annotation's value used as Vault role by default. | com.talend.application  |
| mutatingwebhook.annotations.appServiceLabelKey | Annotation for service's name | com.talend.service  |
| mutatingwebhook.annotations.keyPrefix | Prefix used for all vault sidecar injector annotations | sidecar.vault.talend.org  |
| mutatingwebhook.cert.cacertfile | Default filename for webhook CA certificate (PEM-encoded) in generated or provided Kubernetes Secret | ca.crt |
| mutatingwebhook.cert.certfile | Default filename for webhook certificate (PEM-encoded) in generated or provided Kubernetes Secret | tls.crt |
| mutatingwebhook.cert.certlifetime | Default lifetime in years for generated certificates. Not used if generated is false. | 10 |
| mutatingwebhook.cert.generated | Controls whether webhook certificates, private key and Kubernetes Secret are generated. If not, you have to provide a Kubernetes Secret with name secretName. | true |
| mutatingwebhook.cert.keyfile | Default filename for webhook private key (PEM-encoded) in generated or provided Kubernetes Secret | tls.key |
| mutatingwebhook.cert.secretName | Name of the Kubernetes Secret that contains the webhook certificates and private key. Secret should be in webhook's namespace. To provide if generated is false. | talend-vault-sidecar-injector-cert |
| mutatingwebhook.failurePolicy | Defines how unrecognized errors and timeout errors from the admission webhook are handled. Allowed values are Ignore or Fail | Ignore |
| mutatingwebhook.loglevel | Enable V-leveled logging at the specified level | 4 |
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
$ helm install vault-sidecar-injector \
               <chart_folder_location> \
               --namespace <your namespace> \
               --set <parameter1>=<value1>,<parameter2>=<value2>
```
