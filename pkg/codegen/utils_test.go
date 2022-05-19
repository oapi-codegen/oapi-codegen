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
)

func TestStringOps(t *testing.T) {
	// Test that each substitution works
	assert.Equal(t, "WordWordWORDWordWordWordWordWordWordWordWordWordWord", ToCamelCase("word.word-WORD+Word_word~word(Word)Word{Word}Word[Word]Word:Word;"), "Camel case conversion failed")

	// Make sure numbers don't interact in a funny way.
	assert.Equal(t, "Number1234", ToCamelCase("number-1234"), "Number Camelcasing not working.")
}

func TestSortedSchemaKeys(t *testing.T) {
	dict := map[string]*openapi3.SchemaRef{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedSchemaKeys(dict), "Keys are not sorted properly")
}

func TestSortedPathsKeys(t *testing.T) {
	dict := openapi3.Paths{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedPathsKeys(dict), "Keys are not sorted properly")
}

func TestSortedOperationsKeys(t *testing.T) {
	dict := map[string]*openapi3.Operation{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedOperationsKeys(dict), "Keys are not sorted properly")
}

func TestSortedResponsesKeys(t *testing.T) {
	dict := openapi3.Responses{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedResponsesKeys(dict), "Keys are not sorted properly")
}

func TestSortedContentKeys(t *testing.T) {
	dict := openapi3.Content{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedContentKeys(dict), "Keys are not sorted properly")
}

func TestSortedParameterKeys(t *testing.T) {
	dict := map[string]*openapi3.ParameterRef{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedParameterKeys(dict), "Keys are not sorted properly")
}

func TestSortedRequestBodyKeys(t *testing.T) {
	dict := map[string]*openapi3.RequestBodyRef{
		"f": nil,
		"c": nil,
		"b": nil,
		"e": nil,
		"d": nil,
		"a": nil,
	}

	expected := []string{"a", "b", "c", "d", "e", "f"}

	assert.EqualValues(t, expected, SortedRequestBodyKeys(dict), "Keys are not sorted properly")
}

func TestRefPathToGoType(t *testing.T) {
	old := importMapping
	importMapping = constructImportMapping(map[string]string{
		"doc.json":                    "externalref0",
		"http://deepmap.com/doc.json": "externalref1",
	})
	defer func() { importMapping = old }()

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

func TestOrderedParamsFromUri(t *testing.T) {
	result := OrderedParamsFromUri("/path/{param1}/{.param2}/{;param3*}/foo")
	assert.EqualValues(t, []string{"param1", "param2", "param3"}, result)

	result = OrderedParamsFromUri("/path/foo")
	assert.EqualValues(t, []string{}, result)
}

func TestReplacePathParamsWithStr(t *testing.T) {
	result := ReplacePathParamsWithStr("/path/{param1}/{.param2}/{;param3*}/foo")
	assert.EqualValues(t, "/path/%s/%s/%s/foo", result)
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
		"_foo":         "Foo",
		"=3":           "Equal3",
		"#Tag":         "HashTag",
		".com":         "DotCom",
	} {
		assert.Equal(t, want, SchemaNameToTypeName(in))
	}
}
