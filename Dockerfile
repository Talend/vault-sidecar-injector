FROM golang:1.14.4 AS buildTVSI

COPY . /vaultsidecarinjector
RUN cd /vaultsidecarinjector && make build

FROM centos:7.7.1908

USER root

# Talend home, Talend user/user group/user id
ENV TALEND_HOME=/opt/talend
ENV TALEND_USER=talend
ENV TALEND_USERGROUP=$TALEND_USER
ENV TALEND_UID=61000

# Update CentOS (only with security patches)
RUN set -x \
    && yum -y update --security \
    && yum clean all \
    && rm -rf /var/cache/yum

# Create non-root user $TALEND_USER
RUN set -x \
    && mkdir -p $TALEND_HOME \
    && groupadd -r $TALEND_USERGROUP -g $TALEND_UID \
    && useradd -l -u $TALEND_UID -r -g $TALEND_USERGROUP -m -d /home/$TALEND_USER -s /sbin/nologin $TALEND_USER \
    && chmod 755 /home/$TALEND_USER \
    && chmod -R "g+rwX" $TALEND_HOME \
    && chown -R $TALEND_USER:$TALEND_USERGROUP $TALEND_HOME

WORKDIR $TALEND_HOME
USER $TALEND_UID

LABEL com.talend.maintainer="Talend <support@talend.com>" \
      com.talend.url="https://www.talend.com/" \
      com.talend.vendor="Talend" \
      com.talend.name="Vault Sidecar Injector" \
      com.talend.application="talend-vault-sidecar-injector" \
      com.talend.service="talend-vault-sidecar-injector" \
      com.talend.description="Kubernetes Webhook Admission Server for Vault sidecar injection"

COPY --chown=talend:talend --from=buildTVSI /vaultsidecarinjector/target/vaultinjector-webhook ${TALEND_HOME}/webhook/

ENTRYPOINT ["/opt/talend/webhook/vaultinjector-webhook"]
