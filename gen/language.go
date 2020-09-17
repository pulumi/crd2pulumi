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

package gen

import "os"

// LanguageSettings containers the output paths and package names for each language.
// If a path field is nil, the language won't be generated at all.
type LanguageSettings struct {
	NodeJSPath *string
	PythonPath *string
	DotNetPath *string
	GoPath     *string
	NodeJSName string
	PythonName string
	DotNetName string
	GoName     string
}

// Returns true if at least one of the language-specific output paths already exists. If true, then a slice of the
// paths that already exist are also returned.
func (ls LanguageSettings) hasExistingPaths() (bool, []string) {
	pathExists := func(path string) bool {
		_, err := os.Stat(path)
		return !os.IsNotExist(err)
	}
	var existingPaths []string
	if ls.NodeJSPath != nil && pathExists(*ls.NodeJSPath) {
		existingPaths = append(existingPaths, *ls.NodeJSPath)
	}
	if ls.PythonPath != nil && pathExists(*ls.PythonPath) {
		existingPaths = append(existingPaths, *ls.PythonPath)
	}
	if ls.DotNetPath != nil && pathExists(*ls.DotNetPath) {
		existingPaths = append(existingPaths, *ls.DotNetPath)
	}
	if ls.GoPath != nil && pathExists(*ls.GoPath) {
		existingPaths = append(existingPaths, *ls.GoPath)
	}
	return len(existingPaths) > 0, existingPaths
}

// GeneratesAtLeastOneLanguage returns true if and only if at least one language would be generated.
func (ls LanguageSettings) GeneratesAtLeastOneLanguage() bool {
	return ls.NodeJSPath != nil || ls.PythonPath != nil || ls.DotNetPath != nil || ls.GoPath != nil
}
