# How to invoke Vault Sidecar Injector

- [How to invoke Vault Sidecar Injector](#how-to-invoke-vault-sidecar-injector)
  - [Modes](#modes)
  - [Requirements](#requirements)
  - [Annotations](#annotations)
  - [Secrets Mode](#secrets-mode)
    - [Default template](#default-template)
    - [Template's Syntax](#templates-syntax)
  - [Proxy Mode](#proxy-mode)
  - [Modes and Injection Config Overview](#modes-and-injection-config-overview)

> ⚠️ **Important note** ⚠️: support for sidecars in Kubernetes **jobs** suffers from limitations and issues exposed here: <https://github.com/kubernetes/kubernetes/issues/25908>.
>
> Fortunately, `Vault Sidecar Injector` implements **specific sidecar and signaling mechanism** to properly stop all injected containers on job termination.

## Modes

`Vault Sidecar Injector` supports several high-level features or *modes*:

- [**secrets**](#secrets-mode), the primary mode allowing to retrieve secrets from Vault server's stores, either once (for **static secrets**) or continuously (for **dynamic secrets**), coping with secrets rotations (ie any change will be propagated and updated values made available to consume by applications).
- [**proxy**](#proxy-mode), to enable the injected Vault Agent as a local, authenticated gateway to the remote Vault server. As an example, with this mode on, applications can easily leverage Vault's Transit Engine to cipher/decipher payloads by just sending data to the local proxy without dealing themselves with Vault authentication and tokens.
- **job**, to use when a Kubernetes Job is submitted. This new mode comes in replacement of the now deprecated `sidecar.vault.talend.org/workload` annotation.

For details, refer to [Modes and Injection Config Overview](#modes-and-injection-config-overview).

## Requirements

Invoking `Vault Sidecar Injector` is pretty straightforward: in your application manifest, just add annotation **`sidecar.vault.talend.org/inject: "true"`**. This is the only mandatory annotation (see list of supported annotations below).

By default, when using **secrets** mode, deciphered secrets are made available in file `secrets.properties` (using format `<secret key>=<secret value>`) under folder `/opt/talend/secrets`. You can change the secrets filename using an annotation, and the location by mounting the `secrets` injected volume where you want to.

Refer to provided [sample files](../samples) and [Examples](Examples.md) document.

## Annotations

Following annotations in requesting pods are supported:

| Annotation                            | (M)andatory / (O)ptional |  Apply to mode | Default Value        | Supported Values               | Description |
|---------------------------------------|--------------------------|-----------------|----------------------|--------------------------------|-------------|
| `sidecar.vault.talend.org/inject`     | M           |    N/A          |                      | "true" / "on" / "yes" / "y"  | Ask for injection to get secrets from Vault    |
| `sidecar.vault.talend.org/vault-image` | O          |    N/A          | "<`injectconfig.vault.image.path` Helm value>:<`injectconfig.vault.image.tag` Helm value>"  | Any image with Vault installed | The image to be injected in your pod |
| `sidecar.vault.talend.org/auth`       | O           |    N/A          | "kubernetes"   | "kubernetes" / "approle" | Vault Auth Method to use. **Static secrets only supports "kubernetes" authentication method** |
| `sidecar.vault.talend.org/mode`       | O           |    N/A          | "secrets"      | "secrets" / "proxy" / "job" / Comma-separated values (eg "secrets,proxy") | Enable provided mode(s). **Note: `secrets` mode will be enabled if you only set `job` mode**   |
| `sidecar.vault.talend.org/notify`     | O           |    secrets   | ""   | Comma-separated strings  | List of commands to notify application/service of secrets change, one per secrets path. **Usage context: dynamic secrets only** |
| `sidecar.vault.talend.org/proxy-port` | O           |    proxy        | "8200"    | Any allowed port value  | Port for local Vault proxy |
| `sidecar.vault.talend.org/role`       | O           |    N/A          | "\<`com.talend.application` label\>" | Any string    | **Only used with "kubernetes" Vault Auth Method**. Vault role associated to requesting pod. If annotation not used, role is read from label defined by `mutatingwebhook.annotations.appLabelKey` key (refer to [configuration](Configuration.md)) which is `com.talend.application` by default |
| `sidecar.vault.talend.org/sa-token`   | O           |    N/A         | "/var/run/secrets/kubernetes.io/serviceaccount/token" | Any string | Full path to service account token used for Vault Kubernetes authentication |
| `sidecar.vault.talend.org/secrets-destination` | O     | secrets | "secrets.properties" | Comma-separated strings  | List of secrets filenames (without path), one per secrets path |
| `sidecar.vault.talend.org/secrets-hook`        | O     | secrets |  | "true" / "on" / "yes" / "y" | If set, lifecycle hooks will be added to pod's container(s) to wait for secrets files. **Usage context: dynamic secrets only. Do not use with `job` mode** |
| `sidecar.vault.talend.org/secrets-injection-method` | O   | secrets | "file" | "file" / "env" | Method used to provide secrets to applications. **Note: `env` method only supports static secrets** |
| `sidecar.vault.talend.org/secrets-path`        | O     | secrets | "secret/<`com.talend.application` label>/<`com.talend.service` label>" | Comma-separated strings | List of secrets engines and path. If annotation not used, path is set from labels defined by `mutatingwebhook.annotations.appLabelKey`  and `mutatingwebhook.annotations.appServiceLabelKey` keys (refer to [configuration](Configuration.md))      |
| `sidecar.vault.talend.org/secrets-template`    | O     | secrets  | [Default template](#default-template) | templates separated with `---` | Allow to override default template. Ignore `sidecar.vault.talend.org/secrets-path` annotation if set |
| `sidecar.vault.talend.org/secrets-type` | O  | secrets | "dynamic" | "static" / "dynamic" | Type of secrets to handle (see details [here](announcements/Static-vs-Dynamic-Secrets.md)) |
| `sidecar.vault.talend.org/workload`   | O      | N/A |   | "job" | Type of submitted workload. **⚠️ Deprecated: use `sidecar.vault.talend.org/mode` instead. Using this annotation will enable `job` mode ⚠️** |

Upon successful injection, Vault Sidecar Injector will add annotation(s) to the requesting pods:

| Annotation                        | Value      | Description                                 |
|-----------------------------------|------------|---------------------------------------------|
| `sidecar.vault.talend.org/status` | "injected" | Status set by Vault Sidecar Injector        |

> **Note:** you can change the annotation prefix (set by default to `sidecar.vault.talend.org`) thanks to `mutatingwebhook.annotations.keyPrefix` key in [configuration](Configuration.md).

## Secrets Mode

### Default template

Template below is used by default to fetch all secrets and create corresponding key/value pairs. It is generic enough and should be fine for most use cases:

<!-- {% raw %} -->
```yaml
{{ with secret "<Path to secrets (defaut value or `sidecar.vault.talend.org/secrets-path` annotation)>" }}{{ range $k, $v := .Data }}
{{ $k }}={{ $v }}
{{ end }}{{ end }}
```
<!-- {% endraw %}) -->

Using annotation `sidecar.vault.talend.org/secrets-template` it is nevertheless possible to provide your own list of templates. For some examples have a look at the [Examples](Examples.md) document.

### Template's Syntax

Details on template syntax can be found in Consul Template's documentation (same syntax is supported by Vault Agent Template):

- <https://github.com/hashicorp/consul-template#secret>
- <https://github.com/hashicorp/consul-template#secrets>
- <https://github.com/hashicorp/consul-template#helper-functions>

## Proxy Mode

This mode opens the gate to virtually any Vault features for requesting applications. A [blog entry](announcements/Discovering-Vault-Sidecar-Injector-Proxy.md) introduces this mode and examples are provided.

## Modes and Injection Config Overview

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
