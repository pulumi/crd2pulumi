// Package json has functions for working with JSON objects
package json

import "encoding/json"

// RawMessage takes any JSON object and returns a json.RawMessage representation of it
func RawMessage(v any) (json.RawMessage, error) {
	return json.Marshal(v)
}
