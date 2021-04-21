PROJECT          := github.com/pulumi/crd2pulumi

GO              ?= go

ensure::
	$(GO) mod download

build::
	$(GO) build $(PROJECT)
