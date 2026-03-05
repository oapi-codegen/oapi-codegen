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
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{{
		str:  "",
		want: "",
	}, {
		str:  " foo_bar ",
		want: "FooBar",
	}, {
		str:  "hi hello-hey-hallo",
		want: "HiHelloHeyHallo",
	}, {
		str:  "foo#bar",
		want: "FooBar",
	}, {
		str:  "foo2bar",
		want: "Foo2bar",
	}, {
		// Test that each substitution works
		str:  "word.word-WORD+Word_word~word(Word)Word{Word}Word[Word]Word:Word;",
		want: "WordWordWORDWordWordWordWordWordWordWordWordWordWord",
	}, {
		// Make sure numbers don't interact in a funny way.
		str:  "number-1234",
		want: "Number1234",
	},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.str, func(t *testing.T) {
			require.Equal(t, tt.want, ToCamelCase(tt.str))
		})
	}
}

func TestToCamelCaseWithDigits(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{{
		str:  "",
		want: "",
	}, {
		str:  " foo_bar ",
		want: "FooBar",
	}, {
		str:  "hi hello-hey-hallo",
		want: "HiHelloHeyHallo",
	}, {
		str:  "foo#bar",
		want: "FooBar",
	}, {
		str:  "foo2bar",
		want: "Foo2Bar",
	}, {
		str:  "пир2пир",
		want: "Пир2Пир",
	}, {
		// Test that each substitution works
		str:  "word.word3word-WORD+Word_word~word(Word)Word{Word}Word[Word]Word:Word;",
		want: "WordWord3WordWORDWordWordWordWordWordWordWordWordWordWord",
	}, {
		// Make sure numbers don't interact in a funny way.
		str:  "number-1234",
		want: "Number1234",
	},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.str, func(t *testing.T) {
			require.Equal(t, tt.want, ToCamelCaseWithDigits(tt.str))
		})
	}
}

func TestToCamelCaseWithInitialisms(t *testing.T) {
	tests := []struct {
		str  string
		want string
	}{{
		str:  "",
		want: "",
	}, {
		str:  "hello",
		want: "Hello",
	}, {
		str:  "DBError",
		want: "DBError",
	}, {
		str:  "httpOperationId",
		want: "HTTPOperationID",
	}, {
		str:  "OperationId",
		want: "OperationID",
	}, {
		str:  "peer2peer",
		want: "Peer2Peer",
	}, {
		str:  "makeUtf8",
		want: "MakeUTF8",
	}, {
		str:  "utf8Hello",
		want: "UTF8Hello",
	}, {
		str:  "myDBError",
		want: "MyDBError",
	}, {
		str:  " DbLayer ",
		want: "DBLayer",
	}, {
		str:  "FindPetById",
		want: "FindPetByID",
	}, {
		str:  "MyHttpUrl",
		want: "MyHTTPURL",
	}, {
		str:  "find_user_by_uuid",
		want: "FindUserByUUID",
	}, {
		str:  "HelloПриветWorldМир42",
		want: "HelloПриветWorldМир42",
	}, {
		str:  "пир2пир",
		want: "Пир2Пир",
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.str, func(t *testing.T) {
			require.Equal(t, tt.want, ToCamelCaseWithInitialisms(tt.str))
		})
	}
}

func TestSortedSchemaKeysWithXOrder(t *testing.T) {
	withOrder := func(i float64) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Extensions: map[string]any{"x-order": i},
			},
		}
	}
	dict := map[string]*openapi3.SchemaRef{
		"first":            withOrder(1),
		"minusTenth":       withOrder(-10),
		"zero":             withOrder(0),
		"minusHundredth_2": withOrder(-100),
		"minusHundredth_1": withOrder(-100),
		"afterFirst":       withOrder(2),
		"last":             withOrder(100),
		"middleA":          nil,
		"middleB":          nil,
		"middleC":          nil,
	}

	expected := []string{"minusHundredth_1", "minusHundredth_2", "minusTenth", "zero", "first", "afterFirst", "middleA", "middleB", "middleC", "last"}

	assert.EqualValues(t, expected, SortedSchemaKeys(dict), "Keys are not sorted properly")
}

func TestSortedSchemaKeysWithXOrderFromParsed(t *testing.T) {
	rawSpec := `---
components:
  schemas:
    AlwaysLast:
      type: string
      x-order: 100000
    DateInterval:
      type: object
      required:
        - name
      properties:
        end:
          type: string
          format: date
          x-order: 2
        start:
          type: string
          format: date
          x-order: 1
  `

	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData([]byte(rawSpec))
	require.NoError(t, err)
	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)

	t.Run("for the top-level schemas", func(t *testing.T) {
		expected := []string{"DateInterval", "AlwaysLast"}

		actual := SortedSchemaKeys(spec.Components.Schemas)

		assert.EqualValues(t, expected, actual)
	})

	t.Run("for DateInterval's keys", func(t *testing.T) {
		schemas, found := spec.Components.Schemas["DateInterval"]
		require.True(t, found, "did not find `#/components/schemas/DateInterval`")

		expected := []string{"start", "end"}

		actual := SortedSchemaKeys(schemas.Value.Properties)

		assert.EqualValues(t, expected, actual, "Keys are not sorted properly")
	})

}

func TestRefPathToGoType(t *testing.T) {
	old := globalState.importMapping
	globalState.importMapping = constructImportMapping(
		map[string]string{
			"doc.json":                    "externalref0",
			"http://deepmap.com/doc.json": "externalref1",
			// using the "current package" mapping
			"dj-current-package.yml": "-",
		},
	)
	defer func() { globalState.importMapping = old }()

	tests := []struct {
		name   string
		path   string
		goType string
	}{
		{
			name:   "local-schemas",
			path:   "#/components/schemas/Foo",
			goType: "Foo",
		},
		{
			name:   "local-parameters",
			path:   "#/components/parameters/foo_bar",
			goType: "FooBar",
		},
		{
			name:   "local-responses",
			path:   "#/components/responses/wibble",
			goType: "Wibble",
		},
		{
			name:   "local-mapped-current-package",
			path:   "dj-current-package.yml#/components/schemas/Foo",
			goType: "Foo",
		},
		{
			name:   "remote-root",
			path:   "doc.json#/foo",
			goType: "externalRef0.Foo",
		},
		{
			name:   "remote-pathed",
			path:   "doc.json#/components/parameters/foo",
			goType: "externalRef0.Foo",
		},
		{
			name:   "url-root",
			path:   "http://deepmap.com/doc.json#/foo_bar",
			goType: "externalRef1.FooBar",
		},
		{
			name:   "url-pathed",
			path:   "http://deepmap.com/doc.json#/components/parameters/foo_bar",
			goType: "externalRef1.FooBar",
		},
		{
			name: "local-too-deep",
			path: "#/components/parameters/foo/components/bar",
		},
		{
			name: "remote-too-deep",
			path: "doc.json#/components/parameters/foo/foo_bar",
		},
		{
			name: "url-too-deep",
			path: "http://deepmap.com/doc.json#/components/parameters/foo/foo_bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goType, err := RefPathToGoType(tc.path)
			if tc.goType == "" {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.goType, goType)
		})
	}
}

func TestIsWholeDocumentReference(t *testing.T) {
	assert.Equal(t, false, IsWholeDocumentReference(""))
	assert.Equal(t, false, IsWholeDocumentReference("#/components/schemas/Foo"))
	assert.Equal(t, false, IsWholeDocumentReference("doc.json#/components/schemas/Foo"))
	assert.Equal(t, true, IsWholeDocumentReference("doc.json"))
	assert.Equal(t, true, IsWholeDocumentReference("../doc.json"))
	assert.Equal(t, false, IsWholeDocumentReference("http://deepmap.com/doc.json#/components/parameters/foo_bar"))
	assert.Equal(t, true, IsWholeDocumentReference("http://deepmap.com/doc.json"))
}

func TestIsGoTypeReference(t *testing.T) {
	assert.Equal(t, false, IsGoTypeReference(""))
	assert.Equal(t, true, IsGoTypeReference("#/components/schemas/Foo"))
	assert.Equal(t, true, IsGoTypeReference("doc.json#/components/schemas/Foo"))
	assert.Equal(t, false, IsGoTypeReference("doc.json"))
	assert.Equal(t, false, IsGoTypeReference("../doc.json"))
	assert.Equal(t, true, IsGoTypeReference("http://deepmap.com/doc.json#/components/parameters/foo_bar"))
	assert.Equal(t, false, IsGoTypeReference("http://deepmap.com/doc.json"))
}

func TestSwaggerUriToIrisUri(t *testing.T) {
	assert.Equal(t, "/path", SwaggerUriToIrisUri("/path"))
	assert.Equal(t, "/path/:arg", SwaggerUriToIrisUri("/path/{arg}"))
	assert.Equal(t, "/path/:arg1/:arg2", SwaggerUriToIrisUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/:arg1/:arg2/foo", SwaggerUriToIrisUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToIrisUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToEchoUri(t *testing.T) {
	assert.Equal(t, "/path", SwaggerUriToEchoUri("/path"))
	assert.Equal(t, "/path/:arg", SwaggerUriToEchoUri("/path/{arg}"))
	assert.Equal(t, "/path/:arg1/:arg2", SwaggerUriToEchoUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/:arg1/:arg2/foo", SwaggerUriToEchoUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToEchoUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToGinUri(t *testing.T) {
	assert.Equal(t, "/path", SwaggerUriToGinUri("/path"))
	assert.Equal(t, "/path/:arg", SwaggerUriToGinUri("/path/{arg}"))
	assert.Equal(t, "/path/:arg1/:arg2", SwaggerUriToGinUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/:arg1/:arg2/foo", SwaggerUriToGinUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToGinUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToGorillaUri(t *testing.T) { // TODO
	assert.Equal(t, "/path", SwaggerUriToGorillaUri("/path"))
	assert.Equal(t, "/path/{arg}", SwaggerUriToGorillaUri("/path/{arg}"))
	assert.Equal(t, "/path/{arg1}/{arg2}", SwaggerUriToGorillaUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/{arg1}/{arg2}/foo", SwaggerUriToGorillaUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToGorillaUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToFiberUri(t *testing.T) {
	assert.Equal(t, "/path", SwaggerUriToFiberUri("/path"))
	assert.Equal(t, "/path/:arg", SwaggerUriToFiberUri("/path/{arg}"))
	assert.Equal(t, "/path/:arg1/:arg2", SwaggerUriToFiberUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/:arg1/:arg2/foo", SwaggerUriToFiberUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/:arg/foo", SwaggerUriToFiberUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToChiUri(t *testing.T) {
	assert.Equal(t, "/path", SwaggerUriToChiUri("/path"))
	assert.Equal(t, "/path/{arg}", SwaggerUriToChiUri("/path/{arg}"))
	assert.Equal(t, "/path/{arg1}/{arg2}", SwaggerUriToChiUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/{arg1}/{arg2}/foo", SwaggerUriToChiUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToChiUri("/path/{?arg*}/foo"))
}

func TestSwaggerUriToStdHttpUriUri(t *testing.T) {
	assert.Equal(t, "/{$}", SwaggerUriToStdHttpUri("/"))
	assert.Equal(t, "/path", SwaggerUriToStdHttpUri("/path"))
	assert.Equal(t, "/path/{arg}", SwaggerUriToStdHttpUri("/path/{arg}"))
	assert.Equal(t, "/path/{arg1}/{arg2}", SwaggerUriToStdHttpUri("/path/{arg1}/{arg2}"))
	assert.Equal(t, "/path/{arg1}/{arg2}/foo", SwaggerUriToStdHttpUri("/path/{arg1}/{arg2}/foo"))

	// Make sure all the exploded and alternate formats match too
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{.arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{.arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{;arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{;arg*}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{?arg}/foo"))
	assert.Equal(t, "/path/{arg}/foo", SwaggerUriToStdHttpUri("/path/{?arg*}/foo"))
}

func TestOrderedParamsFromUri(t *testing.T) {
	result := OrderedParamsFromUri("/path/{param1}/{.param2}/{;param3*}/foo")
	assert.EqualValues(t, []string{"param1", "param2", "param3"}, result)

	result = OrderedParamsFromUri("/path/foo")
	assert.EqualValues(t, []string{}, result)

	// A parameter can appear more than once in the URI (e.g. Keycloak API).
	// OrderedParamsFromUri faithfully returns all occurrences.
	result = OrderedParamsFromUri("/admin/realms/{realm}/clients/{client-uuid}/roles/{role-name}/composites/clients/{client-uuid}")
	assert.EqualValues(t, []string{"realm", "client-uuid", "role-name", "client-uuid"}, result)
}

func TestSortParamsByPath(t *testing.T) {
	strSchema := &openapi3.Schema{Type: &openapi3.Types{"string"}}

	t.Run("reorders params to match path order", func(t *testing.T) {
		params := []ParameterDefinition{
			{ParamName: "b", In: "path", Spec: &openapi3.Parameter{Name: "b", Schema: &openapi3.SchemaRef{Value: strSchema}}},
			{ParamName: "a", In: "path", Spec: &openapi3.Parameter{Name: "a", Schema: &openapi3.SchemaRef{Value: strSchema}}},
		}
		sorted, err := SortParamsByPath("/foo/{a}/bar/{b}", params)
		require.NoError(t, err)
		require.Len(t, sorted, 2)
		assert.Equal(t, "a", sorted[0].ParamName)
		assert.Equal(t, "b", sorted[1].ParamName)
	})

	t.Run("errors on missing parameter", func(t *testing.T) {
		params := []ParameterDefinition{
			{ParamName: "a", In: "path", Spec: &openapi3.Parameter{Name: "a", Schema: &openapi3.SchemaRef{Value: strSchema}}},
		}
		_, err := SortParamsByPath("/foo/{a}/bar/{b}", params)
		assert.Error(t, err)
	})

	t.Run("handles duplicate path parameters", func(t *testing.T) {
		// This is the Keycloak-style path where {client-uuid} appears twice.
		// The spec only declares 3 unique parameters.
		params := []ParameterDefinition{
			{ParamName: "realm", In: "path", Spec: &openapi3.Parameter{Name: "realm", Schema: &openapi3.SchemaRef{Value: strSchema}}},
			{ParamName: "client-uuid", In: "path", Spec: &openapi3.Parameter{Name: "client-uuid", Schema: &openapi3.SchemaRef{Value: strSchema}}},
			{ParamName: "role-name", In: "path", Spec: &openapi3.Parameter{Name: "role-name", Schema: &openapi3.SchemaRef{Value: strSchema}}},
		}
		path := "/admin/realms/{realm}/clients/{client-uuid}/roles/{role-name}/composites/clients/{client-uuid}"
		sorted, err := SortParamsByPath(path, params)
		require.NoError(t, err)
		// Should return 3 unique params in first-occurrence order
		require.Len(t, sorted, 3)
		assert.Equal(t, "realm", sorted[0].ParamName)
		assert.Equal(t, "client-uuid", sorted[1].ParamName)
		assert.Equal(t, "role-name", sorted[2].ParamName)
	})
}

func TestReplacePathParamsWithStr(t *testing.T) {
	result := ReplacePathParamsWithStr("/path/{param1}/{.param2}/{;param3*}/foo")
	assert.EqualValues(t, "/path/%s/%s/%s/foo", result)
}

func TestStringToGoStringValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		message  string
	}{
		{
			input:    ``,
			expected: `""`,
			message:  "blank string should be converted to empty Go string literal",
		},
		{
			input:    `application/json`,
			expected: `"application/json"`,
			message:  "typical string should be returned as-is",
		},
		{
			input:    `application/json; foo="bar"`,
			expected: `"application/json; foo=\"bar\""`,
			message:  "string with quotes should include escape characters",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			result := StringToGoString(testCase.input)
			assert.EqualValues(t, testCase.expected, result, testCase.message)
		})
	}
}

func TestStringToGoComment(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		message  string
	}{
		{
			input:    "",
			expected: "",
			message:  "blank string should be ignored due to human unreadable",
		},
		{
			input:    " ",
			expected: "",
			message:  "whitespace should be ignored due to human unreadable",
		},
		{
			input:    "Single Line",
			expected: "// Single Line",
			message:  "single line comment",
		},
		{
			input:    "    Single Line",
			expected: "//     Single Line",
			message:  "single line comment preserving whitespace",
		},
		{
			input: `Multi
Line
  With
    Spaces
	And
		Tabs
`,
			expected: `// Multi
// Line
//   With
//     Spaces
// 	And
// 		Tabs`,
			message: "multi line preserving whitespaces using tabs or spaces",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			result := StringToGoComment(testCase.input)
			assert.EqualValues(t, testCase.expected, result, testCase.message)
		})
	}
}

func TestStringWithTypeNameToGoComment(t *testing.T) {
	testCases := []struct {
		input     string
		inputName string
		expected  string
		message   string
	}{
		{
			input:     "",
			inputName: "",
			expected:  "",
			message:   "blank string should be ignored due to human unreadable",
		},
		{
			input:    " ",
			expected: "",
			message:  "whitespace should be ignored due to human unreadable",
		},
		{
			input:     "Single Line",
			inputName: "SingleLine",
			expected:  "// SingleLine Single Line",
			message:   "single line comment",
		},
		{
			input:     "    Single Line",
			inputName: "SingleLine",
			expected:  "// SingleLine     Single Line",
			message:   "single line comment preserving whitespace",
		},
		{
			input: `Multi
Line
  With
    Spaces
	And
		Tabs
`,
			inputName: "MultiLine",
			expected: `// MultiLine Multi
// Line
//   With
//     Spaces
// 	And
// 		Tabs`,
			message: "multi line preserving whitespaces using tabs or spaces",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.message, func(t *testing.T) {
			result := StringWithTypeNameToGoComment(testCase.input, testCase.inputName)
			assert.EqualValues(t, testCase.expected, result, testCase.message)
		})
	}
}

func TestEscapePathElements(t *testing.T) {
	p := "/foo/bar/baz"
	assert.Equal(t, p, EscapePathElements(p))

	p = "foo/bar/baz"
	assert.Equal(t, p, EscapePathElements(p))

	p = "/foo/bar:baz"
	assert.Equal(t, "/foo/bar%3Abaz", EscapePathElements(p))
}

func TestSchemaNameToTypeName(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"$":            "DollarSign",
		"$ref":         "Ref",
		"no_prefix~+-": "NoPrefix",
		"123":          "N123",
		"-1":           "Minus1",
		"+1":           "Plus1",
		"@timestamp,":  "Timestamp",
		"&now":         "AndNow",
		"~":            "Tilde",
		"_foo":         "UnderscoreFoo",
		"=3":           "Equal3",
		"#Tag":         "HashTag",
		".com":         "DotCom",
		"_1":           "Underscore1",
		">=":           "GreaterThanEqual",
		"<=":           "LessThanEqual",
		"<":            "LessThan",
		">":            "GreaterThan",
	} {
		assert.Equal(t, want, SchemaNameToTypeName(in))
	}
}

func TestTypeDefinitionsEquivalent(t *testing.T) {
	def1 := TypeDefinition{TypeName: "name", Schema: Schema{
		OAPISchema: &openapi3.Schema{},
	}}
	def2 := TypeDefinition{TypeName: "name", Schema: Schema{
		OAPISchema: &openapi3.Schema{},
	}}
	assert.True(t, TypeDefinitionsEquivalent(def1, def2))
}

func TestRefPathToObjName(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"#/components/schemas/Foo":                         "Foo",
		"#/components/parameters/Bar":                      "Bar",
		"#/components/responses/baz_baz":                   "baz_baz",
		"document.json#/Foo":                               "Foo",
		"http://deepmap.com/schemas/document.json#/objObj": "objObj",
	} {
		assert.Equal(t, want, RefPathToObjName(in))
	}
}

func TestLowercaseFirstCharacters(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "id",
			expected: "id",
		},
		{
			name:     "CamelCase",
			expected: "camelCase",
		},
		{
			name:     "ID",
			expected: "id",
		},
		{
			name:     "DBTree",
			expected: "dbTree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LowercaseFirstCharacters(tt.name))
		})
	}
}

func Test_replaceInitialism(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{s: ""},
			want: "",
		},
		{
			name: "no initialism",
			args: args{s: "foo"},
			want: "foo",
		},
		{
			name: "one initialism",
			args: args{s: "fooId"},
			want: "fooID",
		},
		{
			name: "two initialism",
			args: args{s: "fooIdBarApi"},
			want: "fooIDBarAPI",
		},
		{
			name: "already initialism",
			args: args{s: "fooIDBarAPI"},
			want: "fooIDBarAPI",
		},
		{
			name: "one initialism at start",
			args: args{s: "idFoo"},
			want: "idFoo",
		},
		{
			name: "one initialism at start and one in middle",
			args: args{s: "apiIdFoo"},
			want: "apiIDFoo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, replaceInitialism(tt.args.s), "replaceInitialism(%v)", tt.args.s)
		})
	}
}
