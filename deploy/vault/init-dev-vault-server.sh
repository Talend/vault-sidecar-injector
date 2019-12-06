#!/bin/bash

set -e

VAULT_POD="kubectl exec -i vault-0 -- sh -c"

# Create policies to allow read access to our secrets
cat vault-server-policy-test.hcl | ${VAULT_POD} "VAULT_TOKEN=root vault policy write test_pol -"
cat vault-server-policy-test2.hcl | ${VAULT_POD} "VAULT_TOKEN=root vault policy write test_pol2 -"

# Enable KV v1 (KV v2 is enabled with Vault server in dev mode whereas KV v1 is enabled in prod mode: https://www.vaultproject.io/docs/secrets/kv/kv-v2.html#setup)
${VAULT_POD} "VAULT_TOKEN=root vault secrets disable secret/"
${VAULT_POD} "VAULT_TOKEN=root vault secrets enable -version=1 -path=secret kv"

# Enable Transit Secrets Engine and create a test key
${VAULT_POD} "VAULT_TOKEN=root vault secrets enable transit" || true
${VAULT_POD} "VAULT_TOKEN=root vault write -f transit/keys/test-key"

# Enable Vault K8S Auth Method
echo "-> Enable & set up Vault Kubernetes Auth Method"
${VAULT_POD} "VAULT_TOKEN=root vault auth enable kubernetes" || true

# Config Vault K8S Auth Method
export VAULT_SA_NAME=$(kubectl get sa vault -o jsonpath="{.secrets[*]['name']}")
export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -o jsonpath="{.data.token}" | base64 --decode; echo)
export SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)

${VAULT_POD} "VAULT_TOKEN=root vault write auth/kubernetes/config kubernetes_host=\"https://kubernetes:443\" kubernetes_ca_cert=\"$SA_CA_CRT\" token_reviewer_jwt=\"$SA_JWT_TOKEN\""

# Create roles for Vault K8S Auth Method
${VAULT_POD} "VAULT_TOKEN=root vault write auth/kubernetes/role/test bound_service_account_names=default,job-sa bound_service_account_namespaces=default policies=test_pol ttl=5m"
${VAULT_POD} "VAULT_TOKEN=root vault write auth/kubernetes/role/test2 bound_service_account_names=default,job-sa bound_service_account_namespaces=default policies=test_pol2 ttl=5m"

# Enable Vault AppRole Auth Method
echo "-> Enable & set up Vault AppRole Auth Method"
${VAULT_POD} "VAULT_TOKEN=root vault auth enable approle" || true

# Create roles for Vault AppRole Auth Method
${VAULT_POD} "VAULT_TOKEN=root vault write auth/approle/role/test secret_id_ttl=60m token_num_uses=10 token_ttl=20m token_max_ttl=30m secret_id_num_uses=0 policies=test_pol"
${VAULT_POD} "VAULT_TOKEN=root vault write auth/approle/role/test2 secret_id_ttl=60m token_num_uses=10 token_ttl=20m token_max_ttl=30m secret_id_num_uses=0 policies=test_pol2"

# Add some secrets
${VAULT_POD} "VAULT_TOKEN=root vault kv put secret/test/test-app-svc ttl=10s SECRET1=Batman SECRET2=BruceWayne"
${VAULT_POD} "VAULT_TOKEN=root vault kv put secret/test2/test-app2-svc ttl=5s SECRET1=my SECRET2=name SECRET3=is SECRET4=James"

# List Auth Methods and Secrets Engines
echo
echo "Auth Methods"
echo "============"
${VAULT_POD} "VAULT_TOKEN=root vault auth list -detailed"
echo
echo "Secrets Engines"
echo "==============="
${VAULT_POD} "VAULT_TOKEN=root vault secrets list -detailed"