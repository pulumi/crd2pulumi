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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var languages = []string{"dotnet", "go", "nodejs", "python"}

const gkeManagedCertsUrl = "https://raw.githubusercontent.com/GoogleCloudPlatform/gke-managed-certs/master/deploy/managedcertificates-crd.yaml"

// execCrd2Pulumi runs the crd2pulumi binary in a temporary directory
func execCrd2Pulumi(t *testing.T, lang, path string) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("crd2pulumi-tests-%s-%s", time.Now().Local().Format(time.Kitchen), lang))
	require.NoError(t, os.MkdirAll(dir, 0755))
	tmpdir, err := os.MkdirTemp(dir, t.Name()+"-*")
	assert.Nil(t, err, "expected to create a temp dir for the CRD output")
	if os.Getenv("TEST_SKIP_CLEANUP") != "" {
		t.Logf("Test output tmp dir: %q\n", tmpdir)
	} else {
		defer os.RemoveAll(tmpdir)
	}
	binaryPath, err := filepath.Abs("../bin/crd2pulumi")
	if err != nil {
		panic(err)
	}
	args := []string{"--lang", lang, "--outputDir", tmpdir, "--force", path}
	t.Logf("running %q %q", binaryPath, args)
	crdCmd := exec.Command(binaryPath, args...)
	crdOut, err := crdCmd.CombinedOutput()
	t.Logf("output=\n%s\n", crdOut)
	assert.Nil(t, err, "expected no error running crd2pulumi")
}

// TestCRDsFromFile enumerates all CRD YAML files, and generates them in each language.
func TestCRDsFromFile(t *testing.T) {
	filepath.WalkDir("crds", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && (filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml") {
			for _, lang := range languages {
				execCrd2Pulumi(t, lang, path)
			}
		}
		return nil
	})
}

// TestCRDsFromUrl pulls the CRD YAML file from a URL and generates it in each language
func TestCRDsFromUrl(t *testing.T) {
	for _, lang := range languages {
		execCrd2Pulumi(t, lang, gkeManagedCertsUrl)
	}
}
