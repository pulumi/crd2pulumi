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
	"path/filepath"
)

var SupportedLanguages = []string{
	DotNet,
	Go,
	NodeJS,
	Python,
}

const DotNet string = "dotnet"
const Go string = "go"
const NodeJS string = "nodejs"
const Python string = "python"
const Java string = "java"

type CodegenSettings struct {
	Language       string
	OutputDir      string
	PackageName    string
	PackageVersion string
	Overwrite      bool
	ShouldGenerate bool
}

func (cs *CodegenSettings) Path() string {
	if cs.OutputDir == "" {
		cs.OutputDir = filepath.Join(cs.PackageName, cs.Language)
	}
	return cs.OutputDir
}
