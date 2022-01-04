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
	"os"
	"path/filepath"

	"github.com/pulumi/crd2pulumi/gen"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	DotNet string = "dotnet"
	Go     string = "go"
	NodeJS string = "nodejs"
	Python string = "python"
)

const (
	DotNetPath string = "dotnetPath"
	GoPath     string = "goPath"
	NodeJSPath string = "nodejsPath"
	PythonPath string = "pythonPath"
)

const (
	DotNetName string = "dotnetName"
	GoName     string = "goName"
	NodeJSName string = "nodejsName"
	PythonName string = "pythonName"
)

const defaultOutputPath = "crds/"

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

// NewLanguageSettings returns the parsed language settings given a set of flags. Also returns a list of notices for
// possibly misinterpreted flags.
func NewLanguageSettings(flags *pflag.FlagSet) (gen.LanguageSettings, []string) {
	nodejs, _ := flags.GetBool(NodeJS)
	python, _ := flags.GetBool(Python)
	dotnet, _ := flags.GetBool(DotNet)
	golang, _ := flags.GetBool(Go)

	nodejsPath, _ := flags.GetString(NodeJSPath)
	pythonPath, _ := flags.GetString(PythonPath)
	dotnetPath, _ := flags.GetString(DotNetPath)
	goPath, _ := flags.GetString(GoPath)

	nodejsName, _ := flags.GetString(NodeJSName)
	pythonName, _ := flags.GetString(PythonName)
	dotNetName, _ := flags.GetString(DotNetName)
	goName, _ := flags.GetString(GoName)

	var notices []string
	ls := gen.LanguageSettings{
		NodeJSName: nodejsName,
		PythonName: pythonName,
		DotNetName: dotNetName,
		GoName:     goName,
	}
	if nodejsPath != "" {
		ls.NodeJSPath = &nodejsPath
		if nodejs {
			notices = append(notices, "-n is not necessary if --nodejsPath is already set")
		}
	} else if nodejs || nodejsName != gen.DefaultName {
		path := filepath.Join(defaultOutputPath, NodeJS)
		ls.NodeJSPath = &path
	}
	if pythonPath != "" {
		ls.PythonPath = &pythonPath
		if python {
			notices = append(notices, "-p is not necessary if --pythonPath is already set")
		}
	} else if python || pythonName != gen.DefaultName {
		path := filepath.Join(defaultOutputPath, Python)
		ls.PythonPath = &path
	}
	if dotnetPath != "" {
		ls.DotNetPath = &dotnetPath
		if dotnet {
			notices = append(notices, "-d is not necessary if --dotnetPath is already set")
		}
	} else if dotnet || dotNetName != gen.DefaultName {
		path := filepath.Join(defaultOutputPath, DotNet)
		ls.DotNetPath = &path
	}
	if goPath != "" {
		ls.GoPath = &goPath
		if golang {
			notices = append(notices, "-g is not necessary if --goPath is already set")
		}
	} else if golang || goName != gen.DefaultName {
		path := filepath.Join(defaultOutputPath, Go)
		ls.GoPath = &path
	}
	return ls, notices
}

var forceValue bool
var nodeJSValue, pythonValue, dotNetValue, goValue bool
var nodeJSPathValue, pythonPathValue, dotNetPathValue, goPathValue string
var nodeJSNameValue, pythonNameValue, dotNetNameValue, goNameValue string

func Execute() error {
	rootCmd := &cobra.Command{
		Use:     "crd2pulumi [-dgnp] [--nodejsPath path] [--pythonPath path] [--dotnetPath path] [--goPath path] <crd1.yaml> [crd2.yaml ...]",
		Short:   "A tool that generates typed Kubernetes CustomResources",
		Long:    long,
		Example: example,
		Args: func(cmd *cobra.Command, args []string) error {
			if ls, _ := NewLanguageSettings(cmd.Flags()); !ls.GeneratesAtLeastOneLanguage() {
				return errors.New("must specify at least one language")
			}

			err := cobra.MinimumNArgs(1)(cmd, args)
			if err != nil {
				return errors.New("must specify at least one CRD YAML file")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			ls, notices := NewLanguageSettings(cmd.Flags())
			for _, notice := range notices {
				fmt.Println("notice: " + notice)
			}

			err := gen.Generate(ls, args, force)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(-1)
			}

			fmt.Println("Successfully generated code.")
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&forceValue, "force", "f", false, "overwrite existing files")
	rootCmd.PersistentFlags().BoolVarP(&nodeJSValue, NodeJS, "n", false, "generate NodeJS")
	rootCmd.PersistentFlags().BoolVarP(&pythonValue, Python, "p", false, "generate Python")
	rootCmd.PersistentFlags().BoolVarP(&dotNetValue, DotNet, "d", false, "generate .NET")
	rootCmd.PersistentFlags().BoolVarP(&goValue, Go, "g", false, "generate Go")
	rootCmd.PersistentFlags().StringVar(&nodeJSPathValue, NodeJSPath, "", "optional NodeJS output dir")
	rootCmd.PersistentFlags().StringVar(&pythonPathValue, PythonPath, "", "optional Python output dir")
	rootCmd.PersistentFlags().StringVar(&dotNetPathValue, DotNetPath, "", "optional .NET output dir")
	rootCmd.PersistentFlags().StringVar(&goPathValue, GoPath, "", "optional Go output dir")
	rootCmd.PersistentFlags().StringVar(&nodeJSNameValue, NodeJSName, gen.DefaultName, "name of NodeJS package")
	rootCmd.PersistentFlags().StringVar(&pythonNameValue, PythonName, gen.DefaultName, "name of Python package")
	rootCmd.PersistentFlags().StringVar(&dotNetNameValue, DotNetName, gen.DefaultName, "name of .NET package")
	rootCmd.PersistentFlags().StringVar(&goNameValue, GoName, gen.DefaultName, "name of Go package")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of crd2pulumi",
		Long:  `All software has versions. This is crd2pulumi's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(gen.Version)
		},
	})

	return rootCmd.Execute()
}
