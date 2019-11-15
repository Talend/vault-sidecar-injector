FROM golang:1.12.9 AS buildTVSI

COPY . /vaultsidecarinjector
RUN cd /vaultsidecarinjector && make build

FROM centos:7.7.1908

ENV TALEND_HOME=/opt/talend

LABEL com.talend.maintainer="Talend <support@talend.com>" \
      com.talend.url="https://www.talend.com/" \
      com.talend.vendor="Talend" \
      com.talend.name="Vault Sidecar Injector" \
      com.talend.application="talend-vault-sidecar-injector" \
      com.talend.service="talend-vault-sidecar-injector" \
      com.talend.description="Kubernetes Webhook Admission Server for Vault sidecar injection"

COPY --from=buildTVSI /vaultsidecarinjector/target/vaultinjector-webhook ${TALEND_HOME}/webhook/

ENTRYPOINT ["/opt/talend/webhook/vaultinjector-webhook"]
