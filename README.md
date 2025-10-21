# crd2pulumi
Generate typed CustomResources based on Kubernetes CustomResourceDefinitions.

## Goals

`crd2pulumi` is a CLI tool that generates typed CustomResources based on Kubernetes CustomResourceDefinition (CRDs). 
CRDs allow you to extend the Kubernetes API by defining your own schemas for custom objects. While Pulumi lets you create
 [CustomResources](https://www.pulumi.com/docs/reference/pkg/kubernetes/apiextensions/customresource/), there was previously 
 no strong-typing for these objects since every schema was, well, custom. This can be a massive headache for popular CRDs 
 such as [cert-manager](https://github.com/jetstack/cert-manager/tree/master/deploy/crds) or 
 [istio](https://github.com/istio/istio/tree/0321da58ca86fc786fb03a68afd29d082477e4f2/manifests/charts/base/crds), which 
 contain thousands of lines of complex YAML schemas. By generating typed versions of CustomResources, `crd2pulumi` makes 
 filling out their arguments more convenient by allowing you to leverage existing IDE type checking and autocomplete features.

## Building and Installation
If you wish to use `crd2pulumi` without developing the tool itself, you can use one of the [binary releases](https://github.com/pulumi/crd2pulumi/releases) hosted on this repository. 

### Homebrew
`crd2pulumi` can be installed on Mac from the Pulumi Homebrew tap.
```console
brew install pulumi/tap/crd2pulumi
```

`crd2pulumi` uses Go modules to manage dependencies. If you want to develop `crd2pulumi` itself, you'll need to have 
Go installed in order to build. Once you install this prerequisite, run the following to build the `crd2pulumi` binary 
and install it into `$GOPATH/bin`:

```bash
$ go build -ldflags="-X github.com/pulumi/crd2pulumi/gen.Version=dev" -o $GOPATH/bin/crd2pulumi main.go
```
The `ldflags` argument is necessary to dynamically set the `crd2pulumi` version at build time. However, the version 
itself can be anything, so you don't have to set it to `dev`.

Go should then automatically handle pulling the dependencies for you. If `$GOPATH/bin` is not on your path, you may 
want to move the `crd2pulumi` binary from `$GOPATH/bin` into a directory that is on your path.

## Usage
```bash
crd2pulumi is a CLI tool that generates typed Kubernetes
CustomResources to use in Pulumi programs, based on a
CustomResourceDefinition YAML schema.

Usage:
  crd2pulumi [-dgnp] [--nodejsPath path] [--pythonPath path] [--dotnetPath path] [--goPath path] [--javaPath path] <crd1.yaml> [crd2.yaml ...] [flags]
  crd2pulumi [command]

Examples:
crd2pulumi --nodejs crontabs.yaml
crd2pulumi -dgnp crd-certificates.yaml crd-issuers.yaml crd-challenges.yaml
crd2pulumi --pythonPath=crds/python/istio --nodejsPath=crds/nodejs/istio crd-all.gen.yaml crd-mixer.yaml crd-operator.yaml
crd2pulumi --pythonPath=crds/python/gke https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/master/deploy/managedcertificates-crd.yaml

Notice that by just setting a language-specific output path (--pythonPath, --nodejsPath, etc) the code will
still get generated, so setting -p, -n, etc becomes unnecessary.


Available Commands:
  help        Help about any command
  version     Print the version number of crd2pulumi

Flags:
  -d, --dotnet                       generate .NET
      --dotnetName string            name of generated .NET package (default "crds")
      --dotnetNamespace string       namespace of generated .NET package
      --dotnetPath string            optional .NET output dir
  -f, --force                        overwrite existing files
  -g, --go                           generate Go
      --goName string                name of generated Go package (default "crds")
      --goPath string                optional Go output dir
  -h, --help                         help for crd2pulumi
  -j, --java                         generate Java
      --javaBasePackage string       base package of generated Java package
      --javaName string              name of generated Java package (default "crds")
      --javaPath string              optional Java output dir
  -n, --nodejs                       generate NodeJS
      --nodejsName string            name of generated NodeJS package (default "crds")
      --nodejsNamespace string       namespace of generated NodeJS package
      --nodejsPath string            optional NodeJS output dir
  -p, --python                       generate Python
      --pythonName string            name of generated Python package (default "crds")
      --pythonPackagePrefix string   prefix of generated Python package
      --pythonPath string            optional Python output dir


Use "crd2pulumi [command] --help" for more information about a command.
```
Setting only a language-specific flag will output the generated code in the default directory; so `-d` will output to 
`crds/dotnet`, `-g` will output to `crds/go`, `-j` will output to `crds/java`, `-n` will output to `crds/nodejs`, and 
`-p` will output to `crds/python`. You can also specify a language-specific path (`--pythonPath`, `--nodejsPath`, etc) 
to control where the code will be outputted, in which case setting `-p`, `-n`, etc becomes unnecessary.

## Examples
Let's use the example CronTab CRD specified in `resourcedefinition.yaml` from the 
[Kubernetes Documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/). 

### TypeScript
To generate a strongly-typed CronTab CustomResource in TypeScript, we can run this command:
```bash
$ crd2pulumi --nodejsPath ./crontabs resourcedefinition.yaml
```
Now let's import the generated code into a Pulumi program that provisions the CRD and creates an instance of it.
```typescript
import * as crontabs from "./crontabs"
import * as pulumi from "@pulumi/pulumi"
import * as k8s from "@pulumi/kubernetes";

// Register the CronTab CRD.
const cronTabDefinition = new k8s.yaml.ConfigFile("my-crontab-definition", { file: "resourcedefinition.yaml" });

// Instantiate a CronTab resource.
const myCronTab = new crontabs.stable.v1.CronTab("my-new-cron-object",
{
    metadata: {
        name: "my-new-cron-object",
    },
    spec: {
        cronSpec: "* * * * */5",
        image: "my-awesome-cron-image",
    }
})

```
As you can see, the `CronTab` object is typed! For example, if you try to set
`cronSpec` to a non-string or add an extra field, your IDE should immediately warn you.

### Python
```bash
$ crd2pulumi --pythonPath ./crontabs resourcedefinition.yaml
```
```python
import pulumi_kubernetes as k8s
import crontabs.pulumi_crds as crontabs


# Register the CronTab CRD.
crontab_definition = k8s.yaml.ConfigFile("my-crontab-definition", file="resourcedefinition.yaml")

# Instantiate a CronTab resource.
crontab_instance = crontabs.stable.v1.CronTab(
    "my-new-cron-object",
    metadata=k8s.meta.v1.ObjectMetaArgs(
        name="my-new-cron-object"
    ),
    spec=crontabs.stable.v1.CronTabSpecArgs(
        cron_spec="* * * */5",
        image="my-awesome-cron-image",
    )
)

```

### Go
```bash
$ crd2pulumi --goPath ./crontabs resourcedefinition.yaml
```
Now we can access the `NewCronTab()` constructor. Create a `main.go` file with the following code. In this example, 
the Pulumi project's module is named `crds-go-final`, so the import path is `crds-go-final/crontabs/stable/v1`. Make 
sure to swap this out with your own module's name.
```go
package main

import (
	crontabs_v1 "crds-go-final/crontabs/stable/v1"

	meta_v1 "github.com/pulumi/pulumi-kubernetes/sdk/v2/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
    // Register the CronTab CRD.
    _, err := yaml.NewConfigFile(ctx, "my-crontab-definition",
      &yaml.ConfigFileArgs{
        File: "resourcedefinition.yaml",
      },
    )
    if err != nil {
      return err
    }

		// Instantiate a CronTab resource.
		_, err := crontabs_v1.NewCronTab(ctx, "cronTabInstance", &crontabs_v1.CronTabArgs{
			Metadata: &meta_v1.ObjectMetaArgs{
				Name: pulumi.String("my-new-cron-object"),
			},
			Spec: crontabs_v1.CronTabSpecArgs{
				CronSpec: pulumi.String("* * * * */5"),
				Image:    pulumi.String("my-awesome-cron-image"),
				Replicas: pulumi.IntPtr(3),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

```

### C\#
```bash
$ crd2pulumi --dotnetPath ./crontabs resourcedefinition.yaml
```
```csharp
using Pulumi;
using Pulumi.Kubernetes.Yaml;
using Pulumi.Kubernetes.Types.Inputs.Meta.V1;

class MyStack : Stack
{
    public MyStack()
    {
    // Register a CronTab CRD.
    var cronTabDefinition = new Pulumi.Kubernetes.Yaml.ConfigFile("my-crontab-definition",
        new ConfigFileArgs{
            File = "resourcedefinition.yaml"
        }
    );

    // Instantiate a CronTab resource.
    var cronTabInstance = new Pulumi.Crds.Stable.V1.CronTab("cronTabInstance",
        new Pulumi.Kubernetes.Types.Inputs.Stable.V1.CronTabArgs{
            Metadata = new ObjectMetaArgs{
                Name = "my-new-cron-object"
            },
            Spec = new Pulumi.Kubernetes.Types.Inputs.Stable.V1.CronTabSpecArgs{
                CronSpec = "* * * * */5",
                Image = "my-awesome-cron-image"
            }
        });    
    }
}

```

> If you get an `Duplicate 'global::System.Runtime.Versioning.TargetFrameworkAttribute' attribute` error when trying to run `pulumi up`, then try deleting the `crontabs/bin` and `crontabs/obj` folders.

### Java
```bash
$ crd2pulumi --javaPath ./crontabs resourcedefinition.yaml
```
```java
package com.example;

import com.pulumi.Pulumi;

public class MyStack {

    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Register a CronTab CRD (Coming Soon - see https://www.pulumi.com/registry/packages/kubernetes/api-docs/yaml/configfile/)

            // Instantiate a CronTab resource.
            var cronTabInstance = new com.pulumi.crds.stable.v1.CronTab("cronTabInstance",
                    com.pulumi.crds.stable.v1.CronTabArgs.builder()
                            .metadata(com.pulumi.kubernetes.meta.v1.inputs.ObjectMetaArgs.builder()
                                    .name("my-new-cron-object")
                                    .build())
                            .spec(com.pulumi.kubernetes.stable.v1.inputs.CronTabSpecArgs.builder()
                                    .cronSpec("* * * * */5")
                                    .image("my-awesome-cron-image")
                                    .build())
                            .build());
        });
    }
}

```

Now let's run the program and perform the update.
```bash
$ pulumi up
Previewing update (dev):
  Type                                                      Name                Plan
  pulumi:pulumi:Stack                                       examples-dev
 +   ├─ kubernetes:stable.example.com:CronTab                   my-new-cron-object  create
 +   └─ kubernetes:apiextensions.k8s.io:CustomResourceDefinition  my-crontab-definition  create
Resources:
  + 2 to create
  1 unchanged
Do you want to perform this update? yes
Updating (dev):
  Type                                                      Name                Status
  pulumi:pulumi:Stack                                       examples-dev
 +   ├─ kubernetes:stable.example.com:CronTab                   my-new-cron-object  created
 +   └─ kubernetes:apiextensions.k8s.io:CustomResourceDefinition  my-crontab-definition  created
Outputs:
  urn: "urn:pulumi:dev::examples::kubernetes:stable.example.com/v1:CronTab::my-new-cron-object"
Resources:
  + 2 created
  1 unchanged
Duration: 17s
Permalink: https://app.pulumi.com/albert-zhong/examples/dev/updates/4
```
It looks like both the CronTab definition and instance were both created! Finally, let's verify that they were created
by manually viewing the raw YAML data:
```bash
$ kubectl get ct -o yaml
```
```yaml
- apiVersion: stable.example.com/v1
  kind: CronTab
  metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"stable.example.com/v1","kind":"CronTab","metadata":{"labels":{"app.kubernetes.io/managed-by":"pulumi"},"name":"my-new-cron-object"},"spec":{"cronSpec":"* * * * */5","image":"my-awesome-cron-image"}}
  creationTimestamp: "2020-08-10T09:50:38Z"
  generation: 1
  labels:
    app.kubernetes.io/managed-by: pulumi
  name: my-new-cron-object
  namespace: default
  resourceVersion: "1658962"
  selfLink: /apis/stable.example.com/v1/namespaces/default/crontabs/my-new-cron-object
  uid: 5e2c56a2-7332-49cf-b0fc-211a0892c3d5
  spec:
  cronSpec: '* * * * */5'
  image: my-awesome-cron-image
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```
