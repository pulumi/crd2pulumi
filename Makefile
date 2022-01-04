PROJECT      := github.com/pulumi/crd2pulumi
VERSION      ?= $(shell pulumictl get version)
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
