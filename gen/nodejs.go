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

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v2/codegen/nodejs"
)

const nodejsMetaPath = "meta/v1.ts"
const nodejsMetaFile = `import * as k8s from "@pulumi/kubernetes";

export type ObjectMeta = k8s.types.input.meta.v1.ObjectMeta;
`

func (pg *PackageGenerator) genNodeJS(outputDir string, name string) error {
	if files, err := pg.genNodeJSFiles(name); err != nil {
		return err
	} else if err := writeFiles(files, outputDir); err != nil {
		return err
	}
	return nil
}

func (pg *PackageGenerator) genNodeJSFiles(name string) (map[string]*bytes.Buffer, error) {
	pkg := pg.SchemaPackage()

	oldName := pkg.Name
	pkg.Name = name
	pkg.Language["nodejs"] = rawMessage(map[string]interface{}{
		"moduleToPackage": pg.moduleToPackage(),
	})

	files, err := nodejs.GeneratePackage(tool, pkg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate nodejs package")
	}

	pkg.Name = oldName
	delete(pkg.Language, NodeJS)

	// Remove ${VERSION} in package.json
	packageJSON, ok := files["package.json"]
	if !ok {
		return nil, errors.New("cannot find generated package.json")
	}
	files["package.json"] = bytes.ReplaceAll(packageJSON, []byte("${VERSION}"), []byte(""))

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
