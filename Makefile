SHELL=/bin/bash

RELEASE_VERSION:=$(shell cat VERSION_RELEASE)
VSI_VERSION:=$(shell cat VERSION_VSI)
CHART_VERSION:=$(shell cat VERSION_CHART)

OWNER:=Talend
REPO:=vault-sidecar-injector
TARGET_WEBHOOK:=target/vaultinjector-webhook
TARGET_ENV:=target/vaultinjector-env

# Inject VSI version into code at build time
LDFLAGS=-ldflags "-X=main.VERSION=$(VSI_VERSION)"

.SILENT: ;  	# No need for @
.ONESHELL: ; 	# Single shell for a target (required to properly use local variables)
.PHONY: all clean test build-vsi-webhook build-vsi-env build package image image-from-build release
.DEFAULT_GOAL := build

all: release

clean:
	rm -f target/*

test: # for detailed outputs, run 'make test VERBOSE=true'
	if [ -z ${OFFLINE} ] || [ ${OFFLINE} != true ];then \
		echo "Running tests ..."; \
		echo ">> for detailed outputs, run 'make test VERBOSE=true' <<"; \
		go test -mod=mod -v ./...; \
	else \
		echo "Running tests using local vendor folder (ie offline build) ..."; \
		echo ">> for detailed outputs, run 'make test VERBOSE=true' <<"; \
		go test -mod=vendor -v ./...; \
	fi

build-vsi-webhook: clean test # run 'make build-vsi-webhook OFFLINE=true' to build from vendor folder
	if [ -z ${OFFLINE} ] || [ ${OFFLINE} != true ];then \
		echo "Building vsi-webhook ..."; \
		GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -mod=mod -a -o $(TARGET_WEBHOOK) ./cmd/vaultinjector-webhook; \
	else \
		echo "Building vsi-webhook using local vendor folder (ie offline build) ..."; \
		GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -mod=vendor -a -o $(TARGET_WEBHOOK) ./cmd/vaultinjector-webhook; \
	fi
	cd target && sha512sum vaultinjector-webhook > vaultinjector-webhook.sha512

build-vsi-env: # run 'make build-vsi-env OFFLINE=true' to build from vendor folder
	if [ -z ${OFFLINE} ] || [ ${OFFLINE} != true ];then \
		echo "Building vsi-env ..."; \
		GOOS=linux GOARCH=amd64 go build -mod=mod -a -o $(TARGET_ENV) ./cmd/vaultinjector-env; \
	else \
		echo "Building vsi-env using local vendor folder (ie offline build) ..."; \
		GOOS=linux GOARCH=amd64 go build -mod=vendor -a -o $(TARGET_ENV) ./cmd/vaultinjector-env; \
	fi
	cd target && sha512sum vaultinjector-env > vaultinjector-env.sha512

build: build-vsi-webhook build-vsi-env

package:
	set -e
	mkdir -p target && cd target
	echo "Archive Helm chart ..."
	mkdir -p vault-sidecar-injector && cp -R ../README.md ../deploy/helm/* ./vault-sidecar-injector
	sed -i "s/version: 0.0.0/version: ${CHART_VERSION}/;s/appVersion: 0.0.0/appVersion: ${VSI_VERSION}/" ./vault-sidecar-injector/Chart.yaml
	sed -i "s/tag: \"latest\"  # VSI image tag/tag: \"${VSI_VERSION}\"  # VSI image tag/" ./vault-sidecar-injector/values.yaml
	helm package vault-sidecar-injector
	rm -R vault-sidecar-injector
	helm lint ./vault-sidecar-injector-*.tgz --debug

image:
	echo "Build image using Go container and multi-stage build ..."
	docker build -t talend/vault-sidecar-injector:${VSI_VERSION} .
	docker tag talend/vault-sidecar-injector:${VSI_VERSION} talend/vault-sidecar-injector

image-from-build: build
	echo "Build image from local build ..."
	docker build -f Dockerfile.local -t talend/vault-sidecar-injector:${VSI_VERSION} .
	docker tag talend/vault-sidecar-injector:${VSI_VERSION} talend/vault-sidecar-injector

release: image-from-build package
	read -p "Publish image on Docker Hub (y/n)? " answer
	case $$answer in \
	y|Y ) \
		docker login; \
		docker push talend/vault-sidecar-injector:${VSI_VERSION}; \
		if [ "$$?" -ne 0 ]; then \
			echo "Unable to publish image"; \
			exit 1; \
		fi; \
	;; \
	* ) \
		echo "Image not published on Docker Hub"; \
	;; \
	esac
	cd target
	echo "Releasing artifacts ..."
	read -p "- Github user name to use for release: " username
	echo "- Creating release"
	id=$$(curl -u $$username -s -X POST "https://api.github.com/repos/${OWNER}/${REPO}/releases" -d '{"tag_name": "v'${RELEASE_VERSION}'", "name": "v'${RELEASE_VERSION}'", "draft": true, "body": ""}' | jq '.id')
	if [ "$$?" -ne 0 ]; then \
		echo "Unable to create release"; \
		echo $$id; \
		exit 1; \
	fi
	echo "- Release id=$$id"
	echo
	echo "- Publishing release binary"
	for asset_file in $(shell ls ./target); do \
		asset_absolute_path=$$(realpath $$asset_file); \
		echo "Adding file $$asset_absolute_path"; \
		echo; \
		asset_filename=$$(basename $$asset_absolute_path); \
		curl -u $$username -s --data-binary @"$$asset_absolute_path" -H "Content-Type: application/octet-stream" "https://uploads.github.com/repos/${OWNER}/${REPO}/releases/$$id/assets?name=$$asset_filename"; \
		if [ "$$?" -ne 0 ]; then \
			echo "Unable to publish binary $$asset_absolute_path"; \
			exit 1; \
		fi; \
		echo; \
	done
	echo
	echo
	read -p "- Confirm release ok at https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id (y/[n])? " answer
	case $$answer in \
	y|Y ) \
		curl -u $$username -s -X PATCH "https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id" -d '{"draft": false}'; \
		if [ "$$?" -ne 0 ]; then \
			echo "Unable to finish release"; \
			exit 1; \
		fi; \
	;; \
	* ) \
		curl -u $$username -s -X DELETE "https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id"; \
		echo "Aborted"; \
	;; \
	esac
