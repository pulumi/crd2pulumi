# CHANGELOG

## 1.6.0 (2025-10-17)

- Configurable package namespaces/prefixes. (https://github.com/pulumi/crd2pulumi/pull/247)

## 1.5.4 (2024-11-13)

- NodeJS now uses correct input/output types for object metadata. (https://github.com/pulumi/crd2pulumi/issues/158)

## 1.5.3 (2024-09-30)

- Fix crd2pulumi not generating all CRD versions. [#152](https://github.com/pulumi/crd2pulumi/issues/152)
- Fix crd2pulumi generating packages and types with incorrect group names. [#152](https://github.com/pulumi/crd2pulumi/issues/152)

## 1.5.2 (2024-09-16)

- Set the pulumi-kubernetes dependency for Python packages to v4.18.0. [#148](https://github.com/pulumi/crd2pulumi/issues/148)
- Fixed generating Go types for StringMapArrayMap types. [#147](https://github.com/pulumi/crd2pulumi/issues/147)

## 1.5.1 (2024-09-13)

- Fixed Patch varaints not generated for types that end in List. [#146](https://github.com/pulumi/crd2pulumi/pull/146)

## 1.5.0 (2024-09-13)

### Added
- Patch variant resources are now generated for all custom resources. Patch resources allow you to modify and an existing custom resource. For more details on using Patch resources, see our [documentation](https://www.pulumi.com/registry/packages/kubernetes/how-to-guides/managing-resources-with-server-side-apply/#patch-a-resource).

### Changed
- The Pulumi schema generation now utilizes the library from the Pulumi Kubernetes provider, replacing the previous custom implementation. This resolves a number of correctness issues when generating code. [#143](https://github.com/pulumi/crd2pulumi/pull/143)
- Golang package generation now correctly adheres to the `--goPath` CLI flag, aligning with the behavior of other languages. [#89](https://github.com/pulumi/crd2pulumi/issues/89)
- CRDs with oneOf fields are now correctly typed and not generic. [#97](https://github.com/pulumi/crd2pulumi/issues/97)


### Fixed
- Various code generation correctness issues have been addressed, including:
  - Python packages can now be successfully imported and consumed by Pulumi Python programs. [#113](https://github.com/pulumi/crd2pulumi/issues/113)
  - Golang packages no longer produce compilation errors due to duplicate declarations. [#104](https://github.com/pulumi/crd2pulumi/issues/104)
  - NodeJS package names are now properly generated. [#70](https://github.com/pulumi/crd2pulumi/issues/70)
  - Dotnet packages now include the correct imports. [#49](https://github.com/pulumi/crd2pulumi/issues/49)
  - NodeJS object metadata types no longer accept undefined values. [#34](https://github.com/pulumi/crd2pulumi/issues/34)

## 1.4.0 (2024-05-29)

- Fix unpinned Kubernetes version in generated nodejs resources. [#121](https://github.com/pulumi/crd2pulumi/pull/121)
- Fix .NET generated code to use provider v4. [#134](https://github.com/pulumi/crd2pulumi/pull/134)
- Fix invalid generated code due to unnamed properties. [#135](https://github.com/pulumi/crd2pulumi/pull/135)
- Fix a panic when generating code with non-primitive defaults. [#136](https://github.com/pulumi/crd2pulumi/pull/136)
- Add Java generation support. [#129](https://github.com/pulumi/crd2pulumi/pull/129)

## 1.3.0 (2023-12-12)

- Fix: excluding files from unneededGoFiles was not working (<https://github.com/pulumi/crd2pulumi/pull/120>)
- Support kubernetes provider v4 (<https://github.com/pulumi/crd2pulumi/pull/119>)

## 1.2.5 (2023-05-31)

- Remove underscores in generated nested types (<https://github.com/pulumi/crd2pulumi/pull/114>)

## 1.2.4 (2023-03-23)

- Requires Go 1.19 or higher now to build
- Fix issue [#108](https://github.com/pulumi/crd2pulumi/issues/108) - crd2pulumi generator splits types apart into duplicate entires in pulumiTypes.go and pulumiTypes1.go

## 1.2.3 (2022-10-18)

- Fix issue [#43: crd properties with - in name](https://github.com/pulumi/crd2pulumi/issues/43) (<https://github.com/pulumi/crd2pulumi/pull/99>)

## 1.2.2 (2022-07-20)

- Fix regression that caused code in all languages to be generated regardless of selection.

## 1.2.1 (2022-07-19)

This release is a refactor with no user-affecting changes.

- Create public interface for codegen in the `pkg/codegen` namespace
  while placing internal utilities under `internal/`
- Simplify cobra usage, simplify program config substantially
- A new test env var, `TEST_SKIP_CLEANUP`, can be set to instruct the
  `crds_test.go` tests to not perform temp dir cleanup after the test
  run, for the purposes of investigating bad output during test failure.
  Generated code is now placed in temp dirs with friendly, identifiable
  names for each test case.
- General refactoring: removal of dead code, reorganizing functions into
  more appropriately named files or packages.
- Update to latest Pulumi SDK as well as update all other dependencies.
- Update to Go 1.18
- Upgrade to go 1.17 (<https://github.com/pulumi/crd2pulumi/pull/75>)

## 1.2.0 (2022-02-07)

- [python] Do not overwrite _utilities.py (<https://github.com/pulumi/crd2pulumi/pull/73/>)

## 1.1.0 (2022-01-04)

- Update to Pulumi v3.21.0 (<https://github.com/pulumi/crd2pulumi/pull/63>)
- Fix x-kubernetes-int-or-string precedence (<https://github.com/pulumi/crd2pulumi/pull/60>)
- Add generating CRD from URL (<https://github.com/pulumi/crd2pulumi/pull/62>)
