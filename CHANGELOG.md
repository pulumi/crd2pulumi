CHANGELOG
=========

## Unreleased
- Remove underscores in generated nested types (https://github.com/pulumi/crd2pulumi/pull/114)

## 1.2.4 (2023-03-23)
- Requires Go 1.19 or higher now to build
- Fix issue [#108](https://github.com/pulumi/crd2pulumi/issues/108) - crd2pulumi generator splits types apart into duplicate entires in pulumiTypes.go and pulumiTypes1.go

## 1.2.3 (2022-10-18)
- Fix issue [#43: crd properties with - in name](https://github.com/pulumi/crd2pulumi/issues/43) (https://github.com/pulumi/crd2pulumi/pull/99)

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
- Upgrade to go 1.17 (https://github.com/pulumi/crd2pulumi/pull/75)

## 1.2.0 (2022-02-07)
- [python] Do not overwrite _utilities.py (https://github.com/pulumi/crd2pulumi/pull/73/)

## 1.1.0 (2022-01-04)

- Update to Pulumi v3.21.0 (https://github.com/pulumi/crd2pulumi/pull/63)
- Fix x-kubernetes-int-or-string precedence (https://github.com/pulumi/crd2pulumi/pull/60)
- Add generating CRD from URL (https://github.com/pulumi/crd2pulumi/pull/62)
