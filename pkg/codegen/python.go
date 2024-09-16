// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codegen

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

const pythonMetaFile = `from pulumi_kubernetes.meta.v1._inputs import *
import pulumi_kubernetes.meta.v1.outputs
`

func GeneratePython(pg *PackageGenerator, name string) (map[string]*bytes.Buffer, error) {
	pkg := pg.SchemaPackageWithObjectMetaType()

	langName := "python"
	oldName := pkg.Name
	pkg.Name = name

	files, err := python.GeneratePackage(PulumiToolName, pkg, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate Go package: %w", err)
	}

	pkg.Name = oldName
	delete(pkg.Language, langName)

	pythonPackageDir := "pulumi_" + name

	// Remove unneeded files
	var unneededPythonFiles = []string{
		filepath.Join(pythonPackageDir, "README.md"),
	}
	for _, unneededFile := range unneededPythonFiles {
		delete(files, unneededFile)
	}

	// Import the actual SDK ObjectMeta types in place of our placeholder ones
	if pg.HasSchemas() {
		metaPath := filepath.Join(pythonPackageDir, "meta/v1", "__init__.py")
		code, ok := files[metaPath]
		contract.Assertf(ok, "missing meta/v1/__init__.py file")
		files[metaPath] = append(code, []byte(pythonMetaFile)...)
	}

	buffers := map[string]*bytes.Buffer{}
	for name, code := range files {
		if name == "pyproject.toml" {
			code = bytes.ReplaceAll(code, []byte(`0.0.0+dev`), []byte(KubernetesProviderVersion))
		}
		buffers[name] = bytes.NewBuffer(code)
	}
	return buffers, nil
}
