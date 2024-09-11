// Copyright 2016-2020, Pulumi Corporation.
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
	"strings"

	"github.com/go-openapi/jsonreference"
	"github.com/pulumi/pulumi/pkg/v3/codegen/cgstrings"
	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

const (
	definitionPrefix = "#/definitions/"
)

// CustomResourceGenerator generates a Pulumi schema for a single CustomResource
type CustomResourceGenerator struct {
	// CustomResourceDefinition contains the unmarshalled CRD YAML
	CustomResourceDefinition extensionv1.CustomResourceDefinition
	// Schemas represents a mapping from each version in the `spec.versions`
	// list to its corresponding `openAPIV3Schema` field in the CRD YAML
	Schemas map[string]spec.Swagger
	// ApiVersion represents the `apiVersion` field in the CRD YAML
	APIVersion string
	// Kind represents the `spec.names.kind` field in the CRD YAML
	Kind string
	// Plural represents the `spec.names.plural` field in the CRD YAML
	Plural string
	// Group represents the `spec.group` field in the CRD YAML
	Group string
	// Versions is a slice of names of each version supported by this CRD
	Versions []string
	// GroupVersions is a slice of names of each version, in the format
	// <group>/<version>.
	GroupVersions []string
	// ResourceTokens is a slice of the token types of every versioned
	// CustomResource
	ResourceTokens []string
}

// flattenOpenAPI recursively finds all nested objects in the OpenAPI spec and flattens them into a single object as definitions.
func flattenOpenAPI(sw *spec.Swagger) error {
	// Create a stack of definition names to be processed.
	definitionStack := make([]string, 0, len(sw.Definitions))

	// Populate existing definitions into the stack.
	for defName := range sw.Definitions {
		definitionStack = append(definitionStack, defName)
	}

	for len(definitionStack) != 0 {
		// Pop the last definition from the stack.
		definitionName := definitionStack[len(definitionStack)-1]
		definitionStack = definitionStack[:len(definitionStack)-1]
		// Get the definition from the OpenAPI spec.
		definition := sw.Definitions[definitionName]

		for propertyName, propertySchema := range definition.Properties {
			// If the property is already a reference to a URL, we can skip it.
			if propertySchema.Ref.GetURL() != nil {
				continue
			}

			// If the property is not an object or array, we can skip it.
			if !propertySchema.Type.Contains("object") {
				continue
			}

			if propertySchema.Properties == nil && propertySchema.Items == nil {
				continue
			}

			// If the property is an object with additional properties, we can skip it. We only care about
			// nested objects that are explicitly defined.
			if propertySchema.AdditionalProperties != nil {
				continue
			}

			// if propertySchema.Items != nil {
			// 	currNode := propertySchema.Items.Schema
			// 	flattenOpenAPIItems(sw, &definitionStack, definitionName, currNode)
			// 	continue
			// }

			// Create a new definition for the nested object by joining the parent definition name and the property name.
			// This is to ensure that the nested object is unique and does not conflict with other definitions.
			nestedDefinitionName := definitionName + cgstrings.UppercaseFirst(propertyName)
			sw.Definitions[nestedDefinitionName] = propertySchema
			// Add nested object to the stack to be recursively flattened.
			definitionStack = append(definitionStack, nestedDefinitionName)

			// Reset the property to be a reference to the nested object.
			refName := definitionPrefix + nestedDefinitionName
			ref, err := jsonreference.New(refName)
			if err != nil {
				return fmt.Errorf("error creating OpenAPI json reference for nested object: %w", err)
			}

			definition.Properties[propertyName] = spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.Ref{
						Ref: ref,
					},
				},
			}
		}
	}
	return nil
}

// crdToOpenAPI generates the OpenAPI specs for a given CRD manifest.
func crdToOpenAPI(crd *extensionv1.CustomResourceDefinition) ([]*spec.Swagger, error) {
	var openAPIManifests []*spec.Swagger

	setCRDDefaults(crd)

	for _, v := range crd.Spec.Versions {
		if !v.Served {
			continue
		}
		// Defaults are not pruned here, but before being served.
		sw, err := builder.BuildOpenAPIV2(crd, v.Name, builder.Options{V2: true, StripValueValidation: false, StripNullable: false, AllowNonStructural: true})
		if err != nil {
			return nil, err
		}

		err = flattenOpenAPI(sw)
		if err != nil {
			return nil, fmt.Errorf("error flattening OpenAPI spec: %w", err)
		}

		openAPIManifests = append(openAPIManifests, sw)
	}

	return openAPIManifests, nil
}

// fillDefaultNames sets the default names for the CRD if they are not specified.
// This allows the OpenAPI builder to generate the swagger specs correctly with
// the correct defaults.
func setCRDDefaults(crd *extensionv1.CustomResourceDefinition) {
	if crd.Spec.Names.Singular == "" {
		crd.Spec.Names.Singular = strings.ToLower(crd.Spec.Names.Kind)
	}
	if crd.Spec.Names.ListKind == "" {
		crd.Spec.Names.ListKind = crd.Spec.Names.Kind + "List"
	}
}

func NewCustomResourceGenerator(crd extensionv1.CustomResourceDefinition) (CustomResourceGenerator, error) {
	apiVersion := crd.APIVersion
	schemas := map[string]spec.Swagger{}

	swagger, err := crdToOpenAPI(&crd)
	if err != nil {
		return CustomResourceGenerator{}, fmt.Errorf("could not generate OpenAPI spec for CRD: %w", err)
	}

	for _, sw := range swagger {
		schemas[sw.Info.Version] = *sw
	}

	kind := crd.Spec.Names.Kind
	plural := crd.Spec.Names.Plural
	group := crd.Spec.Group

	versions := make([]string, 0, len(schemas))
	groupVersions := make([]string, 0, len(schemas))
	resourceTokens := make([]string, 0, len(schemas))
	for version := range schemas {
		versions = append(versions, version)
		groupVersions = append(groupVersions, group+"/"+version)
		resourceTokens = append(resourceTokens, getToken(group, version, kind))
	}

	crg := CustomResourceGenerator{
		CustomResourceDefinition: crd,
		Schemas:                  schemas,
		APIVersion:               apiVersion,
		Kind:                     kind,
		Plural:                   plural,
		Group:                    group,
		Versions:                 versions,
		GroupVersions:            groupVersions,
		ResourceTokens:           resourceTokens,
	}

	return crg, nil
}

// HasSchemas returns true if the CustomResource specifies at least some schema, and false otherwise.
func (crg *CustomResourceGenerator) HasSchemas() bool {
	return len(crg.Schemas) > 0
}
