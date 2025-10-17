package tests

import (
	"io"
	"strings"
	"testing"

	"github.com/pulumi/crd2pulumi/pkg/codegen"
)

func TestUnneededGoFiles(t *testing.T) {
	mockCRDYaml := `---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: uneeded-go-files-test.pulumi.com
  names:
    plural: testresources
    singular: testresource
    kind: TestResource
  scope: Namespaced
  versions:
    - test:
      storage: true
      served: true
      name: test
      schema:
        openAPIV3Schema:
          properties:
            testProperty:
              type: string`

	yamlSources := []io.ReadCloser{
		io.NopCloser(strings.NewReader(mockCRDYaml)),
	}

	// Invoke ReadPackagesFromSource
	pg, err := codegen.ReadPackagesFromSource("", yamlSources)
	if err != nil {
		t.Fatalf("ReadPackagesFromSource failed: %v", err)
	}

	// Pick a generated file we want to exclude
	unNeededGoFile := "unneededgofilestest/test/testResource.go"
	codegen.UnneededGoFiles.Add(unNeededGoFile)

	// Generate the code from the mocked CRD
	goSettings := &codegen.CodegenSettings{
		Language:    "go",
		PackageName: "crds",
	}
	buffers, err := codegen.GenerateGo(pg, goSettings)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	// Assert that buffers do not contain unneeded file
	if _, exists := buffers["../kubernetes/"+unNeededGoFile]; exists {
		t.Errorf("Unneeded GO file was not excluded by GoGenerate, %s", unNeededGoFile)
	}
}
