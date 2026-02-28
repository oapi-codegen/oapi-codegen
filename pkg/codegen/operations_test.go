// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package codegen

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestIsJson(t *testing.T) {
	type test struct {
		name       string
		mediaTypes []string
		want       bool
	}

	suite := []test{
		{
			name:       "When no MediaType, returns false",
			mediaTypes: []string{},
			want:       false,
		},
		{
			name:       "When not a JSON MediaType, returns false",
			mediaTypes: []string{"application/pdf"},
			want:       false,
		},
		{
			name:       "When more than one MediaTypes, returns false",
			mediaTypes: []string{"application/pdf", "application/json"},
			want:       false,
		},
		{
			name:       "When MediaType ends with json, but isn't JSON, returns false",
			mediaTypes: []string{"application/notjson"},
			want:       false,
		},
		{
			name:       "When MediaType is application/json, returns true",
			mediaTypes: []string{"application/json"},
			want:       true,
		},
		{
			name:       "When MediaType is application/json-patch+json, returns true",
			mediaTypes: []string{"application/json-patch+json"},
			want:       true,
		},
		{
			name:       "When MediaType is application/vnd.api+json, returns true",
			mediaTypes: []string{"application/vnd.api+json"},
			want:       true,
		},
	}
	for _, test := range suite {
		t.Run(test.name, func(t *testing.T) {
			pd := ParameterDefinition{
				Spec: &openapi3.Parameter{
					Content: make(map[string]*openapi3.MediaType),
				},
			}
			for _, mediaType := range test.mediaTypes {
				pd.Spec.Content[mediaType] = nil
			}

			got := pd.IsJson()

			if got != test.want {
				t.Fatalf("IsJson validation failed. Want [%v] Got [%v]", test.want, got)
			}

		})
	}
}

func TestGenerateDefaultOperationID(t *testing.T) {
	type test struct {
		op      string
		path    string
		want    string
		wantErr bool
	}

	suite := []test{
		{
			op:      http.MethodGet,
			path:    "/v1/foo/bar",
			want:    "GetV1FooBar",
			wantErr: false,
		},
		{
			op:      http.MethodGet,
			path:    "/v1/foo/bar/",
			want:    "GetV1FooBar",
			wantErr: false,
		},
		{
			op:      http.MethodPost,
			path:    "/v1",
			want:    "PostV1",
			wantErr: false,
		},
		{
			op:      http.MethodPost,
			path:    "v1",
			want:    "PostV1",
			wantErr: false,
		},
		{
			path:    "v1",
			want:    "",
			wantErr: true,
		},
		{
			path:    "",
			want:    "PostV1",
			wantErr: true,
		},
	}

	for _, test := range suite {
		got, err := generateDefaultOperationID(test.op, test.path, ToCamelCase)
		if err != nil {
			if !test.wantErr {
				t.Fatalf("did not expected error but got %v", err)
			}
		}

		if test.wantErr {
			return
		}
		if got != test.want {
			t.Fatalf("Operation ID generation error. Want [%v] Got [%v]", test.want, got)
		}
	}
}

func TestJsonTag(t *testing.T) {
	t.Run("required param with no extra tags", func(t *testing.T) {
		pd := ParameterDefinition{
			ParamName: "foo",
			Required:  true,
			Spec:      &openapi3.Parameter{},
		}
		assert.Equal(t, "`json:\"foo\"`", pd.JsonTag())
	})

	t.Run("optional param with no extra tags", func(t *testing.T) {
		pd := ParameterDefinition{
			ParamName: "foo",
			Required:  false,
			Spec:      &openapi3.Parameter{},
		}
		assert.Equal(t, "`json:\"foo,omitempty\"`", pd.JsonTag())
	})

	t.Run("extra tags at parameter level", func(t *testing.T) {
		pd := ParameterDefinition{
			ParamName: "foo",
			Required:  true,
			Spec: &openapi3.Parameter{
				Extensions: map[string]any{
					"x-oapi-codegen-extra-tags": map[string]any{
						"validate": "required",
						"db":       "foo_col",
					},
				},
			},
		}
		assert.Equal(t, "`db:\"foo_col\" json:\"foo\" validate:\"required\"`", pd.JsonTag())
	})

	t.Run("extra tags at schema level", func(t *testing.T) {
		pd := ParameterDefinition{
			ParamName: "foo",
			Required:  true,
			Spec: &openapi3.Parameter{
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Extensions: map[string]any{
							"x-oapi-codegen-extra-tags": map[string]any{
								"validate": "required",
							},
						},
					},
				},
			},
		}
		assert.Equal(t, "`json:\"foo\" validate:\"required\"`", pd.JsonTag())
	})

	t.Run("parameter level takes precedence over schema level", func(t *testing.T) {
		pd := ParameterDefinition{
			ParamName: "foo",
			Required:  true,
			Spec: &openapi3.Parameter{
				Extensions: map[string]any{
					"x-oapi-codegen-extra-tags": map[string]any{
						"validate": "param-level",
					},
				},
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Extensions: map[string]any{
							"x-oapi-codegen-extra-tags": map[string]any{
								"validate": "schema-level",
								"db":       "foo_col",
							},
						},
					},
				},
			},
		}
		// Parameter-level "validate" wins, schema-level "db" is kept
		assert.Equal(t, "`db:\"foo_col\" json:\"foo\" validate:\"param-level\"`", pd.JsonTag())
	})
}
