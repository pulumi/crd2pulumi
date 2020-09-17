## 1.0.5 (September 17, 2020)

### Bug Fixes

-   Fix parsing of top-level `x-kubernetes-preserve-unknown-fields` CRD schemas
-   Fix un-marshalling of multiple CRD YAML files

### Improvements

-   Add CLI flags to customize language-specific package names
-   Remove faulty TypeScript CustomResourceDefinition wrappers, and related tests/utility functions

## 1.0.4 (September 10, 2020)

### Bug Fixes

-   Fix parsing of CRDs without schemas (https://github.com/pulumi/pulumi-kubernetes/issues/1302)
-   Fix OpenAPIV3 schema parsing of implicit objects (https://github.com/pulumi/pulumi-kubernetes/issues/1299)
-   Fix generation of non-v1 and non-v1beta1 CustomResources (https://github.com/pulumi/crd2pulumi/issues/2)

## 1.0.3 (September 8, 2020)

### Improvements

-   Add support for all Pulumi SDK languages (TypeScript, Python, C#, Go)
-   Add CLI flags to generate any combination of the four supported languages to an optional output path
