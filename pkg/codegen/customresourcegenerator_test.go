// Copyright 2016-2024, Pulumi Corporation.
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

package codegen

import (
	"testing"

	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetCRDDefaults(t *testing.T) {
	tests := []struct {
		name     string
		crd      extensionv1.CustomResourceDefinition
		expected extensionv1.CustomResourceDefinition
	}{
		{
			name: "Singular and ListKind are empty",
			crd: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind: "TestKind",
					},
				},
			},
			expected: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "testkind",
						ListKind: "TestKindList",
					},
				},
			},
		},
		{
			name: "Singular is set, ListKind is empty",
			crd: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "customsingular",
					},
				},
			},
			expected: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "customsingular",
						ListKind: "TestKindList",
					},
				},
			},
		},
		{
			name: "Singular is empty, ListKind is set",
			crd: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						ListKind: "CustomListKind",
					},
				},
			},
			expected: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "testkind",
						ListKind: "CustomListKind",
					},
				},
			},
		},
		{
			name: "Singular and ListKind are set",
			crd: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "customsingular",
						ListKind: "CustomListKind",
					},
				},
			},
			expected: extensionv1.CustomResourceDefinition{
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:     "TestKind",
						Singular: "customsingular",
						ListKind: "CustomListKind",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCRDDefaults(&tt.crd)
			if tt.crd.Spec.Names.Singular != tt.expected.Spec.Names.Singular {
				t.Errorf("expected Singular %s, got %s", tt.expected.Spec.Names.Singular, tt.crd.Spec.Names.Singular)
			}
			if tt.crd.Spec.Names.ListKind != tt.expected.Spec.Names.ListKind {
				t.Errorf("expected ListKind %s, got %s", tt.expected.Spec.Names.ListKind, tt.crd.Spec.Names.ListKind)
			}
		})
	}
}
func TestNewCustomResourceGenerator(t *testing.T) {
	tests := []struct {
		name     string
		crd      extensionv1.CustomResourceDefinition
		expected CustomResourceGenerator
		wantErr  bool
	}{
		{
			name: "Valid CRD",
			crd: extensionv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apiextensions.k8s.io/v1",
				},
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Group: "example.com",
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:   "TestKind",
						Plural: "testkinds",
					},
					Versions: []extensionv1.CustomResourceDefinitionVersion{
						{Name: "v1"},
					},
				},
			},
			expected: CustomResourceGenerator{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "TestKind",
				Plural:     "testkinds",
				Group:      "example.com",
				Versions:   []string{"v1"},
				GroupVersions: []string{
					"example.com/v1",
				},
				ResourceTokens: []string{
					"example.com:TestKind:v1",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid CRD with multiple versions",
			crd: extensionv1.CustomResourceDefinition{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apiextensions.k8s.io/v1",
				},
				Spec: extensionv1.CustomResourceDefinitionSpec{
					Group: "example.com",
					Names: extensionv1.CustomResourceDefinitionNames{
						Kind:   "TestKind",
						Plural: "testkinds",
					},
					Versions: []extensionv1.CustomResourceDefinitionVersion{
						{Name: "v1"},
						{Name: "v1alpha1"},
					},
				},
			},
			expected: CustomResourceGenerator{
				APIVersion: "apiextensions.k8s.io/v1",
				Kind:       "TestKind",
				Plural:     "testkinds",
				Group:      "example.com",
				Versions:   []string{"v1alpha1", "v1"},
				GroupVersions: []string{
					"example.com/v1",
					"example.com/v1alpha1",
				},
				ResourceTokens: []string{
					"example.com:TestKind:v1",
					"example.com:TestKind:v1alpha1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCustomResourceGenerator(tt.crd)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustomResourceGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.APIVersion != tt.expected.APIVersion {
					t.Errorf("expected APIVersion %s, got %s", tt.expected.APIVersion, got.APIVersion)
				}
				if got.Kind != tt.expected.Kind {
					t.Errorf("expected Kind %s, got %s", tt.expected.Kind, got.Kind)
				}
				if got.Plural != tt.expected.Plural {
					t.Errorf("expected Plural %s, got %s", tt.expected.Plural, got.Plural)
				}
				if got.Group != tt.expected.Group {
					t.Errorf("expected Group %s, got %s", tt.expected.Group, got.Group)
				}
				if len(got.Versions) != len(tt.expected.Versions) {
					t.Errorf("expected Versions %v, got %v", tt.expected.Versions, got.Versions)
				}
				if len(got.GroupVersions) != len(tt.expected.GroupVersions) {
					t.Errorf("expected GroupVersions %v, got %v", tt.expected.GroupVersions, got.GroupVersions)
				}
				if len(got.ResourceTokens) != len(tt.expected.ResourceTokens) {
					t.Errorf("expected ResourceTokens %v, got %v", tt.expected.ResourceTokens, got.ResourceTokens)
				}
			}
		})
	}
}
