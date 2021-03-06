/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resmap

import (
	"reflect"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/loader/loadertest"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var cmap = schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}

func TestNewFromConfigMaps(t *testing.T) {
	type testCase struct {
		description string
		input       []types.ConfigMapArgs
		filepath    string
		content     string
		expected    ResMap
	}

	l := loadertest.NewFakeLoader("/home/seans/project/")
	testCases := []testCase{
		{
			description: "construct config map from env",
			input: []types.ConfigMapArgs{
				{
					Name: "envConfigMap",
					DataSources: types.DataSources{
						EnvSource: "app.env",
					},
				},
			},
			filepath: "/home/seans/project/app.env",
			content:  "DB_USERNAME=admin\nDB_PASSWORD=somepw",
			expected: ResMap{
				resource.NewResId(cmap, "envConfigMap"): resource.NewBehaviorlessResource(
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "envConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"DB_USERNAME": "admin",
								"DB_PASSWORD": "somepw",
							},
						},
					}),
			},
		},
		{
			description: "construct config map from file",
			input: []types.ConfigMapArgs{{
				Name: "fileConfigMap",
				DataSources: types.DataSources{
					FileSources: []string{"app-init.ini"},
				},
			},
			},
			filepath: "/home/seans/project/app-init.ini",
			content:  "FOO=bar\nBAR=baz\n",
			expected: ResMap{
				resource.NewResId(cmap, "fileConfigMap"): resource.NewBehaviorlessResource(
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "fileConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"app-init.ini": `FOO=bar
BAR=baz
`,
							},
						},
					}),
			},
		},
		{
			description: "construct config map from literal",
			input: []types.ConfigMapArgs{
				{
					Name: "literalConfigMap",
					DataSources: types.DataSources{
						LiteralSources: []string{"a=x", "b=y"},
					},
				},
			},
			expected: ResMap{
				resource.NewResId(cmap, "literalConfigMap"): resource.NewBehaviorlessResource(
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":              "literalConfigMap",
								"creationTimestamp": nil,
							},
							"data": map[string]interface{}{
								"a": "x",
								"b": "y",
							},
						},
					}),
			},
		},
		// TODO: add testcase for data coming from multiple sources like
		// files/literal/env etc.
	}

	for _, tc := range testCases {

		if ferr := l.AddFile(tc.filepath, []byte(tc.content)); ferr != nil {
			t.Fatalf("Error adding fake file: %v\n", ferr)
		}
		r, err := NewResMapFromConfigMapArgs(l, tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Fatalf("in testcase: %q got:\n%+v\n expected:\n%+v\n", tc.description, r, tc.expected)
		}
	}
}
