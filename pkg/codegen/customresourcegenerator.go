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
	initialDefinitions := make([]string, 0, len(sw.Definitions))

	for defName := range sw.Definitions {
		initialDefinitions = append(initialDefinitions, defName)
	}

	for _, defName := range initialDefinitions {
		definition := sw.Definitions[defName]
		spec, err := flattenRecursively(sw, defName, definition)
		if err != nil {
			return fmt.Errorf("error flattening OpenAPI spec: %w", err)
		}

		sw.Definitions[defName] = spec
	}

	return nil
}

func flattenRecursively(sw *spec.Swagger, parentName string, currSpec spec.Schema) (spec.Schema, error) {
	// If at bottom of the stack, return the spec.
	if currSpec.Properties == nil && currSpec.Items == nil && currSpec.AdditionalProperties == nil {
		return currSpec, nil
	}

	// If the spec already has a reference to a URL, we can skip it.
	if currSpec.Ref.GetURL() != nil {
		return currSpec, nil
	}

	// If the property is an object with additional properties, we can skip it if it is not an array of inline objects. We only care about
	// nested objects that are explicitly defined.
	if currSpec.AdditionalProperties != nil {
		// Not an array of inline objects, so we can skip processing.
		if currSpec.AdditionalProperties.Schema.Items == nil {
			return currSpec, nil
		}

		// Property is an array of inline objects.
		s, ref, err := flattedArrayObject(parentName, sw, currSpec.AdditionalProperties.Schema.Items.Schema)
		if err != nil {
			return currSpec, fmt.Errorf("error flattening OpenAPI object of array property: %w", err)
		}

		currSpec.AdditionalProperties.Schema.Items.Schema = &s

		if ref != nil {
			currSpec.AdditionalProperties.Schema.Items.Schema.Ref = spec.Ref{Ref: *ref}
			currSpec.AdditionalProperties.Schema.Items.Schema.Type = nil
			currSpec.AdditionalProperties.Schema.Items.Schema.Properties = nil
		}

		return currSpec, nil
	}

	// If the property is an array, we need to remove any inline objects and replace them with references.
	if currSpec.Items != nil {
		if currSpec.Items.Schema == nil {
			return currSpec, fmt.Errorf("error flattening OpenAPI spec: items schema is nil")
		}

		s, ref, err := flattedArrayObject(parentName, sw, currSpec.Items.Schema)
		if err != nil {
			return currSpec, fmt.Errorf("error flattening OpenAPI array property: %w", err)
		}

		currSpec.Items.Schema = &s

		if ref != nil {
			currSpec.Items.Schema.Ref = spec.Ref{Ref: *ref}
			currSpec.Items.Schema.Type = nil
			currSpec.Items.Schema.Properties = nil
		}

		return currSpec, nil
	}

	// Recurse through the properties of the object.
	for nestedPropertyName, nestedProperty := range currSpec.Properties {
		// VistoriaMetrics has some weird fields - likely a typegen issue on their end, so let's skip them.
		if nestedPropertyName == "-" {
			delete(currSpec.Properties, nestedPropertyName)
			continue
		}
		// Create a new definition for the nested object by joining the parent definition name and the property name.
		// This is to ensure that the nested object is unique and does not conflict with other definitions.
		nestedDefinitionName := parentName + sanitizeReferenceName(nestedPropertyName)

		s, err := flattenRecursively(sw, nestedDefinitionName, nestedProperty)
		if err != nil {
			return currSpec, fmt.Errorf("error flattening OpenAPI spec: %w", err)
		}

		// The nested property is not an object, so we can skip adding it to the definitions and creating a reference.
		// We check this here as we want our recursive function to inspect both arrays and objects.
		if len(nestedProperty.Properties) == 0 {
			continue
		}

		sw.Definitions[nestedDefinitionName] = s

		// Reset the property to be a reference to the nested object.
		refName := definitionPrefix + nestedDefinitionName
		ref, err := jsonreference.New(refName)
		if err != nil {
			return currSpec, fmt.Errorf("error creating OpenAPI json reference for nested object: %w", err)
		}

		currSpec.Properties[nestedPropertyName] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Ref: spec.Ref{
					Ref: ref,
				},
			},
		}
	}

	return currSpec, nil
}

// flattedArrayObject flattens an OpenAPI array property.
func flattedArrayObject(parentName string, sw *spec.Swagger, itemsSchema *spec.Schema) (spec.Schema, *jsonreference.Ref, error) {
	nestedDefinitionName := parentName

	s, err := flattenRecursively(sw, nestedDefinitionName, *itemsSchema)
	if err != nil {
		return s, nil, fmt.Errorf("error flattening OpenAPI spec: %w", err)
	}

	if len(s.Properties) == 0 {
		return s, nil, nil
	}

	sw.Definitions[nestedDefinitionName] = s

	refName := definitionPrefix + nestedDefinitionName
	ref, err := jsonreference.New(refName)
	if err != nil {
		return s, nil, fmt.Errorf("error creating OpenAPI json reference for nested object: %w", err)
	}

	return s, &ref, nil
}

func sanitizeReferenceName(fieldName string) string {
	// If the field name is "arg" or "args", we need to change it to "Arguments" to avoid conflicts with Go reserved words.
	if s := strings.ToLower(fieldName); s == "arg" || s == "args" {
		return "Arguments"
	}

	// We need to strip out any hyphens and underscores in the reference.
	fieldName = cgstrings.Unhyphenate(fieldName)
	fieldName = cgstrings.ModifyStringAroundDelimeter(fieldName, "_", cgstrings.UppercaseFirst)

	return cgstrings.UppercaseFirst(fieldName)
}

// crdToOpenAPI generates the OpenAPI specs for a given CRD manifest.
func crdToOpenAPI(crd *extensionv1.CustomResourceDefinition) (map[string]*spec.Swagger, error) {
	openAPIManifests := make(map[string]*spec.Swagger)

	setCRDDefaults(crd)

	for _, v := range crd.Spec.Versions {
		// Defaults are not pruned here, but before being served.
		sw, err := builder.BuildOpenAPIV2(crd, v.Name, builder.Options{V2: true, StripValueValidation: true, StripNullable: true, AllowNonStructural: true, IncludeSelectableFields: true})
		if err != nil {
			return nil, err
		}

		err = flattenOpenAPI(sw)
		if err != nil {
			return nil, fmt.Errorf("error flattening OpenAPI spec: %w", err)
		}

		openAPIManifests[v.Name] = sw
	}

	return openAPIManifests, nil
}

// setCRDDefaults sets the default names for the CRD if they are not specified.
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

	for version, sw := range swagger {
		schemas[version] = *sw
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
