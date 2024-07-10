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
	"regexp"

	"github.com/pulumi/crd2pulumi/internal/versions"
	javaGen "github.com/pulumi/pulumi-java/pkg/codegen/java"
)

func GenerateJava(pg *PackageGenerator, name string) (map[string]*bytes.Buffer, error) {
	pkg := pg.SchemaPackage(true)

	// These fields are required for the Java code generation
	pkg.Description = "Generated Java SDK via crd2pulumi"
	pkg.Repository = "Placeholder"

	// Set up packages
	packages := map[string]string{}
	for _, groupVersion := range pg.GroupVersions {
		group, version, err := versions.SplitGroupVersion(groupVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		groupPrefix, err := versions.GroupPrefix(group)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		packages[groupVersion] = groupPrefix + "." + version
	}
	packages["meta/v1"] = "meta.v1"

	oldName := pkg.Name
	pkg.Name = name

	files, err := javaGen.GeneratePackage("crd2pulumi", pkg, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate Java package: %w", err)
	}

	pkg.Name = oldName

	// Pin the kubernetes provider version used
	utilsPath := "src/main/java/com/pulumi/" + name + "/Utilities.java"
	utils, ok := files[utilsPath]
	if !ok {
		return nil, fmt.Errorf("cannot find generated utilities.ts")
	}
	re := regexp.MustCompile(`static \{(?:[^{}]|{[^{}]*})*}`)
	files[utilsPath] = []byte(re.ReplaceAllString(string(utils), `static {
    	version = "4.9.0";
	}`))

	unneededJavaFiles := []string{
		"src/main/java/com/pulumi/" + name + "/Provider.java",
		"src/main/java/com/pulumi/" + name + "/ProviderArgs.java",
		"src/main/java/com/pulumi/kubernetes/meta/v1/inputs/ObjectMetaArgs.java",
		"src/main/java/com/pulumi/kubernetes/meta/v1/outputs/ObjectMeta.java",
	}

	// Remove unneeded files
	for _, unneededFile := range unneededJavaFiles {
		delete(files, unneededFile)
	}

	buffers := map[string]*bytes.Buffer{}
	for name, code := range files {
		buffers[name] = bytes.NewBuffer(code)
	}

	return buffers, err
}
