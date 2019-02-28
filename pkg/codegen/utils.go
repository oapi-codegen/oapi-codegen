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
    "errors"
    "github.com/getkin/kin-openapi/openapi3"
    "regexp"
    "sort"
    "strings"
    "unicode"
)

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
// use (., -, _, ~, ' ') as valid delimiters for words. So, "word.word-word~word_word word"
// would be converted to WordWordWordWord
func ToCamelCase(str string) string {
    separators := []string{".", "-", "_", "~", " "}
    in :=[]string{str}
    out := make([]string, 0)

    for _, sep := range separators {
        for _, inStr := range in {
            parts := strings.Split(inStr, sep)
            out = append(out, parts...)
        }
        in = out
        out = make([]string, 0)
    }

    words := in

    for i := range words {
        words[i] = UppercaseFirstCharacter(words[i])
    }
    return strings.Join(words, "")
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
// Remote references (document.json#/Foo) are not yet supported
// URL references (http://deepmap.com/schemas/document.json#Foo) are not yet
// supported
// We only support flat references for now, so no components in a schema under
// components.

func RefPathToGoType(refPath string) (string, error) {
    pathParts := strings.Split(refPath, "/")
    if pathParts[0] != "#" {
        return "", errors.New("Only local document references are supported")
    }
    if len(pathParts) != 4 {
        return "", errors.New("Parameter nesting is deeper than supported")
    }
    return ToCamelCase(pathParts[3]), nil
}

// This function converts a swagger style path URI with parameters to a
// Echo compatible path URI. We need to replace all instances of {param} with
// :param
func SwaggerUriToEchoUri(uri string) string {
    exp, err := regexp.Compile("{([^{}]+)}")
    // We don't use a dynamic regexp, so this should always succeed.
    if err != nil {
        panic(err)
    }
    return exp.ReplaceAllString(uri, ":$1")
}
