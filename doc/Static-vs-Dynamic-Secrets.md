# Static vs Dynamic Secrets

*March 2020, [Post by Alain Saint-Sever, Senior Cloud Software Architect (@alstsever)](https://twitter.com/alstsever)*

Available with `Vault Sidecar Injector` version **`6.0.0`**, *static secrets* submode (part of **secrets** mode) allows to handle simpler needs where you only want to fetch secrets that are not meant to change over your workload's lifetime. Such secrets may be database credentials (depending on your credentials rotation policy of course) or any confidential data static by nature.

A new annotation, `sidecar.vault.talend.org/secrets-type`, is supported to explicitly define what kind of secrets you intend to fetch, default being *dynamic secrets*.

When *static secrets* are set, `Vault Sidecar Injector` will only inject an init container in your workload's pod. Fetched secrets will be stored in a file in a shared memory volume, the same way it is already done for *dynamic secrets*. As a result, if you do not enable other modes (e.g. *proxy*, *job*) no sidecar will be added. It also means that you don't have to leverage hooks or wait for the injected Vault Agent to fetch your secrets: your workload can access the values right after its container is started. The drawback of course is that your secrets **will not be automatically refreshed upon changes**, opt for *dynamic secrets* if this behavior is required.

If you enable several modes, you may end up with both init container and sidecar(s) in your workload. A comprehensive table is provided in the main documention in section [Modes and Injection Config Overview](https://github.com/Talend/vault-sidecar-injector/blob/master/README.md#modes-and-injection-config-overview).

New [samples](https://github.com/Talend/vault-sidecar-injector/blob/master/samples) are available to quickly demonstrate how to benefit from this feature:

- Deployment workload with only **secrets** mode on for *static secrets*: [manifest](https://github.com/Talend/vault-sidecar-injector/blob/master/samples/app-dep-4-secrets_static.yaml)
- Deployment workload with both **secrets** and **proxy** modes on to handle *static secrets* and direct use of Vault features (cipher/decipher data here) via the proxy: [manifest](https://github.com/Talend/vault-sidecar-injector/blob/master/samples/app-dep-5-secrets_static-proxy.yaml)
- Job workload with only **secrets** mode on for *static secrets*: [manifest](https://github.com/Talend/vault-sidecar-injector/blob/master/samples/app-job-4-secrets_static.yaml)
