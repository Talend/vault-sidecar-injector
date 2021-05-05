# How to deploy Vault Sidecar Injector

- [How to deploy Vault Sidecar Injector](#how-to-deploy-vault-sidecar-injector)
  - [Prerequisites](#prerequisites)
  - [Vault Sidecar Injector image](#vault-sidecar-injector-image)
  - [Webhook certificates](#webhook-certificates)
  - [Installing the Chart](#installing-the-chart)
  - [Uninstalling the chart](#uninstalling-the-chart)

`Vault Sidecar Injector` consists in a *Webhook Admission Server*, registered in the Kubernetes [Mutating Admission Webhook Controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks), that will mutate resources depending on defined criteriae.

## Prerequisites

> *Note: Vault Sidecar Injector chart version > `4.0.0` requires Helm 3. Helm 2 is no more supported.*

Installation:

- Kubernetes cluster (see compatibility [here](../README.md#kubernetes-compatibility))
- Helm 3

Runtime:

- Vault server deployed (either *in cluster* with official chart <https://github.com/hashicorp/vault-helm> or *out of cluster*), started and reachable through Kubernetes service & endpoint deployed into cluster

<details>
<summary>
<b>Vault Server installation</b>
</summary>

> **Note:** this step is optional if you already have a running Vault server. This section helps you setup a test Vault server with ready to use configuration.

We will install a test Vault server in Kubernetes cluster but an external, out of cluster, Vault server can also be used. Note that we will install Vault server in *dev mode* below, do not use this setup in production.

Using HashiCorp's Vault Helm chart:

```bash
git clone https://github.com/hashicorp/vault-helm.git
cd vault-helm
git checkout v0.9.1
helm install vault . --set injector.enabled=false --set server.dev.enabled=true --set ui.enabled=true --set ui.serviceType="NodePort"
```

Then init Vault server with our test config:

```bash
# Check status
kubectl exec -it vault-0 -- vault status
kubectl logs vault-0

# Set up needed auth methods, secrets engines, policies, roles and secrets
cd vault-sidecar-injector/test/vault
./init-test-vault-server.sh
```
</details>

## Vault Sidecar Injector image

> Note: if you don't intend to perform some tests with the image you can skip this section.

<details>
<summary>
<b>Pulling the image from Docker Hub</b>
</summary>

Official Docker images are published on [Talend's public Docker Hub](https://hub.docker.com/r/talend/vault-sidecar-injector) repository for each `Vault Sidecar Injector` release. Provided Helm chart will pull the image automatically if needed.  

For manual pull of a specific tag:

```bash
docker pull talend/vault-sidecar-injector:<tag>
```
</details>

<details>
<summary>
<b>Building the image</b>
</summary>

A [Dockerfile](../Dockerfile) is also provided to both compile `Vault Sidecar Injector` and build the image locally if you prefer.

Just run following command:

```bash
make image
```

> Note: if you have Go installed on your machine, you can use `make image-from-build` instead. You need Golang 1.14 or higher.

</details>

## Webhook certificates

By default, the webhook certificates (CA and leaf) and private key will be generated as part of the installation. Look at the `mutatingwebhook.cert.*` parameters in [configuration](Configuration.md) for default values.

You can also provide your own certificates and private key by following those steps:

1) set `mutatingwebhook.cert.generated` parameter to `false`
2) as an option, modify the name of the Kubernetes Secret that will host the certificates and private key (`mutatingwebhook.cert.secretName` parameter)
3) generate the CA, leaf certificate and private key (using OpenSSL for e.g.) and save them as PEM-encoded files
4) from those files, create a new Kubernetes Secret using default name or the one you set in step 2:

  ```sh
  kubectl create secret generic <secret name> \
                  --from-file=ca.crt=<PATH>/<CA file, PEM-encoded> \
                  --from-file=tls.crt=<PATH>/<Cert file, PEM-encoded> \
                  --from-file=tls.key=<PATH>/<PrivKey file, PEM-encoded>
                  -n <Namespace where Vault Sidecar Injector is installed>
  ```

## Installing the Chart

Several options to install the chart:

- from [Artifact Hub](https://artifacthub.io/packages/helm/talend/vault-sidecar-injector) leveraging [Talend's public Helm charts registry](https://talend.github.io/helm-charts-public)
- by downloading the chart archive (`.tgz` file) from GitHub [releases](https://github.com/Talend/vault-sidecar-injector/releases)
- or cloning `Vault Sidecar Injector` GitHub repo and cd into `deploy/helm` directory

Depending on what you chose, define a `CHART_LOCATION` env var as follows:

- if you use [Artifact Hub](https://artifacthub.io/packages/helm/talend/vault-sidecar-injector) / [Talend's public Helm charts registry](https://talend.github.io/helm-charts-public):

```bash
helm repo add talend https://talend.github.io/helm-charts-public/stable
helm repo update
export CHART_LOCATION=talend/vault-sidecar-injector
```

- if you use the downloaded chart archive:

```bash
export CHART_LOCATION=./vault-sidecar-injector-<x.y.z>.tgz
```

- if you install from the chart's folder:

> *Note: you previously need to build the image to use this install option, refer to "Building the image" in [Vault Sidecar Injector image](#vault-sidecar-injector-image)*

```bash
cd deploy/helm
export CHART_LOCATION=$(pwd)
```

To see Chart content before installing it, perform a dry run first:

```bash
helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address> --debug --dry-run
```

To install the chart on the cluster:

```bash
helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address>
```

> **Note:** `Vault Sidecar Injector` should be deployed only once (except for testing purpose, see below). It will mutate any "vault-sidecar annotated" pod from any namespace. **It *shall not* be deployed in every namespaces**.

>**Note**: it is possible to deploy an instance in a given namespace **and to restrict injection to this same namespace** if necessary, **in particular in a dev environment where each team wants its own instance of `Vault Sidecar Injector` for testing purpose** with its dedicated configuration (including a dedicated Vault server). Refer to `Installing the chart in a dev environment` section below.

As an example, to install `Vault Sidecar Injector` on our test cluster:

```bash
helm install vault-sidecar-injector $CHART_LOCATION --namespace kube-system --set vault.addr=http://vault:8200 --set vault.ssl.verify=false
```

This command deploys the component on the Kubernetes cluster with modified configuration to target our Vault server in-cluster test instance (no verification of certificates): such settings *are no fit for production*.

The [configuration](Configuration.md) section lists all the parameters that can be configured during installation.

<details>
<summary>
<b>Installing the chart in a dev environment</b>
</summary>

In a dev environment, you may want to install your own test instance of `Vault Sidecar Injector`, connected to your own Vault server and limiting injection to a given namespace. To do so, use following options:

```bash
helm install vault-sidecar-injector $CHART_LOCATION --namespace <your dev namespace> --set vault.addr=<your dev Vault server address> --set mutatingwebhook.namespaceSelector.namespaced=true
```

And then **add a label on your namespace** as follows (if not done, no injection will be performed):

```bash
kubectl label namespace <your dev namespace> vault-injection=<your dev namespace> --overwrite

# check label on namespace
kubectl get namespace -L vault-injection
```
</details>

<details>
<summary>
<b>Restrict injection to specific namespaces</b>
</summary>

By default `Vault Sidecar Injector` monitors all namespaces (except `kube-system` and `kube-public`) and looks after annotations in submitted pods.

If you want to strictly control the list of namespaces where injection is allowed, set value `mutatingwebhook.namespaceSelector.boolean=true` when installing the chart as follows:

```bash
helm install vault-sidecar-injector $CHART_LOCATION --namespace <namespace for deployment> --set vault.addr=<Vault server address> --set mutatingwebhook.namespaceSelector.boolean=true
```

Then apply label `vault-injection=enabled` on **all** required namespaces:

```bash
kubectl label namespace <namespace> vault-injection=enabled

# check label on namespace
kubectl get namespace -L vault-injection
```
</details>

## Uninstalling the chart

To uninstall/delete the `Vault Sidecar Injector` deployment:

```bash
helm delete vault-sidecar-injector -n <namespace for deployment>
```

This command removes all the Kubernetes resources associated with the chart and deletes the Helm release.
