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

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pulumi/crd2pulumi/pkg/codegen"
	"github.com/spf13/cobra"
)

// Version specifies the crd2pulumi version. It should be set by the linker via LDFLAGS. This defaults to dev
var Version = "dev"

const long = `crd2pulumi is a CLI tool that generates typed Kubernetes
CustomResources to use in Pulumi programs, based on a
CustomResourceDefinition YAML schema.`

const example = `crd2pulumi --nodejs crontabs.yaml
crd2pulumi -dgnp crd-certificates.yaml crd-issuers.yaml crd-challenges.yaml
crd2pulumi --pythonPath=crds/python/istio --nodejsPath=crds/nodejs/istio crd-all.gen.yaml crd-mixer.yaml crd-operator.yaml
crd2pulumi --pythonPath=crds/python/gke https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/master/deploy/managedcertificates-crd.yaml

Notice that by just setting a language-specific output path (--pythonPath, --nodejsPath, etc) the code will
still get generated, so setting -p, -n, etc becomes unnecessary.
`

func Execute() error {
	dotNetSettings := &codegen.CodegenSettings{Language: "dotnet"}
	goSettings := &codegen.CodegenSettings{Language: "go"}
	nodejsSettings := &codegen.CodegenSettings{Language: "nodejs"}
	pythonSettings := &codegen.CodegenSettings{Language: "python"}
	javaSettings := &codegen.CodegenSettings{Language: "java"}
	allSettings := []*codegen.CodegenSettings{dotNetSettings, goSettings, nodejsSettings, pythonSettings, javaSettings}

	var force bool
	var packageVersion string

	rootCmd := &cobra.Command{
		Use:          "crd2pulumi [-dgnp] [--nodejsPath path] [--pythonPath path] [--dotnetPath path] [--goPath path] <crd1.yaml> [crd2.yaml ...]",
		Short:        "A tool that generates typed Kubernetes CustomResources",
		Long:         long,
		Example:      example,
		SilenceUsage: true, // Don't show the usage message upon program error
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return errors.New("must specify at least one CRD YAML file")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, cs := range allSettings {
				if force {
					cs.Overwrite = true
				}
				if cs.OutputDir != "" {
					cs.ShouldGenerate = true
				}
				cs.PackageVersion = packageVersion
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var stdinData []byte
			shouldUseStdin := len(args) == 1 && args[0] == "-"
			if shouldUseStdin {
				var err error
				stdinData, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed reading CRDs from stdin: %w", err)
				}
			}
			for _, cs := range allSettings {
				if !cs.ShouldGenerate {
					continue
				}
				var err error
				if shouldUseStdin {
					err = codegen.Generate(cs, []io.ReadCloser{io.NopCloser(bytes.NewBuffer(stdinData))})
				} else {
					err = codegen.GenerateFromFiles(cs, args)
				}
				if err != nil {
					return fmt.Errorf("error generating code: %w", err)
				}
				fmt.Printf("Successfully generated %s code.\n", cs.Language)
			}
			return nil
		},
	}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of crd2pulumi",
		Long:  `All software has versions. This is crd2pulumi's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	})

	f := rootCmd.PersistentFlags()
	f.BoolVarP(&force, "force", "f", false, "overwrite existing files")
	f.StringVarP(&packageVersion, "version", "v", "0.0.0-dev", "version of the generated package")

	f.StringVarP(&dotNetSettings.PackageName, "dotnetName", "", codegen.DefaultName, "name of generated .NET package")
	f.StringVarP(&goSettings.PackageName, "goName", "", codegen.DefaultName, "name of generated Go package")
	f.StringVarP(&nodejsSettings.PackageName, "nodejsName", "", codegen.DefaultName, "name of generated NodeJS package")
	f.StringVarP(&pythonSettings.PackageName, "pythonName", "", codegen.DefaultName, "name of generated Python package")
	f.StringVarP(&javaSettings.PackageName, "javaName", "", codegen.DefaultName, "name of generated Java package")

	f.StringVarP(&dotNetSettings.OutputDir, "dotnetPath", "", "", "optional .NET output dir")
	f.StringVarP(&goSettings.OutputDir, "goPath", "", "", "optional Go output dir")
	f.StringVarP(&nodejsSettings.OutputDir, "nodejsPath", "", "", "optional NodeJS output dir")
	f.StringVarP(&pythonSettings.OutputDir, "pythonPath", "", "", "optional Python output dir")
	f.StringVarP(&javaSettings.OutputDir, "javaPath", "", "", "optional Java output dir")

	f.BoolVarP(&dotNetSettings.ShouldGenerate, "dotnet", "d", false, "generate .NET")
	f.BoolVarP(&goSettings.ShouldGenerate, "go", "g", false, "generate Go")
	f.BoolVarP(&nodejsSettings.ShouldGenerate, "nodejs", "n", false, "generate NodeJS")
	f.BoolVarP(&pythonSettings.ShouldGenerate, "python", "p", false, "generate Python")
	f.BoolVarP(&javaSettings.ShouldGenerate, "java", "j", false, "generate Java")
	return rootCmd.Execute()
}
