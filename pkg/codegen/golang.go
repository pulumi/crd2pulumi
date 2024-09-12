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

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	goGen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
)

var UnneededGoFiles = codegen.NewStringSet(
	// The root directory doesn't define any resources:
	"doc.go",
	"init.go",
	"provider.go",

	// We use the standard Kubernetes meta/v1 types, so skip generating them:
	"meta/v1/pulumiTypes.go",

	// No need to generate these, they are imported from pulumi-kubernetes directly:
	"utilities/pulumiUtilities.go",
	"utilities/pulumiVersion.go",
)

func GenerateGo(pg *PackageGenerator, name string) (buffers map[string]*bytes.Buffer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	pkg := pg.SchemaPackageWithObjectMetaType()
	langName := "go"
	oldName := pkg.Name
	pkg.Name = name
	moduleToPackage, err := pg.ModuleToPackage()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	moduleToPackage["meta/v1"] = "meta/v1"

	files, err := goGen.GeneratePackage("crd2pulumi", pkg, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate Go package: %w", err)
	}

	pkg.Name = oldName
	delete(pkg.Language, langName)

	buffers = map[string]*bytes.Buffer{}
	for path, code := range files {
		if !UnneededGoFiles.Has(path) {
			buffers[path] = bytes.NewBuffer(code)
		}
	}

	return buffers, err
}
