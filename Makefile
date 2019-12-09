RELEASE_VERSION:=5.1.0
VSI_VERSION:=5.0.0

OWNER:=Talend
REPO:=vault-sidecar-injector
TARGET:=target/vaultinjector-webhook
SRC:=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Inject VSI version into code at build time
LDFLAGS=-ldflags "-X=main.VERSION=$(VSI_VERSION)"

.SILENT: ;  	# No need for @
.ONESHELL: ; 	# Single shell for a target (required to properly use all of our local variables)
.PHONY: all clean fmt test build release image
.DEFAULT_GOAL := build

all: release

clean:
	rm -f target/*

fmt:
	gofmt -l -w $(SRC)

test:
	echo "Running tests ..."
	go test -v ./...

build: clean test
	echo "Building ..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -a -o $(TARGET)
	cd target && sha512sum vaultinjector-webhook > vaultinjector-webhook.sha512

package:
	mkdir -p target && cd target
	echo "Archive Helm chart ..."
	mkdir -p vault-sidecar-injector && cp -R ../README.md ../deploy/helm/* ./vault-sidecar-injector
	helm package vault-sidecar-injector
	rm -R vault-sidecar-injector
	helm lint ./vault-sidecar-injector-*.tgz --debug

image:
	echo "Build image from sources ..."
	docker build -t talend/vault-sidecar-injector:${VSI_VERSION} .

image-from-build: build
	echo "Build image from local build ..."
	docker build -f Dockerfile.local -t talend/vault-sidecar-injector:${VSI_VERSION} .

release: image-from-build package
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
