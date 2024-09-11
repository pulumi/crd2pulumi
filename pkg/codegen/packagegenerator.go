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
	"fmt"
	"io"

	"github.com/pulumi/crd2pulumi/internal/unstruct"
	"github.com/pulumi/crd2pulumi/internal/versions"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// PackageGenerator generates code for multiple CustomResources
type PackageGenerator struct {
	// CustomResourceGenerators contains a slice of all CustomResourceGenerators
	CustomResourceGenerators []CustomResourceGenerator
	// ResourceTokens is a slice of the token types of every CustomResource
	ResourceTokens []string
	// GroupVersions is a slice of the names of every CustomResource's versions,
	// in the format <group>/<version>
	GroupVersions []string
	// Types is a mapping from every type's token name to its ComplexTypeSpec
	Types map[string]pschema.ComplexTypeSpec
	// Version is the semver that will be stamped into the generated package
	Version string
	// schemaPackage is the Pulumi schema package used to generate code for
	// languages that do not need an ObjectMeta type (NodeJS)
	schemaPackage *pschema.Package
	// schemaPackageWithObjectMetaType is the Pulumi schema package used to
	// generate code for languages that need an ObjectMeta type (Python, Go, and .NET)
	schemaPackageWithObjectMetaType *pschema.Package
}

// ReadPackagesFromSource reads one or more documents and returns a PackageGenerator that can be used to generate Pulumi code.
// Calling this function will fully read and close each document.
func ReadPackagesFromSource(version string, yamlSources []io.ReadCloser) (*PackageGenerator, error) {
	yamlData := make([][]byte, len(yamlSources))

	for i, yamlSource := range yamlSources {
		defer yamlSource.Close()
		var err error
		yamlData[i], err = io.ReadAll(yamlSource)
		if err != nil {
			return nil, fmt.Errorf("failed to read YAML: %w", err)
		}
	}

	crds, err := unstruct.UnmarshalYamls(yamlData)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml file(s): %w", err)
	}

	if len(crds) == 0 {
		return nil, fmt.Errorf("could not find any CRD YAML files")
	}

	resourceTokensSize := 0
	groupVersionsSize := 0

	crgs := make([]CustomResourceGenerator, 0, len(crds))
	for i, crd := range crds {
		crg, err := NewCustomResourceGenerator(crd)
		if err != nil {
			return nil, fmt.Errorf("could not parse crd %d: %w", i, err)
		}
		resourceTokensSize += len(crg.ResourceTokens)
		groupVersionsSize += len(crg.GroupVersions)
		crgs = append(crgs, crg)
	}

	baseRefs := make([]string, 0, resourceTokensSize)
	groupVersions := make([]string, 0, groupVersionsSize)
	for _, crg := range crgs {
		baseRefs = append(baseRefs, crg.ResourceTokens...)
		groupVersions = append(groupVersions, crg.GroupVersions...)
	}

	pg := &PackageGenerator{
		CustomResourceGenerators: crgs,
		ResourceTokens:           baseRefs,
		GroupVersions:            groupVersions,
		Version:                  version,
	}
	return pg, nil
}

// SchemaPackage returns the Pulumi schema package with no ObjectMeta type.
// This is only necessary for NodeJS and Python.
func (pg *PackageGenerator) SchemaPackage() *pschema.Package {
	if pg.schemaPackage == nil {
		pkg, err := genPackage(pg.Version, pg.CustomResourceGenerators, false)
		contract.AssertNoErrorf(err, "could not parse Pulumi package")
		pg.schemaPackage = pkg
	}
	return pg.schemaPackage
}

// SchemaPackageWithObjectMetaType returns the Pulumi schema package with
// an ObjectMeta type. This is only necessary for Go and .NET.
func (pg *PackageGenerator) SchemaPackageWithObjectMetaType() *pschema.Package {
	if pg.schemaPackageWithObjectMetaType == nil {
		pkg, err := genPackage(pg.Version, pg.CustomResourceGenerators, true)
		contract.AssertNoErrorf(err, "could not parse Pulumi package")
		pg.schemaPackageWithObjectMetaType = pkg
	}
	return pg.schemaPackageWithObjectMetaType
}

// Returns language-specific 'ModuleToPackage' map. Creates a mapping from
// every groupVersion string <group>/<version> to <groupPrefix>/<version>.
func (pg *PackageGenerator) ModuleToPackage() (map[string]string, error) {
	moduleToPackage := map[string]string{}
	for _, groupVersion := range pg.GroupVersions {
		group, version, err := versions.SplitGroupVersion(groupVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		prefix, err := versions.GroupPrefix(group)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		moduleToPackage[groupVersion] = prefix + "/" + version
	}
	return moduleToPackage, nil
}

// HasSchemas returns true if there exists at least one CustomResource with a schema in this package.
func (pg *PackageGenerator) HasSchemas() bool {
	for _, crg := range pg.CustomResourceGenerators {
		if crg.HasSchemas() {
			return true
		}
	}
	return false
}
