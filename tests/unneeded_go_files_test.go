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
	uneededGoFile := "uneededgofilestest/test/testResource.go"
	codegen.UnneededGoFiles.Add(uneededGoFile)

	// Generate the code from the mocked CRD
	buffers, err := codegen.GenerateGo(pg, "crds")
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	// Assert that buffers do not contain unneeded file
	if _, exists := buffers["../kubernetes/"+uneededGoFile]; exists {
		t.Errorf("Uneeded GO file was not excluded by GoGenerate, %s", uneededGoFile)
	}
}
