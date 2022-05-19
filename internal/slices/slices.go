// Package slices has functions for working with slices
package slices

// ToAny converts any type of slice to new []any slice.
func ToAny[T any](slice []T) []any {
	anySlice := make([]any, len(slice))
	for i, v := range slice {
		anySlice[i] = v
	}
	return anySlice
}
