PROJECT          := github.com/pulumi/crd2pulumi

GO              ?= go
GOMODULE = GO111MODULE=on

ensure::
	$(GOMODULE) $(GO) mod tidy

build::
	$(GOMODULE) $(GO) build $(PROJECT)
