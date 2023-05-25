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
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var languages = []string{"dotnet", "go", "nodejs", "python"}

const gkeManagedCertsUrl = "https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/master/deploy/managedcertificates-crd.yaml"

// execCrd2Pulumi runs the crd2pulumi binary in a temporary directory
func execCrd2Pulumi(t *testing.T, lang, path string, additionalValidation func(t *testing.T, path string)) {
	tmpdir, err := ioutil.TempDir("", "crd2pulumi_test")
	assert.Nil(t, err, "expected to create a temp dir for the CRD output")
	t.Cleanup(func() {
		t.Logf("removing temp dir %q for %s test", tmpdir, lang)
		os.RemoveAll(tmpdir)
	})
	langFlag := fmt.Sprintf("--%sPath", lang) // e.g. --dotnetPath
	binaryPath, err := filepath.Abs("../bin/crd2pulumi")
	if err != nil {
		t.Fatalf("unable to create absolute path to binary: %s", err)
	}

	t.Logf("%s %s=%s %s: running", binaryPath, langFlag, tmpdir, path)
	crdCmd := exec.Command(binaryPath, langFlag, tmpdir, "--force", path)
	crdOut, err := crdCmd.CombinedOutput()
	t.Logf("%s %s=%s %s: output=\n%s", binaryPath, langFlag, tmpdir, path, crdOut)
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
	for _, lang := range languages {
		lang := lang
		t.Run(lang, func(t *testing.T) {
			t.Parallel()
			execCrd2Pulumi(t, lang, gkeManagedCertsUrl, nil)
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
		pythonInputs, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("expected to read generated Python code: %s", err)
		}
		assert.Contains(t, string(pythonInputs), "NetworkPolicySpecAppsIncomingArgs", "expected to find camelCased field name in generated Python code")
		assert.NotContains(t, string(pythonInputs), "NetworkPolicySpecApps_incomingArgs", "expected to not find underscored field name in generated Python code")

		// Ensure outputs are camelCased.
		filename = filepath.Join(path, "pulumi_crds", "juice", "v1alpha1", "outputs.py")
		t.Logf("validating underscored field names in %s", filename)
		pythonInputs, err = ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("expected to read generated Python code: %s", err)
		}
		assert.Contains(t, string(pythonInputs), "NetworkPolicySpecAppsIncoming", "expected to find camelCased field name in generated Python code")
		assert.NotContains(t, string(pythonInputs), "NetworkPolicySpecApps_incoming", "expected to not find underscored field name in generated Python code")
	}

	execCrd2Pulumi(t, "python", "crds/underscored-types/networkpolicy.yaml", validateUnderscore)
}
