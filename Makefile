PROJECT      := github.com/pulumi/crd2pulumi
VERSION      ?= $(shell (command -v pulumictl && pulumictl get version || echo "0.0.0-dev"))
VERSION_PATH := gen.Version

GO              ?= go

all:: ensure build test

ensure::
	$(GO) mod download

build::
	go build -o bin/crd2pulumi -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" $(PROJECT)

install::
	go install -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}"

test::
	$(GO) test -v ./tests/
