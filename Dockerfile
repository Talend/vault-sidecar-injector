FROM centos:7.6.1810

ENV TALEND_HOME=/opt/talend

LABEL com.talend.maintainer="Talend <support@talend.com>" \
      com.talend.url="https://www.talend.com/" \
      com.talend.vendor="Talend" \
      com.talend.name="Talend Vault Sidecar Injector" \
      com.talend.application="talend-vault-sidecar-injector" \
      com.talend.service="talend-vault-sidecar-injector" \
      com.talend.description="Kubernetes Webhook Admission Server for Vault sidecar injection"

COPY target/vaultinjector-webhook ${TALEND_HOME}/webhook/vaultinjector-webhook

ENTRYPOINT ["/opt/talend/webhook/vaultinjector-webhook"]
