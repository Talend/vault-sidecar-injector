VERSION:=3.0.0

TARGET:=target/vaultinjector-webhook
SRC:=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Inject version into code at build time
LDFLAGS=-ldflags "-X=main.VERSION=$(VERSION)"

.SILENT: ;  # No need for @
.PHONY: all clean fmt test build
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
