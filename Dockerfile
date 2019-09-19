FROM artifactory.datapwn.com/tlnd-docker-prod/talend/common/tsbi/centos-base:1.2.3-20190718150644

LABEL com.talend.name="Talend Vault Sidecar Injector" \
      com.talend.application="talend-vault-sidecar-injector" \
      com.talend.service="talend-vault-sidecar-injector" \
      com.talend.description="Kubernetes Webhook Admission Server for Vault sidecar injection"

COPY --chown=talend:talend target/vaultinjector-webhook ${TALEND_HOME}/webhook/vaultinjector-webhook

ENTRYPOINT ["/opt/talend/webhook/vaultinjector-webhook"]
