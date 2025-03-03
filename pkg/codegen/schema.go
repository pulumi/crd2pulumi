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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pulumi/crd2pulumi/internal/slices"
	"github.com/pulumi/crd2pulumi/internal/unstruct"
	"github.com/pulumi/pulumi-kubernetes/provider/v4/pkg/gen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// DefaultName specifies the default value for the package name
const DefaultName = "crds"

// pulumiKubernetesNameShim is a hack name since upstream schemagen creates types
// in the `kubernetes` namespace. When binding, we need the package name to be the
// same namespace as the types.
const pulumiKubernetesNameShim = "kubernetes"

const (
	Boolean string = "boolean"
	Integer string = "integer"
	Number  string = "number"
	String  string = "string"
	Array   string = "array"
	Object  string = "object"
)

const anyTypeRef = "pulumi.json#/Any"

var anyTypeSpec = pschema.TypeSpec{
	Ref: anyTypeRef,
}

var arbitraryJSONTypeSpec = pschema.TypeSpec{
	Type:                 Object,
	AdditionalProperties: &anyTypeSpec,
}

var emptySpec = pschema.ComplexTypeSpec{
	ObjectTypeSpec: pschema.ObjectTypeSpec{
		Type:       Object,
		Properties: map[string]pschema.PropertySpec{},
	},
}

const (
	objectMetaRef        = "#/types/kubernetes:meta/v1:ObjectMeta"
	objectMetaToken      = "kubernetes:meta/v1:ObjectMeta"
	objectMetaPatchToken = "kubernetes:meta/v1:ObjectMetaPatch"
)

// Union type of integer and string
var intOrStringTypeSpec = pschema.TypeSpec{
	OneOf: []pschema.TypeSpec{
		{
			Type: Integer,
		},
		{
			Type: String,
		},
	},
}

// mergeSpecs merges a slice of OpenAPI specs into a single OpenAPI spec.
func mergeSpecs(specs []*spec.Swagger) (*spec.Swagger, error) {
	if len(specs) == 0 {
		return nil, errors.New("no OpenAPI specs to merge")
	}

	mergedSpecs, err := builder.MergeSpecs(specs[0], specs[1:]...)
	if err != nil {
		return nil, fmt.Errorf("error merging OpenAPI specs: %w", err)
	}

	return mergedSpecs, nil
}

// Returns the Pulumi package given a types map and a slice of the token types
// of every CustomResource. If includeObjectMetaType is true, then a
// ObjectMetaType type is also generated.
func genPackage(version string, crgenerators []CustomResourceGenerator, includeObjectMetaType bool) (*pschema.Package, error) {
	var allCRDSpecs []*spec.Swagger
	// Merge all OpenAPI specs into a single OpenAPI spec.
	for _, crg := range crgenerators {
		for _, spec := range crg.Schemas {
			allCRDSpecs = append(allCRDSpecs, &spec)
		}
	}

	mergedSpec, err := mergeSpecs(allCRDSpecs)
	if err != nil {
		return &pschema.Package{}, fmt.Errorf("could not merge OpenAPI specs: %w", err)
	}

	marshaledOpenAPISchema, err := json.Marshal(mergedSpec)
	if err != nil {
		return nil, fmt.Errorf("error marshalling OpenAPI spec: %v", err)
	}

	unstructuredOpenAPISchema := make(map[string]any)
	err = json.Unmarshal(marshaledOpenAPISchema, &unstructuredOpenAPISchema)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling OpenAPI spec: %v", err)
	}
	// We need to allow hyphens in the property names since Kubernetes CRDs could contain fields that have them and
	// crd2pulumi should not panic for backwards compatibility.
	// This will only currently work for Go and Python as they have the correct annotations to serialize/deserialize
	// the hyphenated fields to their non-hyphenated equivalents.
	// See: https://github.com/pulumi/crd2pulumi/issues/43
	pkgSpec := gen.PulumiSchema(unstructuredOpenAPISchema, gen.WithAllowHyphens(true), gen.WithPulumiKubernetesDependency(KubernetesProviderVersion))

	// Populate the package spec with information used in previous versions of crd2pulumi to maintain consistency
	// with older versions.
	pkgSpec.Name = pulumiKubernetesNameShim
	pkgSpec.Version = version
	pkgSpec.Config = pschema.ConfigSpec{}
	pkgSpec.Provider = pschema.ResourceSpec{}

	if !includeObjectMetaType {
		delete(pkgSpec.Types, objectMetaToken)
		delete(pkgSpec.Types, objectMetaPatchToken)
	}

	// Remove excess resources generated from the OpenAPI spec.
	for resourceName := range pkgSpec.Resources {
		if strings.HasPrefix(resourceName, "kubernetes:meta/v1:") {
			delete(pkgSpec.Resources, resourceName)
		}
	}

	pkg, err := pschema.ImportSpec(pkgSpec, nil, pschema.ValidationOptions{})
	if err != nil {
		return &pschema.Package{}, fmt.Errorf("could not import spec: %w", err)
	}

	pkg.Name = DefaultName

	return pkg, nil
}

// Returns true if the given TypeSpec is of type any; returns false otherwise
func isAnyType(typeSpec pschema.TypeSpec) bool {
	return typeSpec.Ref == anyTypeRef
}

// AddType converts the given OpenAPI `schema` to a ObjectTypeSpec and adds it
// to the `types` map under the given `name`. Recursively converts and adds all
// nested schemas as well.
func AddType(schema map[string]any, name string, types map[string]pschema.ComplexTypeSpec) {
	properties, foundProperties, _ := unstructured.NestedMap(schema, "properties")
	description, _, _ := unstructured.NestedString(schema, "description")
	schemaType, _, _ := unstructured.NestedString(schema, "type")
	required, _, _ := unstructured.NestedStringSlice(schema, "required")

	propertySpecs := map[string]pschema.PropertySpec{}
	for propertyName := range properties {
		// Ignore unnamed properties like "-".
		camelCase := strcase.ToCamel(propertyName)
		if camelCase == "" {
			continue
		}
		propertySchema, _, _ := unstructured.NestedMap(properties, propertyName)
		propertyDescription, _, _ := unstructured.NestedString(propertySchema, "description")
		typeSpec := GetTypeSpec(propertySchema, name+strcase.ToCamel(propertyName), types)
		// Pulumi's schema doesn't support defaults for objects, so ignore them.
		var defaultValue any
		if !(typeSpec.Type == "object" || typeSpec.Type == "array") {
			defaultValue, _, _ = unstructured.NestedFieldNoCopy(propertySchema, "default")
		}
		propertySpecs[propertyName] = pschema.PropertySpec{
			TypeSpec:    typeSpec,
			Description: propertyDescription,
			Default:     defaultValue,
		}
	}

	// If the type wasn't specified but we found properties, then we can infer that the type is an object
	if foundProperties && schemaType == "" {
		schemaType = Object
	}

	types[name] = pschema.ComplexTypeSpec{
		ObjectTypeSpec: pschema.ObjectTypeSpec{
			Type:        schemaType,
			Properties:  propertySpecs,
			Required:    required,
			Description: description,
		},
	}
}

// GetTypeSpec returns the corresponding pschema.TypeSpec for a OpenAPI v3
// schema. Handles nested pschema.TypeSpecs in case the schema type is an array,
// object, or "combined schema" (oneOf, allOf, anyOf). Also recursively converts
// and adds all schemas of type object to the types map.
func GetTypeSpec(schema map[string]any, name string, types map[string]pschema.ComplexTypeSpec) pschema.TypeSpec {
	if schema == nil {
		return anyTypeSpec
	}

	intOrString, foundIntOrString, _ := unstructured.NestedBool(schema, "x-kubernetes-int-or-string")
	if foundIntOrString && intOrString {
		return intOrStringTypeSpec
	}

	// If the schema is of the `oneOf` type: return a TypeSpec with the `OneOf`
	// field filled with the TypeSpec of all sub-schemas.
	oneOf, foundOneOf, _ := unstruct.NestedMapSlice(schema, "oneOf")
	if foundOneOf {
		oneOfTypeSpecs := make([]pschema.TypeSpec, 0, len(oneOf))
		for i, oneOfSchema := range oneOf {
			oneOfTypeSpec := GetTypeSpec(oneOfSchema, name+"OneOf"+strconv.Itoa(i), types)
			if isAnyType(oneOfTypeSpec) {
				return anyTypeSpec
			}
			oneOfTypeSpecs = append(oneOfTypeSpecs, oneOfTypeSpec)
		}
		return pschema.TypeSpec{
			OneOf: oneOfTypeSpecs,
		}
	}

	// If the schema is of `allOf` type: combine `properties` and `required`
	// fields of sub-schemas into a single schema. Then return the `TypeSpec`
	// of that combined schema.
	allOf, foundAllOf, _ := unstruct.NestedMapSlice(schema, "allOf")
	if foundAllOf {
		combinedSchema := CombineSchemas(true, allOf...)
		return GetTypeSpec(combinedSchema, name, types)
	}

	// If the schema is of `anyOf` type: combine only `properties` of
	// sub-schemas into a single schema, with all `properties` set to optional.
	// Then return the `TypeSpec` of that combined schema.
	anyOf, foundAnyOf, _ := unstruct.NestedMapSlice(schema, "anyOf")
	if foundAnyOf {
		combinedSchema := CombineSchemas(false, anyOf...)
		return GetTypeSpec(combinedSchema, name, types)
	}

	preserveUnknownFields, foundPreserveUnknownFields, _ := unstructured.NestedBool(schema, "x-kubernetes-preserve-unknown-fields")
	if foundPreserveUnknownFields && preserveUnknownFields {
		return arbitraryJSONTypeSpec
	}

	// If the the schema wasn't some combination of other types (`oneOf`,
	// `allOf`, `anyOf`), then it must have a "type" field, otherwise we
	// cannot represent it. If we cannot represent it, we simply set it to be
	// any type.
	schemaType, foundSchemaType, _ := unstructured.NestedString(schema, "type")
	if !foundSchemaType {
		return anyTypeSpec
	}

	switch schemaType {
	case Array:
		items, _, _ := unstructured.NestedMap(schema, "items")
		arrayTypeSpec := GetTypeSpec(items, name, types)
		return pschema.TypeSpec{
			Type:  Array,
			Items: &arrayTypeSpec,
		}
	case Object:
		AddType(schema, name, types)
		// If `additionalProperties` has a sub-schema, then we generate a type for a map from string --> sub-schema type
		additionalProperties, foundAdditionalProperties, _ := unstructured.NestedMap(schema, "additionalProperties")
		if foundAdditionalProperties {
			additionalPropertiesTypeSpec := GetTypeSpec(additionalProperties, name, types)
			return pschema.TypeSpec{
				Type:                 Object,
				AdditionalProperties: &additionalPropertiesTypeSpec,
			}
		}
		// `additionalProperties: true` is equivalent to `additionalProperties: {}`, meaning a map from string -> any
		additionalPropertiesIsTrue, additionalPropertiesIsTrueFound, _ := unstructured.NestedBool(schema, "additionalProperties")
		if additionalPropertiesIsTrueFound && additionalPropertiesIsTrue {
			return pschema.TypeSpec{
				Type:                 Object,
				AdditionalProperties: &anyTypeSpec,
			}
		}
		// If no properties are found, then it can be arbitrary JSON
		_, foundProperties, _ := unstructured.NestedMap(schema, "properties")
		if !foundProperties {
			return arbitraryJSONTypeSpec
		}
		// If properties are found, then we must specify those in a seperate interface
		return pschema.TypeSpec{
			Type: Object,
			Ref:  "#/types/" + name,
		}
	case Integer:
		fallthrough
	case Boolean:
		fallthrough
	case String:
		fallthrough
	case Number:
		return pschema.TypeSpec{
			Type: schemaType,
		}
	default:
		return anyTypeSpec
	}
}

// CombineSchemas combines the `properties` fields of the given sub-schemas into
// a single schema. Returns nil if no schemas are given. Returns the schema if
// only 1 schema is given. If combineRequired == true, then each sub-schema's
// `required` fields are also combined. In this case the combined schema's
// `required` field is of type []any, not []string.
func CombineSchemas(combineRequired bool, schemas ...map[string]any) map[string]any {
	if len(schemas) == 0 {
		return nil
	}
	if len(schemas) == 1 {
		return schemas[0]
	}

	combinedProperties := map[string]any{}
	combinedRequired := make([]string, 0)

	for _, schema := range schemas {
		properties, _, _ := unstructured.NestedMap(schema, "properties")
		for propertyName := range properties {
			propertySchema, _, _ := unstructured.NestedMap(properties, propertyName)
			combinedProperties[propertyName] = propertySchema
		}
		if combineRequired {
			required, foundRequired, _ := unstructured.NestedStringSlice(schema, "required")
			if foundRequired {
				combinedRequired = append(combinedRequired, required...)
			}
		}
	}

	combinedSchema := map[string]any{
		"type":       Object,
		"properties": combinedProperties,
	}
	if combineRequired {
		combinedSchema["required"] = slices.ToAny(combinedRequired)
	}
	return combinedSchema
}

func getToken(group, version, kind string) string {
	return fmt.Sprintf("kubernetes:%s/%s:%s", group, version, kind)
}
