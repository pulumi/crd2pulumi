// Package versions has useful functions for working with versions.
package versions

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var alphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// SplitGroupVersion returns the <group> and <version> field of a string in the
// format <group>/<version>
func SplitGroupVersion(groupVersion string) (string, string, error) {
	parts := strings.Split(groupVersion, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected a version string with the format <group>/<version>, but got %q", groupVersion)
	}
	return parts[0], parts[1], nil
}

// groupPrefix returns the first word in the dot-separated group string, with
// all non-alphanumeric characters removed.
func GroupPrefix(group string) (string, error) {
	if group == "" {
		return "", fmt.Errorf("group cannot be empty")
	}
	return removeNonAlphanumeric(strings.Split(group, ".")[0]), nil
}

// Capitalizes and returns the given version. For example,
// VersionToUpper("v2beta1") returns "V2Beta1".
func VersionToUpper(version string) string {
	var sb strings.Builder
	for i, r := range version {
		if unicode.IsLetter(r) && (i == 0 || !unicode.IsLetter(rune(version[i-1]))) {
			sb.WriteRune(unicode.ToUpper(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

//  removeNonAlphanumeric removes all non-alphanumeric characters
func removeNonAlphanumeric(input string) string {
	return alphanumericRegex.ReplaceAllString(input, "")
}
