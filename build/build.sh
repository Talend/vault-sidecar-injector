#!/bin/bash

set -e

SCRIPT_PATH="$(dirname "$0")"
cd "$SCRIPT_PATH"/..

echo "Building Talend Webhook Admission Server for Vault sidecar injection ..."
docker run --rm -v "${PWD}":/vaultinjector-webhook -w /vaultinjector-webhook golang:1.12.9 sh -c "make all; chmod -R a+w target"

echo "Building Docker image..."
docker build -t talend/common/tsbi/k8s/vault-sidecar-injector .
