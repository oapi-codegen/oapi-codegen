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
	"go/format"
	"net/http"
	"strings"
	"testing"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		got, err := generateDefaultOperationID(test.op, test.path)
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

func TestSummaryAsComment(t *testing.T) {
	tests := []struct {
		name    string
		summary string
		prefix  string
		want    string
	}{
		{
			name:    "empty summary returns empty string",
			summary: "",
			prefix:  "GetFoo",
			want:    "",
		},
		{
			name:    "single line summary with prefix",
			summary: "Get a foo",
			prefix:  "GetFoo",
			want:    "// GetFoo Get a foo",
		},
		{
			name:    "single line summary with empty prefix",
			summary: "Get a foo",
			prefix:  "",
			want:    "// Get a foo",
		},
		{
			name:    "multiline summary with prefix only prefixes first line",
			summary: "Get a foo\nReturns the foo resource",
			prefix:  "GetFoo",
			want:    "// GetFoo Get a foo\n// Returns the foo resource",
		},
		{
			name:    "trailing newline is trimmed",
			summary: "Get a foo\n\r",
			prefix:  "GetFoo",
			want:    "// GetFoo Get a foo",
		},
		{
			name:    "CRLF line endings are treated as newlines",
			summary: "Get a foo\r\nReturns the foo resource",
			prefix:  "GetFoo",
			want:    "// GetFoo Get a foo\n// Returns the foo resource",
		},
		{
			name:    "bare CR is treated as a newline, not left in output",
			summary: "Get a foo\rReturns the foo resource",
			prefix:  "GetFoo",
			want:    "// GetFoo Get a foo\n// Returns the foo resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := OperationDefinition{Summary: tt.summary}
			got := op.SummaryAsComment(tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeprecationComment(t *testing.T) {
	tests := []struct {
		name string
		op   *OperationDefinition
		want string
	}{
		{
			name: "nil spec returns empty string",
			op:   &OperationDefinition{Spec: nil},
			want: "",
		},
		{
			name: "non-deprecated operation returns empty string",
			op: &OperationDefinition{
				Spec: &openapi3.Operation{Deprecated: false},
			},
			want: "",
		},
		{
			name: "deprecated operation without x-deprecated-reason uses default message",
			op: &OperationDefinition{
				Spec: &openapi3.Operation{Deprecated: true},
			},
			want: "// Deprecated: this operation has been marked as deprecated upstream, but no `x-deprecated-reason` was set",
		},
		{
			name: "deprecated operation with x-deprecated-reason uses the reason",
			op: &OperationDefinition{
				Spec: &openapi3.Operation{
					Deprecated: true,
					Extensions: map[string]any{
						"x-deprecated-reason": "Use /v2/foo instead.",
					},
				},
			},
			want: "// Deprecated: Use /v2/foo instead.",
		},
		{
			name: "deprecated operation with non-string x-deprecated-reason falls back to default message",
			op: &OperationDefinition{
				Spec: &openapi3.Operation{
					Deprecated: true,
					Extensions: map[string]any{
						"x-deprecated-reason": 42,
					},
				},
			},
			want: "// Deprecated: this operation has been marked as deprecated upstream, but no `x-deprecated-reason` was set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.op.DeprecationComment())
		})
	}
}

func TestOperationDefinition_GenerateFunctionComment(t *testing.T) {
	opWithBody := func(summary string) OperationDefinition {
		return OperationDefinition{
			OperationId: "GetFoo",
			Method:      "GET",
			Path:        "/foo",
			Summary:     summary,
			Spec: &openapi3.Operation{
				RequestBody: &openapi3.RequestBodyRef{},
			},
		}
	}
	opWithoutBody := func(summary string) OperationDefinition {
		return OperationDefinition{
			OperationId: "GetFoo",
			Method:      "GET",
			Path:        "/foo",
			Summary:     summary,
			Spec:        &openapi3.Operation{},
		}
	}
	opWithDescription := func(summary, description string) OperationDefinition {
		return OperationDefinition{
			OperationId: "GetFoo",
			Method:      "GET",
			Path:        "/foo",
			Summary:     summary,
			Spec:        &openapi3.Operation{Description: description},
		}
	}
	opWithBodyAndDescription := func(summary, description string) OperationDefinition {
		return OperationDefinition{
			OperationId: "GetFoo",
			Method:      "GET",
			Path:        "/foo",
			Summary:     summary,
			Spec: &openapi3.Operation{
				Description: description,
				RequestBody: &openapi3.RequestBodyRef{},
			},
		}
	}

	tests := []struct {
		name                    string
		op                      OperationDefinition
		originalFunctionName    string
		functionSuffix          string
		isFunctionWithResponses bool
		want                    string
	}{
		{
			name:                    "no summary, no body, not with responses",
			op:                      opWithoutBody(""),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want:                    "// GetFoo performs a GET /foo (the `GetFoo` operationId) request.",
		},
		{
			name:                    "no summary, no body, with responses",
			op:                      opWithoutBody(""),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse performs a GET /foo (the `GetFoo` operationId) request.\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).",
		},
		{
			name:                    "no summary, has body, not with responses",
			op:                      opWithBody(""),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo performs a GET /foo (the `GetFoo` operationId) request,\n" +
				"// with any type of body and a specified content type.",
		},
		{
			name:                    "no summary, has body, with responses",
			op:                      opWithBody(""),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse performs a GET /foo (the `GetFoo` operationId) request,\n" +
				"// with any type of body and a specified content type.\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).",
		},
		{
			name:                    "has summary, no body, not with responses",
			op:                      opWithoutBody("Get a foo"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "has summary, no body, with responses",
			op:                      opWithoutBody("Get a foo"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse Get a foo\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "has summary, has body, not with responses",
			op:                      opWithBody("Get a foo"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"//\n" +
				"// Takes any type of body and a specified content type.\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "has summary, has body, with responses",
			op:                      opWithBody("Get a foo"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse Get a foo\n" +
				"//\n" +
				"// Takes any type of body and a specified content type, and returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "no summary, description, no body, not with responses",
			op:                      opWithDescription("", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo performs a GET /foo (the `GetFoo` operationId) request.\n" +
				"//\n" +
				"// Detailed description.",
		},
		{
			name:                    "no summary, description, no body, with responses",
			op:                      opWithDescription("", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse performs a GET /foo (the `GetFoo` operationId) request.\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).",
		},
		{
			name:                    "no summary, description, has body, not with responses",
			op:                      opWithBodyAndDescription("", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo performs a GET /foo (the `GetFoo` operationId) request,\n" +
				"// with any type of body and a specified content type.\n" +
				"//\n" +
				"// Detailed description.",
		},
		{
			name:                    "has summary, description, no body, not with responses",
			op:                      opWithDescription("Get a foo", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "has summary, description, no body, with responses",
			op:                      opWithDescription("Get a foo", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse Get a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "has summary, description, has body, with responses",
			op:                      opWithBodyAndDescription("Get a foo", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse Get a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Takes any type of body and a specified content type, and returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "no summary, description, has body, with responses",
			op:                      opWithBodyAndDescription("", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "WithResponse",
			isFunctionWithResponses: true,
			want: "// GetFooWithResponse performs a GET /foo (the `GetFoo` operationId) request,\n" +
				"// with any type of body and a specified content type.\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Returns a wrapper object for the known response body format(s).",
		},
		{
			name:                    "has summary, description, has body, not with responses",
			op:                      opWithBodyAndDescription("Get a foo", "Detailed description."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Takes any type of body and a specified content type.\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "multiline description is rendered correctly",
			op:                      opWithDescription("Get a foo", "Line one.\nLine two."),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"//\n" +
				"// Line one.\n" +
				"// Line two.\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name: "multiline summary strips embedded newlines in individual parts",
			op: OperationDefinition{
				OperationId: "GetFoo",
				Method:      "GET",
				Path:        "/foo",
				Summary:     "Line one\nLine two",
				Spec:        &openapi3.Operation{},
			},
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Line one\n" +
				"// Line two\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "CRLF in summary does not inject code",
			op:                      opWithoutBody("Get a foo\r\nLine two"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"// Line two\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
		{
			name:                    "bare CR in summary does not inject code",
			op:                      opWithoutBody("Get a foo\rLine two"),
			originalFunctionName:    "GetFoo",
			functionSuffix:          "",
			isFunctionWithResponses: false,
			want: "// GetFoo Get a foo\n" +
				"// Line two\n" +
				"//\n" +
				"// Corresponds with GET /foo (the `GetFoo` operationId).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.op.GenerateFunctionComment(tt.originalFunctionName, tt.functionSuffix, tt.isFunctionWithResponses)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestBodyDefinition_GenerateFunctionComment(t *testing.T) {
	parentWithSummary := OperationDefinition{
		OperationId: "CreateFoo",
		Method:      "POST",
		Path:        "/foo",
		Summary:     "Create a foo",
		Spec:        &openapi3.Operation{},
	}
	parentNoSummary := OperationDefinition{
		OperationId: "CreateFoo",
		Method:      "POST",
		Path:        "/foo",
		Spec:        &openapi3.Operation{},
	}
	parentWithSummaryAndDescription := OperationDefinition{
		OperationId: "CreateFoo",
		Method:      "POST",
		Path:        "/foo",
		Summary:     "Create a foo",
		Spec:        &openapi3.Operation{Description: "Detailed description."},
	}
	parentNoSummaryWithDescription := OperationDefinition{
		OperationId: "CreateFoo",
		Method:      "POST",
		Path:        "/foo",
		Spec:        &openapi3.Operation{Description: "Detailed description."},
	}
	body := RequestBodyDefinition{
		ContentType: "application/json",
	}

	tests := []struct {
		name                    string
		body                    RequestBodyDefinition
		originalFunctionName    string
		parent                  OperationDefinition
		functionSuffix          string
		isFunctionWithResponses bool
		want                    string
	}{
		{
			name:                    "no summary, not with responses",
			body:                    body,
			parent:                  parentNoSummary,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody performs a POST /foo (the `CreateFoo` operationId) request.\n" +
				"// Takes a body of the `application/json` content type.",
		},
		{
			name:                    "no summary, with responses",
			body:                    body,
			parent:                  parentNoSummary,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBodyWithResponse",
			isFunctionWithResponses: true,
			want: "// CreateFooWithJSONBodyWithResponse performs a POST /foo (the `CreateFoo` operationId) request.\n" +
				"// Takes a body of the `application/json` content type, and returns a wrapper object for the known response body format(s).",
		},
		{
			name:                    "has summary, not with responses",
			body:                    body,
			parent:                  parentWithSummary,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody Create a foo\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type.\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
		{
			name:                    "has summary, with responses",
			body:                    body,
			parent:                  parentWithSummary,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBodyWithResponse",
			isFunctionWithResponses: true,
			want: "// CreateFooWithJSONBodyWithResponse Create a foo\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type, and returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
		{
			name:                    "no summary, description, not with responses",
			body:                    body,
			parent:                  parentNoSummaryWithDescription,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody performs a POST /foo (the `CreateFoo` operationId) request.\n" +
				"// Takes a body of the `application/json` content type.\n" +
				"//\n" +
				"// Detailed description.",
		},
		{
			name:                    "no summary, description, with responses",
			body:                    body,
			parent:                  parentNoSummaryWithDescription,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBodyWithResponse",
			isFunctionWithResponses: true,
			want: "// CreateFooWithJSONBodyWithResponse performs a POST /foo (the `CreateFoo` operationId) request.\n" +
				"// Takes a body of the `application/json` content type, and returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Detailed description.",
		},
		{
			name:                    "has summary, description, not with responses",
			body:                    body,
			parent:                  parentWithSummaryAndDescription,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody Create a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type.\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
		{
			name:                    "has summary, description, with responses",
			body:                    body,
			parent:                  parentWithSummaryAndDescription,
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBodyWithResponse",
			isFunctionWithResponses: true,
			want: "// CreateFooWithJSONBodyWithResponse Create a foo\n" +
				"//\n" +
				"// Detailed description.\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type, and returns a wrapper object for the known response body format(s).\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
		{
			name: "CRLF in parent summary does not inject code",
			body: body,
			parent: OperationDefinition{
				OperationId: "CreateFoo",
				Method:      "POST",
				Path:        "/foo",
				Summary:     "Create a foo\r\nLine two",
				Spec:        &openapi3.Operation{},
			},
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody Create a foo\n" +
				"// Line two\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type.\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
		{
			name: "bare CR in parent summary does not inject code",
			body: body,
			parent: OperationDefinition{
				OperationId: "CreateFoo",
				Method:      "POST",
				Path:        "/foo",
				Summary:     "Create a foo\rLine two",
				Spec:        &openapi3.Operation{},
			},
			originalFunctionName:    "CreateFoo",
			functionSuffix:          "WithJSONBody",
			isFunctionWithResponses: false,
			want: "// CreateFooWithJSONBody Create a foo\n" +
				"// Line two\n" +
				"//\n" +
				"// Takes a body of the `application/json` content type.\n" +
				"//\n" +
				"// Corresponds with POST /foo (the `CreateFoo` operationId).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.body.GenerateFunctionComment(tt.originalFunctionName, tt.parent, tt.functionSuffix, tt.isFunctionWithResponses)
			assert.Equal(t, tt.want, got)
		})
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

// TestGenerateFunctionComment_GofmtHeading tests that a description emitted as
// a standalone one-line paragraph with no terminal punctuation causes gofmt
// (Go ≥ 1.19) to promote it to an old-style heading ("// # <text>").  This
// applies both when summary == description (fix #1: drop the duplicate) and
// when they differ but the description still looks like a heading (fix #2:
// append a period).  These tests pin the current broken behaviour so we can
// verify both fixes eliminate the "// # " lines.
func TestGenerateFunctionComment_GofmtHeading(t *testing.T) {
	tests := []struct {
		name    string
		comment string
	}{
		{
			name: "OperationDefinition: summary equals description",
			comment: func() string {
				op := OperationDefinition{
					OperationId: "ValidatePets",
					Method:      "POST",
					Path:        "/pets:validate",
					Summary:     "Validate pets",
					Spec:        &openapi3.Operation{Description: "Validate pets"},
				}
				return op.GenerateFunctionComment("ValidatePets", "", false)
			}(),
		},
		{
			name: "RequestBodyDefinition: summary equals description",
			comment: func() string {
				parent := OperationDefinition{
					OperationId: "ValidatePets",
					Method:      "POST",
					Path:        "/pets:validate",
					Summary:     "Validate pets",
					Spec:        &openapi3.Operation{Description: "Validate pets"},
				}
				body := RequestBodyDefinition{ContentType: "application/json"}
				return body.GenerateFunctionComment("ValidatePets", parent, "WithJSONBody", false)
			}(),
		},
		{
			name: "OperationDefinition: description differs from summary but still looks like a heading",
			comment: func() string {
				op := OperationDefinition{
					OperationId: "ValidatePets",
					Method:      "POST",
					Path:        "/pets:validate",
					Summary:     "Validate pets",
					Spec:        &openapi3.Operation{Description: "Validates all pets in the store"},
				}
				return op.GenerateFunctionComment("ValidatePets", "", false)
			}(),
		},
		{
			name: "RequestBodyDefinition: description differs from summary but still looks like a heading",
			comment: func() string {
				parent := OperationDefinition{
					OperationId: "ValidatePets",
					Method:      "POST",
					Path:        "/pets:validate",
					Summary:     "Validate pets",
					Spec:        &openapi3.Operation{Description: "Validates all pets in the store"},
				}
				body := RequestBodyDefinition{ContentType: "application/json"}
				return body.GenerateFunctionComment("ValidatePets", parent, "WithJSONBody", false)
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap the comment in a minimal valid Go source file so that
			// go/format treats it as a doc comment on a top-level declaration,
			// which is where gofmt's old-style heading promotion fires.
			src := "package p\n\n" + tt.comment + "\nfunc F() {}\n"
			formatted, err := format.Source([]byte(src))
			require.NoError(t, err)
			assert.NotContains(t, string(formatted), "// #",
				"gofmt must not promote any description paragraph to a heading")
		})
	}
}

func TestNewInitiatorTemplateDataRejectsPathParams(t *testing.T) {
	_, err := NewInitiatorTemplateData("Webhook", []OperationDefinition{
		{
			OperationId: "PetUpdated",
			PathParams:  []ParameterDefinition{{ParamName: "petId"}},
		},
	})
	require.ErrorContains(t, err, "path parameters")

	data, err := NewInitiatorTemplateData("Callback", []OperationDefinition{
		{OperationId: "PetUpdated"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Callback", data.Prefix)
	assert.Equal(t, "callback", data.PrefixLower)
}

// buildStrictTestTrees loads the real embedded templates and per-framework
// clones the same way Generate() does, for exercising GenerateStrictServer.
func buildStrictTestTrees(t *testing.T) (*template.Template, map[string]*template.Template) {
	t.Helper()
	funcs := make(template.FuncMap, len(TemplateFunctions)+1)
	for k, v := range TemplateFunctions {
		funcs[k] = v
	}
	// Generate() injects "opts" before loading templates; stub it here.
	funcs["opts"] = func() Configuration { return Configuration{} }
	base := template.New("codegen").Funcs(funcs)
	require.NoError(t, LoadTemplates(templates, base))
	clones, err := buildServerTemplates(templates, base)
	require.NoError(t, err)
	return base, clones
}

func TestGenerateStrictServerInterfaceDedup(t *testing.T) {
	base, clones := buildStrictTestTrees(t)
	ops := []OperationDefinition{{OperationId: "Ping"}}

	// Frameworks whose shared interface template renders identically may be
	// combined; the interface must be emitted exactly once.
	out, err := GenerateStrictServer(base, clones, ops, Configuration{
		Generate: GenerateOptions{ChiServer: true, GinServer: true},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(out, "type StrictServerInterface interface"))

	// Fiber v2 and v3 render the shared interface template with different
	// context types; combining them can never compile and must error.
	_, err = GenerateStrictServer(base, clones, ops, Configuration{
		Generate: GenerateOptions{FiberServer: true, FiberV3Server: true},
	})
	require.ErrorContains(t, err, "cannot be generated together")
}
