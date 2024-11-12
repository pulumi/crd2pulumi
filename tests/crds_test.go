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

package tests

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pulumi/crd2pulumi/cmd"
	"github.com/pulumi/crd2pulumi/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var languages = []string{"dotnet", "go", "nodejs", "python", "java"}

// execCrd2Pulumi runs the crd2pulumi binary in a temporary directory
func execCrd2Pulumi(t *testing.T, lang, path string, additionalValidation func(t *testing.T, path string)) {
	tmpdir, err := os.MkdirTemp("", "crd2pulumi_test")
	assert.Nil(t, err, "expected to create a temp dir for the CRD output")
	t.Cleanup(func() {
		t.Logf("removing temp dir %q for %s test", tmpdir, lang)
		os.RemoveAll(tmpdir)
	})
	langFlag := fmt.Sprintf("--%sPath", lang) // e.g. --dotnetPath

	cmd := cmd.New()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetArgs([]string{langFlag, tmpdir, "--force", path})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	t.Logf("crd2pulumi %s=%s %s: running", langFlag, tmpdir, path)
	err = cmd.Execute()
	t.Logf("%s=%s %s: output=\n%s", langFlag, tmpdir, path, stdout.String()+stderr.String())
	if err != nil {
		t.Fatalf("expected crd2pulumi for '%s=%s %s' to succeed", langFlag, tmpdir, path)
	}

	// Run additional validation if provided.
	if additionalValidation != nil {
		additionalValidation(t, tmpdir)
	}
}

// TestCRDsFromFile enumerates all CRD YAML files, and generates them in each language.
func TestCRDsFromFile(t *testing.T) {
	filepath.WalkDir("crds", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && (filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml") {
			for _, lang := range languages {
				lang := lang
				name := fmt.Sprintf("%s-%s", lang, filepath.Base(path))
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					execCrd2Pulumi(t, lang, path, nil)
				})
			}
		}
		return nil
	})
}

// TestCRDsFromUrl pulls the CRD YAML file from a URL and generates it in each language
func TestCRDsFromUrl(t *testing.T) {
	validateNodeCompiles := func(t *testing.T, path string) {
		withDir(t, path, func() {
			runRequireNoError(t, exec.Command("npm", "install"))
			runRequireNoError(t, exec.Command("npm", "run", "build"))
		})
	}

	validateGolangCompiles := func(t *testing.T, path string) {
		withDir(t, path, func() {
			runRequireNoError(t, exec.Command("go", "mod", "init", "fakepackage"))
			runRequireNoError(t, exec.Command("go", "mod", "tidy"))
			runRequireNoError(t, exec.Command("go", "vet", "./..."))
		})
	}

	validateDotnetCompiles := func(t *testing.T, path string) {
		withDir(t, path, func() {
			runRequireNoError(t, exec.Command("dotnet", "build"))
		})
	}

	// TODO(#145): Also run compilation tests for java and python.
	compileValidationFn := map[string]func(t *testing.T, path string){
		"nodejs": validateNodeCompiles,
		"go":     validateGolangCompiles,
		"python": nil,
		"java":   nil,
		"dotnet": validateDotnetCompiles,
	}

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "GKEManagedCerts",
			url:  "https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/c514101/deploy/managedcertificates-crd.yaml",
		},
		{
			name: "VictoriaMetrics",
			url:  "https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/fdb7dfe/charts/victoria-metrics-operator/crd.yaml",
		},
		{
			name: "GatewayClasses",
			url:  "https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v0.3.0/config/crd/bases/networking.x-k8s.io_gatewayclasses.yaml",
		},
		{
			name: "Contours",
			url:  "https://raw.githubusercontent.com/projectcontour/contour-operator/f8c07498803d062e30c255976270cbc82cd619b0/config/crd/bases/operator.projectcontour.io_contours.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/141
			name: "ElementDeployment",
			url:  "https://raw.githubusercontent.com/element-hq/ess-starter-edition-core/d7e792bf8a872f06f02f59d807a1c16ee933862b/roles/elementdeployment/files/elementdeployment-schema.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/142
			name: "Keycloak",
			url:  "https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/25.0.4/kubernetes/keycloaks.k8s.keycloak.org-v1.yml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/115
			name: "CertManager",
			url:  "https://gist.githubusercontent.com/RouxAntoine/b7dfb9ce327a4ad40a76ff6552c7fd5e/raw/4b5922da11643e14d04e6b52f7a0fca982e4dace/1-crds.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/104
			name: "TracingPolicies",
			url:  "https://raw.githubusercontent.com/cilium/tetragon/v1.2.0/install/kubernetes/tetragon/crds-yaml/cilium.io_tracingpolicies.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/104
			name: "Argo Rollouts",
			url:  "https://raw.githubusercontent.com/argoproj/argo-rollouts/74c1a947ab36670ae01a45993a0c5abb44af4677/manifests/crds/rollout-crd.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/92
			name: "Grafana",
			url:  "https://raw.githubusercontent.com/bitnami/charts/main/bitnami/grafana-operator/crds/grafanas.integreatly.org.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/70
			name: "Percona",
			url:  "https://raw.githubusercontent.com/percona/percona-server-mongodb-operator/main/deploy/crd.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/49
			name: "Traefik",
			url:  "https://raw.githubusercontent.com/traefik/traefik/eb99c8c/docs/content/reference/dynamic-configuration/traefik.io_traefikservices.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/29
			name: "Istio",
			url:  "https://raw.githubusercontent.com/istio/istio/c132663/manifests/charts/base/crds/crd-all.gen.yaml",
		},
		{
			name: "Argo Application Set",
			url:  "https://raw.githubusercontent.com/argoproj/argo-cd/master/manifests/crds/applicationset-crd.yaml",
		},
		{
			// https://github.com/pulumi/crd2pulumi/issues/147
			name: "Prometheus Operator",
			url:  "https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.76.2/bundle.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, lang := range languages {
				t.Run(lang, func(t *testing.T) {
					if lang == "dotnet" {
						if tt.name == "CertManager" || tt.name == "GKEManagedCerts" {
							t.Skip("Skipping compilation for dotnet. See https://github.com/pulumi/crd2pulumi/issues/17")
						}

						if tt.name == "Percona" {
							t.Skip("Skipping dotnet compilation for Percona as we generate invalid code with hyphens that are not allowed in C# identifiers.")
						}

					}

					execCrd2Pulumi(t, lang, tt.url, compileValidationFn[lang])
				})
			}
		})
	}
}

// TestCRDsWithUnderscore tests that CRDs with underscores field names are camelCased for the
// generated types. Currently this test only runs for Python, and we're hardcoding the field name
// detection logic in the test for simplicity. This is brittle and we should improve this in the
// future.
// TODO: properly detect field names in the generated Python code instead of grep'ing for them.
func TestCRDsWithUnderscore(t *testing.T) {
	// Callback function to run additional validation on the generated Python code after running
	// crd2pulumi.
	validateUnderscore := func(t *testing.T, path string) {
		// Ensure inputs are camelCased.
		filename := filepath.Join(path, "pulumi_crds", "juice", "v1alpha1", "_inputs.py")
		t.Logf("validating underscored field names in %s", filename)
		pythonInputs, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("expected to read generated Python code: %s", err)
		}
		assert.Contains(t, string(pythonInputs), "NetworkPolicySpecAppsIncomingArgs", "expected to find camelCased field name in generated Python code")
		assert.NotContains(t, string(pythonInputs), "NetworkPolicySpecApps_incomingArgs", "expected to not find underscored field name in generated Python code")

		// Ensure outputs are camelCased.
		filename = filepath.Join(path, "pulumi_crds", "juice", "v1alpha1", "outputs.py")
		t.Logf("validating underscored field names in %s", filename)
		pythonInputs, err = os.ReadFile(filename)
		if err != nil {
			t.Fatalf("expected to read generated Python code: %s", err)
		}
		assert.Contains(t, string(pythonInputs), "NetworkPolicySpecAppsIncoming", "expected to find camelCased field name in generated Python code")
		assert.NotContains(t, string(pythonInputs), "NetworkPolicySpecApps_incoming", "expected to not find underscored field name in generated Python code")
	}

	execCrd2Pulumi(t, "python", "crds/underscored-types/networkpolicy.yaml", validateUnderscore)
}

func TestKubernetesVersionNodeJs(t *testing.T) {
	validateVersion := func(t *testing.T, path string) {
		// enter and build the generated package
		withDir(t, path, func() {
			runRequireNoError(t, exec.Command("npm", "install"))
			runRequireNoError(t, exec.Command("npm", "run", "build"))

			// extract the version returned by a resource
			appendFile(t, "bin/index.js", "\nconsole.log((new k8sversion.test.TestResource('test')).__version)")

			version, err := exec.Command("node", "bin/index.js").Output()
			require.NoError(t, err)
			assert.Equal(t, codegen.KubernetesProviderVersion+"\n", string(version))
		})
	}

	execCrd2Pulumi(t, "nodejs", "crds/k8sversion/mock_crd.yaml", validateVersion)
}

func TestNodeJsObjectMeta(t *testing.T) {
	validateVersion := func(t *testing.T, path string) {
		// enter and build the generated package
		withDir(t, path, func() {
			runRequireNoError(t, exec.Command("npm", "install"))
			runRequireNoError(t, exec.Command("npm", "run", "build"))

			filename := filepath.Join(path, "k8sversion", "test", "testResource.ts")
			t.Logf("validating objectmeta type in %s", filename)

			testResource, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("expected to read generated NodeJS code: %s", err)
			}

			assert.Contains(t, string(testResource), "public readonly metadata!: pulumi.Output<outputs.meta.v1.ObjectMeta>;", "expected metadata output type")
			assert.Contains(t, string(testResource), "metadata?: pulumi.Input<inputs.meta.v1.ObjectMeta>;", "expected metadata input type")
		})
	}

	execCrd2Pulumi(t, "nodejs", "crds/k8sversion/mock_crd.yaml", validateVersion)
}

func withDir(t *testing.T, dir string, f func()) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(pwd)

	require.NoError(t, os.Chdir(dir))

	f()
}

func appendFile(t *testing.T, filename, content string) {
	// extract the version returned by a resource
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.WriteString(content)
}

func runRequireNoError(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(bytes))
	}
	require.NoError(t, err)
}
