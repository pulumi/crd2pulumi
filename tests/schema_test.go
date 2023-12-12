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

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/pulumi/crd2pulumi/pkg/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const TestCombineSchemasYAML = "test-combineschemas.yaml"
const TestGetTypeSpecYAML = "test-gettypespec.yaml"
const TestGetTypeSpecJSON = "test-gettypespec.json"

// UnmarshalYaml un-marshals one and only one YAML document from a file
func UnmarshalYaml(yamlFile []byte) (map[string]any, error) {
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlFile), 128)
	var value map[string]any
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return value, nil
}

func UnmarshalSchemas(yamlPath string) (map[string]any, error) {
	yamlFile, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	return UnmarshalYaml(yamlFile)
}

func UnmarshalTypeSpecJSON(jsonPath string) (map[string]pschema.TypeSpec, error) {
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", jsonPath, err)
	}
	var v map[string]pschema.TypeSpec
	err = json.Unmarshal(jsonFile, &v)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal %s: %w", jsonPath, err)
	}
	return v, nil
}

func TestCombineSchemas(t *testing.T) {
	// Test that CombineSchemas on no schemas returns nil
	assert.Nil(t, codegen.CombineSchemas(false))
	assert.Nil(t, codegen.CombineSchemas(true))

	// Unmarshal some testing schemas
	schemas, err := UnmarshalSchemas(TestCombineSchemasYAML)
	assert.NoError(t, err)
	person := schemas["person"].(map[string]any)
	employee := schemas["employee"].(map[string]any)

	// Test that CombineSchemas with 1 schema returns the same schema
	assert.Equal(t, person, codegen.CombineSchemas(true, person))
	assert.Equal(t, person, codegen.CombineSchemas(false, person))

	// Test CombineSchemas with 2 schemas and combineSchemas = true
	personAndEmployeeWithRequiredExpected := schemas["personAndEmployeeWithRequired"].(map[string]any)
	personAndEmployeeWithRequiredActual := codegen.CombineSchemas(true, person, employee)
	assert.EqualValues(t, personAndEmployeeWithRequiredExpected, personAndEmployeeWithRequiredActual)

	// Test CombineSchemas with 2 schemas and combineSchemas = false
	personAndEmployeeWithoutRequiredExpected := schemas["personAndEmployeeWithoutRequired"].(map[string]any)
	personAndEmployeeWithoutRequiredActual := codegen.CombineSchemas(false, person, employee)
	assert.EqualValues(t, personAndEmployeeWithoutRequiredExpected, personAndEmployeeWithoutRequiredActual)
}

func TestGetTypeSpec(t *testing.T) {
	// codegen.GetTypeSpec wants us to pass in a types map
	// (map[string]pschema.ObjectTypeSpec{}) to add object refs when we see
	// them. However we only want the returned pschema.TypeSpec, so this
	// wrapper function creates a placeholder types map and just returns
	// the pschema.TypeSpec. Since our initial name arg is "", this causes all
	// objects to have the ref "#/types/"
	getOnlyTypeSpec := func(schema map[string]any) pschema.TypeSpec {
		placeholderTypes := map[string]pschema.ComplexTypeSpec{}
		return codegen.GetTypeSpec(schema, "", placeholderTypes)
	}

	// Load YAML schemas
	schemas, err := UnmarshalSchemas(TestGetTypeSpecYAML)
	assert.NoError(t, err)

	// Load expected TypeSpec outputs as JSON
	typeSpecs, err := UnmarshalTypeSpecJSON(TestGetTypeSpecJSON)
	assert.NoError(t, err)

	for name := range schemas {
		expected, ok := typeSpecs[name]
		assert.True(t, ok)

		schema := schemas[name].(map[string]any)
		actual := getOnlyTypeSpec(schema)

		assert.EqualValues(t, expected, actual)
	}
}
