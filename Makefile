PROJECT      := github.com/pulumi/crd2pulumi
VERSION      ?= $(shell (command -v pulumictl > /dev/null && pulumictl get version || echo "0.0.0-dev"))
VERSION_PATH := cmd.Version

GO              ?= go

all:: ensure build test

ensure::
	$(GO) mod download

build::
	go build -o bin/crd2pulumi -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" $(PROJECT)

install::
	go install -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}"

test::
	$(GO) test -v -coverprofile="coverage.txt" -covermode=atomic -coverpkg=./... ./...
