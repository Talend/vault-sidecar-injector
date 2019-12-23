# Changelog for Vault Sidecar Injector

## Release v5.1.1 - xxx

TODO

Vault 1.3.1

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
