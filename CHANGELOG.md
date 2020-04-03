# Changelog for Vault Sidecar Injector

## Release v6.0.1 - 2020-04-xx

This is a minor release to update Vault image to `1.3.4` by default to benefit from CVE fixes (see details [here](https://github.com/hashicorp/vault/blob/master/CHANGELOG.md#134-march-19th-2020)).

## Release v6.0.0 - 2020-03-04

This is a major release introducing new features and complete code refactoring for clear isolation of modes.

Highlights:

- New Static Secrets feature, part of `secrets` mode (now supporting both **dynamic** and **static** secrets)
- Kubernetes Jobs are now handled as a *Vault Sidecar Injector mode*. Annotation `sidecar.vault.talend.org/workload` is **still supported but deprecated**: make use of `sidecar.vault.talend.org/mode` to enable job mode
- HashiCorp Vault image updated to `1.3.2`

**Added**

- [VSI #20](https://github.com/Talend/vault-sidecar-injector/pull/20) - Static secrets. Feature announcement [here](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Static-vs-Dynamic-Secrets.md).

## Release v5.1.1 - 2019-12-23

**Added**

- [VSI #18](https://github.com/Talend/vault-sidecar-injector/pull/18) - Basis for new inline injection feature

**Fixed**

- [VSI #16](https://github.com/Talend/vault-sidecar-injector/issues/16) - secrets-template with >1 templates that include range statement causes dest/template mismatch [Thanks @smurfralf]
- [VSI #15](https://github.com/Talend/vault-sidecar-injector/issues/15) - Document requirement for configured certificates api [Thanks @drpebcak]

## Release v5.1.0 - 2019-12-09

- [VSI #14](https://github.com/Talend/vault-sidecar-injector/pull/14) - Minor updates to Helm chart and documentation.

## Release v5.0.0 - 2019-12-06

- [VSI #13](https://github.com/Talend/vault-sidecar-injector/pull/13) - New [Proxy](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Discovering-Vault-Sidecar-Injector-Proxy.md) mode. Injected Vault Agent sidecar can act as a local proxy forwarding application requests to Vault server.

## Release v4.1.0 - 2019-11-24

- [VSI #12](https://github.com/Talend/vault-sidecar-injector/pull/12) - Image based on CentOS `7.7` and run as non-root, chart available on Helm Hub

## Release v4.0.0 - 2019-11-15

- [VSI #9](https://github.com/Talend/vault-sidecar-injector/pull/9) - Remove Consul Template sidecar and use Vault 1.3.0 new agent template feature to fetch secrets. See announcement [here](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Leveraging-Vault-Agent-Template.md).
- [VSI #10](https://github.com/Talend/vault-sidecar-injector/pull/10) - Helm chart is now part of the released artifacts.

## Release v3.0.0 - 2019-10-11

- First open source release of Talend Vault Sidecar Injector component
