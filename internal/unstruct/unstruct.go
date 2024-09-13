// unstruct has utilities for working with k8s unstructured data
package unstruct

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// NestedMapSlice returns a copy of []map[string]any value of a nested field.
// Returns false if value is not found and an error if not a []any or contains non-map items in the slice.
// If the value is found but not of type []any, this still returns true.
func NestedMapSlice(obj map[string]any, fields ...string) ([]map[string]any, bool, error) {
	val, found, err := unstructured.NestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}
	m, ok := val.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []any", jsonPath(fields), val, val)
	}
	mapSlice := make([]map[string]any, 0, len(m))
	for _, v := range m {
		if strMap, ok := v.(map[string]any); ok {
			mapSlice = append(mapSlice, strMap)
		} else {
			return nil, false, fmt.Errorf("%v accessor error: contains non-map key in the slice: %v is of the type %T, expected map[string]any", jsonPath(fields), v, v)
		}
	}
	return mapSlice, true, nil
}

func jsonPath(fields []string) string {
	return "." + strings.Join(fields, ".")
}

const CRD = "CustomResourceDefinition"

// UnmarshalYamls un-marshals the YAML documents in the given file into a slice of unstruct.Unstructureds, one for each
// CRD. Only returns the YAML files for Kubernetes manifests that are CRDs and ignores others. Returns an error if any
// document failed to unmarshal.
func UnmarshalYamls(yamlFiles [][]byte) ([]extensionv1.CustomResourceDefinition, error) {
	var crds []extensionv1.CustomResourceDefinition
	for _, yamlFile := range yamlFiles {
		var err error
		dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlFile), 128)
		for err != io.EOF {
			var crd extensionv1.CustomResourceDefinition
			if err = dec.Decode(&crd); err != nil && err != io.EOF {
				return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
			}
			if crd.Kind == CRD {
				crds = append(crds, crd)
			}
		}
	}
	return crds, nil
}
