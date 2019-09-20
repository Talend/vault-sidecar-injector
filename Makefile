VERSION:=3.0.0

OWNER=Talend
REPO=vault-sidecar-injector
ARTIFACT=vaultinjector-webhook
TARGET:=target/$(ARTIFACT)
SRC:=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Inject version into code at build time
LDFLAGS=-ldflags "-X=main.VERSION=$(VERSION)"

.SILENT: ;  	# No need for @
.ONESHELL: ; 	# Single shell for a target (required to properly use all of our local variables)
.PHONY: all clean fmt test build release image
.DEFAULT_GOAL := build

all: build

clean:
	rm -f $(TARGET)*

fmt:
	gofmt -l -w $(SRC)

test:
	echo "Running tests ..."
	go test -v ./...

build: clean test
	echo "Building ..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -a -o $(TARGET)

release: build
	echo "Releasing artifacts ..."
	read -p "- Github user name to use for release: " username
	echo "- Creating release"
	id=$$(curl -u $$username -X POST "https://api.github.com/repos/${OWNER}/${REPO}/releases" -d '{"tag_name": "talend-vault-sidecar-injector-'${VERSION}'", "name": "talend-vault-sidecar-injector-'${VERSION}'", "draft": true, "body": ""}' | jq '.id')
	if [ "$$?" -ne 0 ]; then \
		echo "Unable to create release"; \
		echo $$id; \
		exit 1; \
	fi
	echo
	echo "- Publishing release binary"
	asset_absolute_path=$$(realpath ${TARGET})
	asset_filename=$$(basename $$asset_absolute_path)
	curl -u $$username --data-binary @"$$asset_absolute_path" -H "Content-Type: application/octet-stream" "https://uploads.github.com/repos/${OWNER}/${REPO}/releases/$$id/assets?name=$$asset_filename"
	if [ "$$?" -ne 0 ]; then \
		echo "Unable to publish binary $$asset_absolute_path"; \
		exit 1; \
	fi
	echo
	echo
	read -p "- Confirm release ok at https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id (y/[n])? " answer
	case $$answer in \
	y|Y ) \
		curl -u $$username -X PATCH "https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id" -d '{"draft": false}'; \
		if [ "$$?" -ne 0 ]; then \
			echo "Unable to finish release"; \
			exit 1; \
		fi; \
	;; \
	* ) \
		curl -u $$username -X DELETE "https://api.github.com/repos/${OWNER}/${REPO}/releases/$$id"; \
		echo "Aborted"; \
	;; \
	esac

image: build
	echo "Build image ..."
	docker build -t talend/common/tsbi/k8s/vault-sidecar-injector .
