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
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"

	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

var pathParamRE *regexp.Regexp

// list of valid word characters in an identifier (excluding '/') used by ToCamelCase
var separatorRE *regexp.Regexp

func init() {
	pathParamRE = regexp.MustCompile("{[.;?]?([^{}*]+)\\*?}")
	separatorRE = regexp.MustCompile(`[^.+:;_~ (){}\[\]\-]+`)
}

// Uppercase the first character in a string. This assumes UTF-8, so we have
// to be careful with unicode, don't treat it as a byte array.
func UppercaseFirstCharacter(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Same as above, except lower case
func LowercaseFirstCharacter(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// This function will convert query-arg style strings to CamelCase. We will
// use `., -, +, :, ;, _, ~, ' ', (, ), {, }, [, ]` as valid delimiters for words.
// So, "word.word-word+word:word;word_word~word word(word)word{word}[word]"
// would be converted to WordWordWordWordWordWordWordWordWordWordWordWordWord
func ToCamelCase(str string) string {
	separators := "-#@!$&=.+:;_~ (){}[]"
	s := strings.Trim(str, " ")

	n := ""
	capNext := true
	for _, v := range s {
		if unicode.IsUpper(v) {
			n += string(v)
		}
		if unicode.IsDigit(v) {
			n += string(v)
		}
		if unicode.IsLower(v) {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}

		if strings.ContainsRune(separators, v) {
			capNext = true
		} else {
			capNext = false
		}
	}
	return n
}

// This function returns the keys of the given SchemaRef dictionary in sorted
// order, since Golang scrambles dictionary keys
func SortedSchemaKeys(dict map[string]*openapi3.SchemaRef) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This function is the same as above, except it sorts the keys for a Paths
// dictionary.
func SortedPathsKeys(dict openapi3.Paths) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This function returns Operation dictionary keys in sorted order
func SortedOperationsKeys(dict map[string]*openapi3.Operation) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This function returns Responses dictionary keys in sorted order
func SortedResponsesKeys(dict openapi3.Responses) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This returns Content dictionary keys in sorted order
func SortedContentKeys(dict openapi3.Content) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This returns string map keys in sorted order
func SortedStringKeys(dict map[string]string) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This returns sorted keys for a ParameterRef dict
func SortedParameterKeys(dict map[string]*openapi3.ParameterRef) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

func SortedRequestBodyKeys(dict map[string]*openapi3.RequestBodyRef) []string {
	keys := make([]string, len(dict))
	i := 0
	for key := range dict {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// This function checks whether the specified string is present in an array
// of strings
func StringInArray(str string, array []string) bool {
	for _, elt := range array {
		if elt == str {
			return true
		}
	}
	return false
}

// This function takes a $ref value and converts it to a Go typename.
// #/components/schemas/Foo -> Foo
// #/components/parameters/Bar -> Bar
// #/components/responses/Baz -> Baz
//
// To specify remote paths there must also be a typeImport map with the types
// mapped to import paths.
// https://foo.com/bar#/components/parameters/Bar -> packagename.Bar
// foo.json#/components/parameters/Bar -> packagename.Bar
func RefPathToGoType(refPath string, importedTypes map[string]TypeImportSpec) (string, error) {
	s := strings.Split(refPath, "#")
	if len(s) < 2 {
		return "", errors.New("Missing fragment marker - this does not seem like a local or remote reference")
	} else if len(s) > 2 {
		return "", errors.New("Extra fragment marker")
	}

	pathParts := strings.Split(s[len(s)-1], "/")

	if len(pathParts) != 4 {
		return "", errors.New("Parameter nesting is deeper than supported")
	}
	schemaName := SchemaNameToTypeName(pathParts[len(pathParts)-1])

	if s[0] != "" {
		if importedTypes == nil {
			return "", errors.New("detected remote document component but no TypeImports specified")
		}
		importedType, ok := importedTypes[schemaName]
		if !ok {
			return "", errors.New("detected remote document component not specified in TypeImports")
		}

		return importedType.PackageName + "." + schemaName, nil
	}
	return schemaName, nil
}

// This function converts a swagger style path URI with parameters to a
// Echo compatible path URI. We need to replace all of Swagger parameters with
// ":param". Valid input parameters are:
//   {param}
//   {param*}
//   {.param}
//   {.param*}
//   {;param}
//   {;param*}
//   {?param}
//   {?param*}
func SwaggerUriToEchoUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, ":$1")
}

// This function converts a swagger style path URI with parameters to a
// Chi compatible path URI. We need to replace all of Swagger parameters with
// "{param}". Valid input parameters are:
//   {param}
//   {param*}
//   {.param}
//   {.param*}
//   {;param}
//   {;param*}
//   {?param}
//   {?param*}
func SwaggerUriToChiUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "{$1}")
}

// Returns the argument names, in order, in a given URI string, so for
// /path/{param1}/{.param2*}/{?param3}, it would return param1, param2, param3
func OrderedParamsFromUri(uri string) []string {
	matches := pathParamRE.FindAllStringSubmatch(uri, -1)
	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m[1]
	}
	return result
}

// Replaces path parameters with %s
func ReplacePathParamsWithStr(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "%s")
}

// Replaces path parameters with Par1, Par2...
func ReplacePathParamsWithParNStr(uri string) string {
	p := 0
	return pathParamRE.ReplaceAllStringFunc(uri, func(_ string) string { p++; return fmt.Sprintf("Par%d", p) })
}

// Reorders the given parameter definitions to match those in the path URI.
func SortParamsByPath(path string, in []ParameterDefinition) ([]ParameterDefinition, error) {
	pathParams := OrderedParamsFromUri(path)
	n := len(in)
	if len(pathParams) != n {
		return nil, fmt.Errorf("path '%s' has %d positional parameters, but spec has %d declared",
			path, len(pathParams), n)
	}
	out := make([]ParameterDefinition, len(in))
	for i, name := range pathParams {
		p := ParameterDefinitions(in).FindByName(name)
		if p == nil {
			return nil, fmt.Errorf("path '%s' refers to parameter '%s', which doesn't exist in specification",
				path, name)
		}
		out[i] = *p
	}
	return out, nil
}

// Returns whether the given string is a go keyword
func IsGoKeyword(str string) bool {
	keywords := []string{
		"break",
		"case",
		"chan",
		"const",
		"continue",
		"default",
		"defer",
		"else",
		"fallthrough",
		"for",
		"func",
		"go",
		"goto",
		"if",
		"import",
		"interface",
		"map",
		"package",
		"range",
		"return",
		"select",
		"struct",
		"switch",
		"type",
		"var",
	}

	for _, k := range keywords {
		if k == str {
			return true
		}
	}
	return false
}

// Converts a Schema name to a valid Go type name. It converts to camel case, and makes sure the name is
// valid in Go
func SchemaNameToTypeName(name string) string {
	name = ToCamelCase(name)
	// Prepend "N" to schemas starting with a number
	if name != "" && unicode.IsDigit([]rune(name)[0]) {
		name = "N" + name
	}
	return name
}

// According to the spec, additionalProperties may be true, false, or a
// schema. If not present, true is implied. If it's a schema, true is implied.
// If it's false, no additional properties are allowed. We're going to act a little
// differently, in that if you want additionalProperties code to be generated,
// you must specify an additionalProperties type
// If additionalProperties it true/false, this field will be non-nil.
func SchemaHasAdditionalProperties(schema *openapi3.Schema) bool {
	if schema.AdditionalPropertiesAllowed != nil {
		return *schema.AdditionalPropertiesAllowed
	}
	if schema.AdditionalProperties != nil {
		return true
	}
	return false
}

// This converts a path, like Object/field1/nestedField into a go
// type name.
func PathToTypeName(path []string) string {
	for i, p := range path {
		path[i] = ToCamelCase(p)
	}
	return strings.Join(path, "_")
}

// StringToGoComment renders a possible multi-line string as a valid Go-Comment.
// Each line is prefixed as a comment.
func StringToGoComment(in string) string {
	// Normalize newlines from Windows/Mac to Linux
	in = strings.Replace(in, "\r\n", "\n", -1)
	in = strings.Replace(in, "\r", "\n", -1)

	// Add comment to each line
	var lines []string
	for _, line := range strings.Split(in, "\n") {
		lines = append(lines, fmt.Sprintf("// %s", line))
	}
	in = strings.Join(lines, "\n")

	// in case we have a multiline string which ends with \n, we would generate
	// empty-line-comments, like `// `. Therefore remove this line comment.
	in = strings.TrimSuffix(in, "\n// ")
	return in
}

// This converts a path, like Object/field1/nestedField into a dot separated single string
func DotSeparatedPath(path []string) string {
	return strings.Join(path, ".")
}

// DotSeparatedPath converts a dot separated path string, like Object.field1 into a go identifier
func DottedStringToTypeName(path string) string {
	return strings.ReplaceAll(path, ".", "_")
}
