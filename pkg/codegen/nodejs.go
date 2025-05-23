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
	"bytes"
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
)

const nodejsName = "nodejs"
const nodejsMetaPath = "meta/v1.ts"
const nodejsMetaFile = `import * as k8s from "@pulumi/kubernetes";

export type ObjectMeta = k8s.types.input.meta.v1.ObjectMeta;
export type ObjectMetaPatch = k8s.types.input.meta.v1.ObjectMetaPatch;
`

func GenerateNodeJS(pg *PackageGenerator, name string) (map[string]*bytes.Buffer, error) {
	pkg := pg.SchemaPackageWithObjectMetaType()
	oldName := pkg.Name
	pkg.Name = name

	files, err := nodejs.GeneratePackage(PulumiToolName, pkg, nil, nil, true, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate nodejs package: %w", err)
	}

	pkg.Name = oldName
	delete(pkg.Language, nodejsName)

	// Pin the kubernetes provider version used
	utilities, ok := files["utilities.ts"]
	if !ok {
		return nil, fmt.Errorf("cannot find generated utilities.ts")
	}
	files["utilities.ts"] = bytes.ReplaceAll(utilities,
		[]byte("export function getVersion(): string {"),
		[]byte(fmt.Sprintf(`export const getVersion: () => string = () => "%s"

function unusedGetVersion(): string {`, KubernetesProviderVersion)))

	// Create a helper `meta/v1.ts` script that exports the ObjectMeta class from the SDK. If there happens to already
	// be a `meta/v1.ts` file, then just append the script.
	if code, ok := files[nodejsMetaPath]; !ok {
		files[nodejsMetaPath] = []byte(nodejsMetaFile)
	} else {
		files[nodejsMetaPath] = append(code, []byte("\n"+nodejsMetaFile)...)
	}

	buffers := map[string]*bytes.Buffer{}
	for name, code := range files {
		buffers[name] = bytes.NewBuffer(code)
	}

	return buffers, nil
}
