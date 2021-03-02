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
	goType, err := RefPathToGoType("#/components/schemas/Foo")
	assert.Equal(t, "Foo", goType)
	assert.NoError(t, err, "Expecting no error")

	goType, err = RefPathToGoType("#/components/parameters/foo_bar")
	assert.Equal(t, "FooBar", goType)
	assert.NoError(t, err, "Expecting no error")

	_, err = RefPathToGoType("http://deepmap.com/doc.json#/components/parameters/foo_bar")
	assert.Errorf(t, err, "Expected an error on URL reference")

	_, err = RefPathToGoType("doc.json#/components/parameters/foo_bar")
	assert.Errorf(t, err, "Expected an error on remote reference")

	_, err = RefPathToGoType("#/components/parameters/foo/components/bar")
	assert.Errorf(t, err, "Expected an error on reference depth")
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
			expected: "// ",
			message:  "blank string should be preserved with comment",
		},
		{
			input:    " ",
			expected: "//  ",
			message:  "whitespace should be preserved with comment",
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
