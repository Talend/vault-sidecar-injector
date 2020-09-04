
# Vault Sidecar Injector

[![GitHub All Releases](https://img.shields.io/github/downloads/Talend/vault-sidecar-injector/total?style=for-the-badge)](https://github.com/Talend/vault-sidecar-injector/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/talend/vault-sidecar-injector?style=for-the-badge)](https://hub.docker.com/r/talend/vault-sidecar-injector)

`Vault Sidecar Injector` allows to dynamically inject HashiCorp Vault Agent as either an init or a sidecar container, along with configuration and volumes, in any matching pod manifest **to seamlessly fetch secrets from Vault**. Pods willing to benefit from this feature just have to add some custom annotations to ask for the injection **at deployment time**.

## Announcements

- 2020-06: [Inject secrets in your pod as environment variables](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/Injecting-secrets-in-env.md)
- 2020-03: [Vault Sidecar Injector vs HashiCorp Vault Agent Injector - Features Comparison](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/HashiCorp-Vault-Agent-Injector.md)
- 2020-03: [Static vs Dynamic secrets](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/Static-vs-Dynamic-Secrets.md)
- 2019-12: [Discovering Vault Sidecar Injector's Proxy feature](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/Discovering-Vault-Sidecar-Injector-Proxy.md)
- 2019-11: [Vault Sidecar Injector now leverages Vault Agent Template feature](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/Leveraging-Vault-Agent-Template.md)
- 2019-10: [Open-sourcing Vault Sidecar Injector](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/announcements/Open-sourcing-Vault-Sidecar-Injector.md)

## Kubernetes compatibility

`Vault Sidecar Injector` can be deployed on Kubernetes `1.12` and higher. Deployment on earlier versions *may work* but has not been tested.

## Usage

- [How to invoke Vault Sidecar Injector](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Usage.md)
- [Examples](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Examples.md)

## Installation

- [How to deploy Vault Sidecar Injector](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Deploy.md)
- [Configuration](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Configuration.md)

## Observability

- [Metrics](https://github.com/Talend/vault-sidecar-injector/blob/master/doc/Metrics.md)

## List of changes

Look at changes for Vault Sidecar Injector releases in [CHANGELOG](https://github.com/Talend/vault-sidecar-injector/blob/master/CHANGELOG.md) file.

## Contributing

Feel free to create [issues](https://github.com/Talend/vault-sidecar-injector/issues) or submit [pull requests](https://github.com/Talend/vault-sidecar-injector/pulls).

## License

This project is licensed under the terms of the Apache 2.0 license.
