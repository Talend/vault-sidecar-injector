# Changelog for Vault Sidecar Injector

## Next Release

Default Vault image set to `1.4.2` to fix several CVEs (CVE-2020-13223, CVE-2020-12757: see HashiCorp's [CHANGELOG](https://github.com/hashicorp/vault/blob/master/CHANGELOG.md#142-may-21st-2020))

**Changed**

- [VSI #29](https://github.com/Talend/vault-sidecar-injector/pull/29) - Update HashiCorp Vault image to 1.4.2

## Release v6.1.0 - 2020-05-18

This release fixes VSI deployment on Kubernetes 1.18+ clusters. It also comes with better AppRole integration and updated Vault image.

**Changed**

- [VSI #27](https://github.com/Talend/vault-sidecar-injector/pull/27) - Update HashiCorp Vault image to 1.4.1

**Added**

- [VSI #26](https://github.com/Talend/vault-sidecar-injector/pull/26) - Improve AppRole support: add tests, enforce check over secrets type, tune Vault Agent config

**Fixed**

- [VSI #25](https://github.com/Talend/vault-sidecar-injector/pull/25) - Fix RBAC following breaking change in [Kubernetes 1.18 Certificates API](https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/20190607-certificates-api.md). See also associated PR [86476](https://github.com/kubernetes/kubernetes/pull/86476) & [86933](https://github.com/kubernetes/kubernetes/pull/86933).

## Release v6.0.1 - 2020-04-06

This is a minor release to update Vault image to `1.3.4` by default (CVE fixes, see details [here](https://github.com/hashicorp/vault/blob/master/CHANGELOG.md#134-march-19th-2020)) and enable offline builds by vendoring dependencies (use `make build OFFLINE=true`).

**Changed**

- [VSI #23](https://github.com/Talend/vault-sidecar-injector/pull/23) - Update HashiCorp Vault image (CVE fixes)

**Added**

- [VSI #24](https://github.com/Talend/vault-sidecar-injector/pull/24) - Vendoring

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
