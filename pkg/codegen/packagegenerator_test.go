// Copyright 2024, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package codegen

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPackagesFromSource(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "argocd.yaml"))
	require.NoError(t, err)

	gen, err := ReadPackagesFromSource("0.0.1", []io.ReadCloser{f})
	require.NoError(t, err)

	rollout, ok := gen.Types["kubernetes:argoproj.io/v1alpha1:Rollout"]
	require.True(t, ok)

	// Required inputs

	_, ok = rollout.Properties["spec"]
	assert.True(t, ok)
	assert.Contains(t, rollout.Required, "spec")

	_, ok = rollout.Properties["metadata"]
	assert.True(t, ok)
	assert.Contains(t, rollout.Required, "metadata")

	_, ok = rollout.Properties["apiVersion"]
	assert.True(t, ok)
	assert.Contains(t, rollout.Required, "apiVersion")

	_, ok = rollout.Properties["kind"]
	assert.True(t, ok)
	assert.Contains(t, rollout.Required, "kind")

	// The CRD isn't able to communicate that "status" is a required output.
	// Verify it's at least not a required input for now. _, ok =
	// rollout.Properties["status"]
	assert.True(t, ok)
	assert.NotContains(t, rollout.Required, "status")
	assert.Subset(t, rollout.Language["nodejs"], []byte(`"status"`))

	// p-k branch has 32 types, 10 resources
	pkg := gen.SchemaPackage(true)
	assert.Equal(t, pkg.Resources, 1)
}
