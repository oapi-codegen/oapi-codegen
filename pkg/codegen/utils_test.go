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
	tests := []struct {
		in, out, msg string
	}{
		{in: "word.word-WORD_word~word", out: "WordWordWORDWordWord", msg: "Camel case conversion failed"},
		{in: "number-1234", out: "Number1234", msg: "Number Camelcasing not working."},
		{in: "path~to~res", out: "PathToRes", msg: "Failed converting path separator"},
		{in: "path-{pathName}", out: "PathPathName", msg: "Failed converting path parameter"},
		{in: "0-a.b+c:d;e_f~g(h)i{j}k[l]m", out: "0ABCDEFGHIJKLM", msg: "Separator test failed"},
		{in: "0-A.B+C:D;E_F~G(H)I{J}K[L]M", out: "0ABCDEFGHIJKLM", msg: "Separator test failed"},
	}

	for _, test := range tests {
		assert.Equal(t, test.out, ToCamelCase(test.in), test.msg)
	}
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
	testTable := map[string]struct {
		path          string
		importedTypes map[string]TypeImportSpec
		goType        string
		isErr         bool
	}{
		"Pascal case": {
			"#/components/schemas/Foo",
			nil,
			"Foo",
			false,
		},
		"Snake case": {
			"#/components/parameters/foo_bar",
			nil,
			"FooBar",
			false,
		},
		"remote ref, no typemap": {
			"http://deepmap.com/doc.json#/components/parameters/foo_bar",
			nil,
			"",
			true,
		},
		"path ref, no typemap": {
			"doc.json#/components/parameters/foo_bar",
			nil,
			"",
			true,
		},
		"remote ref, matches typemap": {
			"http://deepmap.com/doc.json#/components/parameters/foo_bar",
			map[string]TypeImportSpec{"FooBar": TypeImportSpec{Name: "FooBar", PackageName: "mypkg", ImportPath: "github.com/me/mypkg"}},
			"mypkg.FooBar",
			false,
		},
		"remote ref, matches typemap, using PackageName": {
			"http://deepmap.com/doc.json#/components/parameters/foo_bar",
			map[string]TypeImportSpec{"FooBar": TypeImportSpec{Name: "FooBar", PackageName: "that_pkg", ImportPath: "github.com/me/mypkg"}},
			"that_pkg.FooBar",
			false,
		},
		"path ref, matches typemap": {
			"doc.json#/components/parameters/foo_bar",
			map[string]TypeImportSpec{"FooBar": TypeImportSpec{Name: "FooBar", PackageName: "mypkg", ImportPath: "github.com/me/mypkg"}},
			"mypkg.FooBar",
			false,
		},
		"remote ref, no match in typemap": {
			"http://deepmap.com/doc.json#/components/parameters/foo_bar",
			map[string]TypeImportSpec{"Foo": TypeImportSpec{Name: "Foo", PackageName: "mypkg", ImportPath: "github.com/me/mypkg"}},
			"",
			true,
		},
		"path ref, no match in typemap": {
			"doc.json#/components/parameters/foo_bar",
			map[string]TypeImportSpec{"Foo": TypeImportSpec{Name: "Foo", PackageName: "mypkg", ImportPath: "github.com/me/mypkg"}},
			"",
			true,
		},
		"reference depth incorrect": {
			"#/components/parameters/foo/components/bar",
			nil,
			"",
			true,
		},
		"invalid path, too many #": {
			"#/components/parameters/foo#",
			nil,
			"",
			true,
		},
		"invalid path, no #": {
			"/components/parameters/foo",
			nil,
			"",
			true,
		},
	}
	for name, test := range testTable {
		t.Run(name, func(t *testing.T) {
			goType, err := RefPathToGoType(test.path, test.importedTypes)
			assert.Equal(t, test.goType, goType)
			if test.isErr {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expecting no error")
			}
		})
	}

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

func TestReplacePathParamsWithParNStr(t *testing.T) {
	result := ReplacePathParamsWithParNStr("GET /path/{param1}/{.param2}/{;param3*}/foo")
	assert.EqualValues(t, "GET /path/Par1/Par2/Par3/foo", result)
}
