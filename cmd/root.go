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

const example = `crd2pulumi --lang typescript crontabs.yaml
crd2pulumi --lang go crd-certificates.yaml crd-issuers.yaml crd-challenges.yaml
crd2pulumi --lang python --packageName myProject --outputDir=crds/python/istio crd-all.gen.yaml crd-mixer.yaml crd-operator.yaml
crd2pulumi --lang csharp --outputDir=crds/dotnet/gke https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/master/deploy/managedcertificates-crd.yaml
`

func Execute() error {
	cs := &codegen.CodegenSettings{}

	rootCmd := &cobra.Command{
		Use:          "crd2pulumi --lang languageName [--outputDir path] [--packageName name] [--force] <crd1.yaml> [crd2.yaml ...]",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if len(args) == 1 && args[0] == "-" {
				err = codegen.Generate(cs, []io.ReadCloser{os.Stdin})
			} else {
				err = codegen.GenerateFromFiles(cs, args)
			}
			if err != nil {
				return fmt.Errorf("error generating code: %w", err)
			}
			fmt.Println("Successfully generated code.")
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
	f.BoolVarP(&cs.Overwrite, "force", "f", false, "force overwrite of existing files")
	f.StringVarP(&cs.OutputDir, "outputDir", "o", "", "optional output path for generated code")
	f.StringVarP(&cs.Language, "lang", "l", "", fmt.Sprintf("language to generate code for, one of: %+v", codegen.SupportedLanguages))
	f.StringVarP(&cs.PackageName, "package", "p", codegen.DefaultName, "name of the generated package")
	f.StringVarP(&cs.PackageVersion, "version", "v", "0.0.0-dev", "version of the generated package")
	return rootCmd.Execute()
}
