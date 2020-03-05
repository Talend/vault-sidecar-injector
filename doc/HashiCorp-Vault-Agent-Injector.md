# Vault Sidecar Injector vs HashiCorp Vault Agent Injector - Features Comparison

*March 2020, [Post by Alain Saint-Sever, Senior Cloud Software Architect (@alstsever)](https://twitter.com/alstsever)*

- [Vault Sidecar Injector vs HashiCorp Vault Agent Injector - Features Comparison](#vault-sidecar-injector-vs-hashicorp-vault-agent-injector---features-comparison)
  - [Intro](#intro)
  - [What we will test](#what-we-will-test)
    - [Cluster and Vault server installation](#cluster-and-vault-server-installation)
    - [Vault Sidecar Injector installation](#vault-sidecar-injector-installation)
    - [HashiCorp Vault Agent Injector installation](#hashicorp-vault-agent-injector-installation)
  - [Test Workloads](#test-workloads)
  - [Features Comparison](#features-comparison)

## Intro

In December 2019, HashiCorp announced availaility of their Vault Agent Injector to fulfill the same need we address with our injector: provide a transparent way to fetch static and dynamic secrets from Vault stores.

Following major release of version `6.0.0` of `Vault Sidecar Injector`, and nearly three months past HashiCorp's injector first delivery, let's seize the opportunity to do a features to features comparison to assess how well (or how far) we stand and identify any improvements we may add to our roadmap to still bring value to the end users.

Links:

- Injecting Vault Secrets Into Kubernetes Pods via a Sidecar -- <https://www.hashicorp.com/blog/injecting-vault-secrets-into-kubernetes-pods-via-a-sidecar/>
- Agent Sidecar Injector -- <https://www.vaultproject.io/docs/platform/k8s/injector/index.html>

## What we will test

Tests and comparisons will be done based on following versions (last releases at the time of writing):

| Injector | Version | Helm chart version | Vault Agent version | Repo |
|----------|---------|--------------------|---------------------|------|
| Vault Sidecar Injector | `6.0.0` | `3.2.0` | `1.3.2` | <https://github.com/Talend/vault-sidecar-injector/releases/tag/v6.0.0><BR><https://hub.helm.sh/charts/talend/vault-sidecar-injector/3.2.0> |
| HashiCorp Vault Agent Injector | `0.2.0` | `0.4.0` | `1.3.2` | <https://github.com/hashicorp/vault-k8s/releases/tag/v0.2.0><BR><https://github.com/hashicorp/vault-helm/releases/tag/v0.4.0> |

For both injectors, installation will be performed using the Helm chart, which is the recommended installation method according to HashiCorp (see <https://github.com/hashicorp/vault-k8s#installation>).

The test Vault server will also be deployed using the Helm chart from HashiCorp and set up with the initialization script we provide as part of our test environment: [init-dev-vault-server.sh](https://github.com/Talend/vault-sidecar-injector/tree/v6.0.0/deploy/vault/init-dev-vault-server.sh).

The test platform will consist in a Minikube cluster (Minikube version `1.7.3` running Kubernetes `1.17.3`).

Last, Helm 2 will be used to deploy the charts but this is not a requirement (Helm 3 can be used instead with little changes in the instructions given below).

### Cluster and Vault server installation

Our Minikube cluster is started on a Linux Ubuntu 18.04.4 LTS station using KVM. We allocate 4 cpus and 16G of memory to the VM.

```bash
$ minikube start --cpus 4 --memory 16384 --vm-driver kvm2
```

Once cluster is ready, we deploy our test Vault server (note that the HashiCorp injector *is not installed at this stage* to clearly separate each step):

```bash
$ git clone https://github.com/hashicorp/vault-helm.git
$ cd vault-helm
$ git checkout v0.4.0
$ helm install . --name=vault --set injector.enabled=false --set server.dev.enabled=true --set ui.enabled=true --set ui.serviceType="NodePort"
```

Next, apply our test configuration:

```bash
$ git clone https://github.com/Talend/vault-sidecar-injector.git
$ cd vault-sidecar-injector
$ git checkout v6.0.0
$ cd deploy/vault
$ ./init-dev-vault-server.sh
```

Your Vault server is now ready, UI is also available on following local URL:

```bash
$ minikube service list|grep vault-ui
```

Open a browser on given URL, enter *root* as token value to sign in.

### Vault Sidecar Injector installation

From `vault-sidecar-injector` GitHub repo:

```bash
$ cd vault-sidecar-injector/deploy/helm
$ helm install . --name vault-sidecar-injector --namespace kube-system --set vault.addr=http://vault:8200 --set vault.ssl.verify=false
```

To uninstall:

```bash
$ helm delete --purge vault-sidecar-injector
```

### HashiCorp Vault Agent Injector installation

From `vault-helm` GitHub repo:

```bash
$ cd vault-helm
$ helm install . --name=hashicorp-vault-injector --set injector.enabled=true --set injector.externalVaultAddr=http://vault:8200 --set server.standalone.enabled=false --set server.service.enabled=false --set server.dataStorage.enabled=false
```

Using this command line, we **only** install the HashiCorp injector and tell him to make use of the Vault server we deployed previously. By doing so, it is easy to test one injector or the other, keeping our Vault server in place. Of course, it is possible to install both Vault server and the HashiCorp injector at once if you want to.

To uninstall:

```bash
$ helm delete --purge hashicorp-vault-injector
```

## Test Workloads

Assessment has been done with the help of following workloads:

- For Vault Sidecar Injector: [samples](https://github.com/Talend/vault-sidecar-injector/tree/v6.0.0/samples)
- For HashiCorp Vault Agent Injector: [hashicorp-injector-samples](https://github.com/Talend/vault-sidecar-injector/tree/v6.0.0/doc/hashicorp-injector-samples)

## Features Comparison

Features comparison `Vault Sidecar Injector` vs `HashiCorp Vault Agent Injector` **as of March 2020**:

> Note: both injectors rely on the same version of the injected Vault Agent container. Thus feature gaps highlighted below lie in the webhook implementation itself.

|Features|Vault Sidecar Injector|HashiCorp Vault Agent Injector|
|--------|----------------------|------------------------------|
|Vault Agent image|               At webhook level *(helm chart value)*    |    Both at webhook *(helm chart value)* and pod levels *(via annotation)* |
| Shared secrets volume| Injection of In-memory Volume **only** if not defined. VolumeMount is not injected if not defined | Injection of both In-memory Volume and VolumeMount. **Failure** if `vault-secrets` Volume already defined |
| Secrets volume mount path |  Any path associated to the `secrets` Volume   | Cannot be changed *(set to `/vault/secrets`)* |
| Kubernetes Job support for both static and dynamic secrets | **Yes**, with special mechanism to monitor injected sidecars | **Static secrets only**, do not properly support dynamic secrets (sidecar **never ends**)  |
| Vault K8S Auth with custom Service Account Token | At pod level *(using annotation)* | **Not possible** |
| Multi-secrets, multi-templates support | At pod level *(using annotation)* | At pod level *(using annotation)* |
| Vault Role | At pod level *(using annotation or label)* | At pod level *(using annotation)* |
| Vault secrets path | At pod level *(using annotation or label)* | At pod level *(using annotation)* |
| Vault proxy mode | At pod level *(using annotation)* | By providing K8S ConfigMap with custom Vault Agent config + using annotation to load config *(`vault.hashicorp.com/agent-configmap`)* |
| Custom Vault Agent config | **Not possible** | By providing K8S ConfigMap with custom Vault Agent config + using annotation to load config *(`vault.hashicorp.com/agent-configmap`)* |
| Resources (CPU, mem) for webhook | Using Helm chart values | Using Helm chart values |
| Resources (CPU, mem) for injected container(s) | At webhook level *(helm chart values)* | At pod level *(default values or custom via annotations)* |
| Vault server to use | At webhook level *(helm chart value)* | Both at webhook *(helm chart value)* and pod levels *(via annotation)* |
| Vault K8S Auth path | At webhook level *(helm chart value)* | At pod level *(default to `kubernetes` or custom via annotation)* |
| Vault Agent's check of Vault's TLS cert | At webhook level *(helm chart value)* | At pod level *(using annotation)* |
| Vault AppRole Auth support | At pod level *(using annotation)* | By providing K8S ConfigMap with custom Vault Agent config + using annotation to load config *(`vault.hashicorp.com/agent-configmap`)* |
| Injection of init and sidecar containers | At pod level, injection content **fully based on enabled modes** (see [table](https://github.com/Talend/vault-sidecar-injector/tree/v6.0.0/README.md#modes-and-injection-config-overview)) *(using annotations)* | At pod level, injection content **only based on annotations** (`vault.hashicorp.com/agent-pre-populate` and `vault.hashicorp.com/agent-pre-populate-only`) *(inject both init and sidecar container by default)* |
| Handling Vault Dynamic Secrets (e.g. AWS, Azure) | | |

Most of the differences are less the result of technical choice than philosophical ones: the Vault Sidecar Injector is more "user friendly" in this regard by easily giving access to Vault proxy mode or other technical features through a simple annotation, where the same capability on HashiCorp's injector will require the user to provide a complete Vault Agent config wrapped into a Kubernetes ConfigMap. This distinct approach can also be seen with Vault Sidecar Injector's modes that completely free the user from knowing whether he needs an init container or a sidecar or both of them to handle a use case. On the other end, with the HashiCorp's injector, the user has control over the injected content and this "complexity" allows for greater flexibility.

The major advantage brought by the Vault Sidecar Injector lies in how it supports dynamic secrets in Kubernetes Jobs, a feature currently not properly implemented on HashiCorp side.

Results from this comparison test are instrumental to drive the roadmap and content of the next releases of the Vault Sidecar Injector. So stay tuned!
