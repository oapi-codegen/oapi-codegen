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
	"bytes"
	"fmt"
	"go/token"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	pathParamRE    *regexp.Regexp
	predeclaredSet map[string]struct{}
	separatorSet   map[rune]struct{}
	nameNormalizer NameNormalizer = ToCamelCase
)

type NameNormalizerFunction string

const (
	// NameNormalizerFunctionUnset is the default case, where the `name-normalizer` option hasn't been set. This will use the `ToCamelCase` function.
	//
	// See the docs for `NameNormalizerFunctionToCamelCase` for more details.
	NameNormalizerFunctionUnset NameNormalizerFunction = ""
	// NameNormalizerFunctionToCamelCase will use the `ToCamelCase` function.
	//
	// For instance:
	//
	// - `getHttpPet`   => `GetHttpPet`
	// - `OneOf2things` => `OneOf2things`
	NameNormalizerFunctionToCamelCase NameNormalizerFunction = "ToCamelCase"
	// NameNormalizerFunctionToCamelCaseWithDigits will use the `NameNormalizerFunctionToCamelCaseWithDigits` function.
	//
	// For instance:
	//
	// - `getHttpPet`   => `GetHttpPet`
	// - `OneOf2things` => `OneOf2Things`
	NameNormalizerFunctionToCamelCaseWithDigits NameNormalizerFunction = "ToCamelCaseWithDigits"
	// NameNormalizerFunctionToCamelCaseWithInitialisms will use the `NameNormalizerFunctionToCamelCaseWithInitialisms` function.
	//
	// For instance:
	//
	// - `getHttpPet`   => `GetHTTPPet`
	// - `OneOf2things` => `OneOf2things`
	NameNormalizerFunctionToCamelCaseWithInitialisms NameNormalizerFunction = "ToCamelCaseWithInitialisms"
)

// NameNormalizer is a function that takes a type name, and returns that type name converted into a different format.
//
// This may be an Operation ID i.e. `retrieveUserRequests` or a Schema name i.e. `BigBlockOfCheese`
//
// NOTE: this must return a string that can be used as a valid Go type name
type NameNormalizer func(string) string

type NameNormalizerMap map[NameNormalizerFunction]NameNormalizer

func (m NameNormalizerMap) Options() []string {
	options := make([]string, 0, len(m))

	for key := range NameNormalizers {
		options = append(options, string(key))
	}

	sort.Strings(options)

	return options
}

// NameNormalizers contains the valid options for `NameNormalizerFunction`s that `oapi-codegen` supports.
//
// If you are calling `oapi-codegen` as a library, this allows you to specify your own normalisation types before generating code.
var NameNormalizers = NameNormalizerMap{
	NameNormalizerFunctionUnset:                      ToCamelCase,
	NameNormalizerFunctionToCamelCase:                ToCamelCase,
	NameNormalizerFunctionToCamelCaseWithDigits:      ToCamelCaseWithDigits,
	NameNormalizerFunctionToCamelCaseWithInitialisms: ToCamelCaseWithInitialisms,
}

func init() {
	pathParamRE = regexp.MustCompile(`{[.;?]?([^{}*]+)\*?}`)

	predeclaredIdentifiers := []string{
		// Types
		"bool",
		"byte",
		"complex64",
		"complex128",
		"error",
		"float32",
		"float64",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"rune",
		"string",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"uintptr",
		// Constants
		"true",
		"false",
		"iota",
		// Zero value
		"nil",
		// Functions
		"append",
		"cap",
		"close",
		"complex",
		"copy",
		"delete",
		"imag",
		"len",
		"make",
		"new",
		"panic",
		"print",
		"println",
		"real",
		"recover",
	}
	predeclaredSet = map[string]struct{}{}
	for _, id := range predeclaredIdentifiers {
		predeclaredSet[id] = struct{}{}
	}

	separators := "-#@!$&=.+:;_~ (){}[]"
	separatorSet = map[rune]struct{}{}
	for _, r := range separators {
		separatorSet[r] = struct{}{}
	}
}

// UppercaseFirstCharacter Uppercases the first character in a string. This assumes UTF-8, so we have
// to be careful with unicode, don't treat it as a byte array.
func UppercaseFirstCharacter(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Uppercase the first character in a identifier with pkg name. This assumes UTF-8, so we have
// to be careful with unicode, don't treat it as a byte array.
func UppercaseFirstCharacterWithPkgName(str string) string {
	if str == "" {
		return ""
	}

	segs := strings.Split(str, ".")
	var prefix string
	if len(segs) == 2 {
		prefix = segs[0] + "."
		str = segs[1]
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return prefix + string(runes)
}

// LowercaseFirstCharacter Lowercases the first character in a string. This assumes UTF-8, so we have
// to be careful with unicode, don't treat it as a byte array.
func LowercaseFirstCharacter(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Lowercase the first upper characters in a string for case of abbreviation.
// This assumes UTF-8, so we have to be careful with unicode, don't treat it as a byte array.
func LowercaseFirstCharacters(str string) string {
	if str == "" {
		return ""
	}

	runes := []rune(str)

	for i := 0; i < len(runes); i++ {
		next := i + 1
		if i != 0 && next < len(runes) && unicode.IsLower(runes[next]) {
			break
		}

		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

// ToCamelCase will convert query-arg style strings to CamelCase. We will
// use `., -, +, :, ;, _, ~, ' ', (, ), {, }, [, ]` as valid delimiters for words.
// So, "word.word-word+word:word;word_word~word word(word)word{word}[word]"
// would be converted to WordWordWordWordWordWordWordWordWordWordWordWordWord
func ToCamelCase(str string) string {
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
		_, capNext = separatorSet[v]
	}
	return n
}

// ToCamelCaseWithDigits function will convert query-arg style strings to CamelCase. We will
// use `., -, +, :, ;, _, ~, ' ', (, ), {, }, [, ]` as valid delimiters for words.
// The difference of ToCamelCase that letter after a number becomes capitalized.
// So, "word.word-word+word:word;word_word~word word(word)word{word}[word]3word"
// would be converted to WordWordWordWordWordWordWordWordWordWordWordWordWord3Word
func ToCamelCaseWithDigits(s string) string {
	res := bytes.NewBuffer(nil)
	capNext := true
	for _, v := range s {
		if unicode.IsUpper(v) {
			res.WriteRune(v)
			capNext = false
			continue
		}
		if unicode.IsDigit(v) {
			res.WriteRune(v)
			capNext = true
			continue
		}
		if unicode.IsLower(v) {
			if capNext {
				res.WriteRune(unicode.ToUpper(v))
			} else {
				res.WriteRune(v)
			}
			capNext = false
			continue
		}
		capNext = true
	}
	return res.String()
}

// ToCamelCaseWithInitialisms function will convert query-arg style strings to CamelCase with initialisms in uppercase.
// So, httpOperationId would be converted to HTTPOperationID
func ToCamelCaseWithInitialisms(s string) string {
	parts := camelCaseMatchParts.FindAllString(ToCamelCaseWithDigits(s), -1)
	for i := range parts {
		if v, ok := initialismsMap[strings.ToLower(parts[i])]; ok {
			parts[i] = v
		}
	}
	return strings.Join(parts, "")
}

var camelCaseMatchParts = regexp.MustCompile(`[\p{Lu}\d]+([\p{Ll}\d]+|$)`)

// initialismsMap stores initialisms as "lower(initialism) -> initialism" map.
// List of initialisms was taken from https://staticcheck.io/docs/configuration/options/#initialisms.
var initialismsMap = makeInitialismsMap([]string{
	"ACL", "API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON",
	"QPS", "RAM", "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP", "TLS", "TTL", "UDP", "UI", "GID", "UID", "UUID",
	"URI", "URL", "UTF8", "VM", "XML", "XMPP", "XSRF", "XSS", "SIP", "RTP", "AMQP", "DB", "TS",
})

func makeInitialismsMap(l []string) map[string]string {
	m := make(map[string]string, len(l))
	for i := range l {
		m[strings.ToLower(l[i])] = l[i]
	}
	return m
}

func ToCamelCaseWithInitialism(str string) string {
	return replaceInitialism(ToCamelCase(str))
}

func replaceInitialism(s string) string {
	// These strings do not apply CamelCase
	// Do not do CamelCase when these characters match when the preceding character is lowercase
	// ["Acl", "Api", "Ascii", "Cpu", "Css", "Dns", "Eof", "Guid", "Html", "Http", "Https", "Id", "Ip", "Json", "Qps", "Ram", "Rpc", "Sla", "Smtp", "Sql", "Ssh", "Tcp", "Tls", "Ttl", "Udp", "Ui", "Gid", "Uid", "Uuid", "Uri", "Url", "Utf8", "Vm", "Xml", "Xmpp", "Xsrf", "Xss", "Sip", "Rtp", "Amqp", "Db", "Ts"]
	targetWordRegex := regexp.MustCompile(`(?i)(Acl|Api|Ascii|Cpu|Css|Dns|Eof|Guid|Html|Http|Https|Id|Ip|Json|Qps|Ram|Rpc|Sla|Smtp|Sql|Ssh|Tcp|Tls|Ttl|Udp|Ui|Gid|Uid|Uuid|Uri|Url|Utf8|Vm|Xml|Xmpp|Xsrf|Xss|Sip|Rtp|Amqp|Db|Ts)`)
	return targetWordRegex.ReplaceAllStringFunc(s, func(s string) string {
		// If the preceding character is lowercase, do not do CamelCase
		if unicode.IsLower(rune(s[0])) {
			return s
		}
		return strings.ToUpper(s)
	})
}

// mediaTypeToCamelCase converts a media type to a PascalCase representation
func mediaTypeToCamelCase(s string) string {
	// ToCamelCase doesn't - and won't - add `/` to the characters it'll allow word boundary
	s = strings.Replace(s, "/", "_", 1)
	// including a _ to make sure that these are treated as word boundaries by `ToCamelCase`
	s = strings.Replace(s, "*", "Wildcard_", 1)
	s = strings.Replace(s, "+", "Plus_", 1)

	return ToCamelCaseWithInitialism(s)
}

// SortedMapKeys takes a map with keys of type string and returns a slice of those
// keys sorted lexicographically.
func SortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SortedSchemaKeys returns the keys of the given SchemaRef dictionary in sorted
// order, since Golang scrambles dictionary keys. This isn't a generic key sort, because
// we support an extension to grant specific orders to schemas to help control output
// ordering.
func SortedSchemaKeys(dict map[string]*openapi3.SchemaRef) []string {
	keys := make([]string, len(dict))
	orders := make(map[string]int64, len(dict))
	i := 0

	for key, v := range dict {
		keys[i], orders[key] = key, int64(len(dict))
		i++

		if order, ok := schemaXOrder(v); ok {
			orders[key] = order
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		if i, j := orders[keys[i]], orders[keys[j]]; i != j {
			return i < j
		}
		return keys[i] < keys[j]
	})
	return keys
}

func schemaXOrder(v *openapi3.SchemaRef) (int64, bool) {
	if v == nil {
		return 0, false
	}

	// YAML parsing picks up the x-order as a float64
	if order, ok := v.Extensions[extOrder].(float64); ok {
		return int64(order), true
	}

	if v.Value == nil {
		return 0, false
	}

	// if v.Value is set, then this is actually a `$ref`, and we should check if there's an x-order set on that

	// YAML parsing picks up the x-order as a float64
	if order, ok := v.Value.Extensions[extOrder].(float64); ok {
		return int64(order), true
	}

	return 0, false
}

// StringInArray checks whether the specified string is present in an array
// of strings
func StringInArray(str string, array []string) bool {
	for _, elt := range array {
		if elt == str {
			return true
		}
	}
	return false
}

// RefPathToObjName returns the name of referenced object without changes.
//
//	#/components/schemas/Foo -> Foo
//	#/components/parameters/Bar -> Bar
//	#/components/responses/baz_baz -> baz_baz
//	document.json#/Foo -> Foo
//	http://deepmap.com/schemas/document.json#/objObj -> objObj
//
// Does not check refPath correctness.
func RefPathToObjName(refPath string) string {
	parts := strings.Split(refPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// RefPathToGoType takes a $ref value and converts it to a Go typename.
// #/components/schemas/Foo -> Foo
// #/components/parameters/Bar -> Bar
// #/components/responses/Baz -> Baz
// Remote components (document.json#/Foo) are supported if they present in --import-mapping
// URL components (http://deepmap.com/schemas/document.json#/Foo) are supported if they present in --import-mapping
// Remote and URL also support standard local paths even though the spec doesn't mention them.
func RefPathToGoType(refPath string) (string, error) {
	return refPathToGoType(refPath, true)
}

// refPathToGoType returns the Go typename for refPath given its
func refPathToGoType(refPath string, local bool) (string, error) {
	if refPath[0] == '#' {
		return refPathToGoTypeSelf(refPath, local)
	}
	pathParts := strings.Split(refPath, "#")
	if len(pathParts) != 2 {
		return "", fmt.Errorf("unsupported reference: %s", refPath)
	}
	remoteComponent, flatComponent := pathParts[0], pathParts[1]
	goPkg, ok := globalState.importMapping[remoteComponent]

	if !ok {
		return "", fmt.Errorf("unrecognized external reference '%s'; please provide the known import for this reference using option --import-mapping", remoteComponent)
	}

	if goPkg.Path == importMappingCurrentPackage {
		return refPathToGoTypeSelf(fmt.Sprintf("#%s", pathParts[1]), local)
	}

	return refPathToGoTypeRemote(flatComponent, goPkg)

}

func refPathToGoTypeSelf(refPath string, local bool) (string, error) {
	pathParts := strings.Split(refPath, "/")
	depth := len(pathParts)
	if local {
		if depth != 4 {
			return "", fmt.Errorf("unexpected reference depth: %d for ref: %s local: %t", depth, refPath, local)
		}
	} else if depth != 4 && depth != 2 {
		return "", fmt.Errorf("unexpected reference depth: %d for ref: %s local: %t", depth, refPath, local)
	}

	// Schemas may have been renamed locally, so look up the actual name in
	// the spec.
	name, err := findSchemaNameByRefPath(refPath, globalState.spec)
	if err != nil {
		return "", fmt.Errorf("error finding ref: %s in spec: %v", refPath, err)
	}
	if name != "" {
		return name, nil
	}
	// lastPart now stores the final element of the type path. This is what
	// we use as the base for a type name.
	lastPart := pathParts[len(pathParts)-1]
	return SchemaNameToTypeName(lastPart), nil
}

func refPathToGoTypeRemote(flatComponent string, goPkg goImport) (string, error) {
	goType, err := refPathToGoType("#"+flatComponent, false)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", goPkg.Name, goType), nil
}

// IsGoTypeReference takes a $ref value and checks if it has link to go type.
// #/components/schemas/Foo                     -> true
// ./local/file.yml#/components/parameters/Bar  -> true
// ./local/file.yml                             -> false
// IsGoTypeReference can be used to check whether RefPathToGoType($ref) is possible.
func IsGoTypeReference(ref string) bool {
	return ref != "" && !IsWholeDocumentReference(ref)
}

// IsWholeDocumentReference takes a $ref value and checks if it is whole document reference.
// #/components/schemas/Foo                             -> false
// ./local/file.yml#/components/parameters/Bar          -> false
// ./local/file.yml                                     -> true
// http://deepmap.com/schemas/document.json             -> true
// http://deepmap.com/schemas/document.json#/Foo        -> false
func IsWholeDocumentReference(ref string) bool {
	return ref != "" && !strings.ContainsAny(ref, "#")
}

// SwaggerUriToIrisUri converts a OpenAPI style path URI with parameters to an
// Iris compatible path URI. We need to replace all of OpenAPI parameters with
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToIrisUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, ":$1")
}

// SwaggerUriToEchoUri converts a OpenAPI style path URI with parameters to an
// Echo compatible path URI. We need to replace all of OpenAPI parameters with
// ":param". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToEchoUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, ":$1")
}

// SwaggerUriToFiberUri converts a OpenAPI style path URI with parameters to a
// Fiber compatible path URI. We need to replace all of OpenAPI parameters with
// ":param". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToFiberUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, ":$1")
}

// SwaggerUriToChiUri converts a swagger style path URI with parameters to a
// Chi compatible path URI. We need to replace all Swagger parameters with
// "{param}". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToChiUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "{$1}")
}

// SwaggerUriToGinUri converts a swagger style path URI with parameters to a
// Gin compatible path URI. We need to replace all Swagger parameters with
// ":param". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToGinUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, ":$1")
}

// SwaggerUriToGorillaUri converts a swagger style path URI with parameters to a
// Gorilla compatible path URI. We need to replace all Swagger parameters with
// ":param". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToGorillaUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "{$1}")
}

// SwaggerUriToStdHttpUri converts a swagger style path URI with parameters to a
// Chi compatible path URI. We need to replace all Swagger parameters with
// "{param}". Valid input parameters are:
//
//	{param}
//	{param*}
//	{.param}
//	{.param*}
//	{;param}
//	{;param*}
//	{?param}
//	{?param*}
func SwaggerUriToStdHttpUri(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "{$1}")
}

// OrderedParamsFromUri returns the argument names, in order, in a given URI string, so for
// /path/{param1}/{.param2*}/{?param3}, it would return param1, param2, param3
func OrderedParamsFromUri(uri string) []string {
	matches := pathParamRE.FindAllStringSubmatch(uri, -1)
	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m[1]
	}
	return result
}

// ReplacePathParamsWithStr replaces path parameters of the form {param} with %s
func ReplacePathParamsWithStr(uri string) string {
	return pathParamRE.ReplaceAllString(uri, "%s")
}

// SortParamsByPath reorders the given parameter definitions to match those in the path URI.
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

// IsGoKeyword returns whether the given string is a go keyword
func IsGoKeyword(str string) bool {
	return token.IsKeyword(str)
}

// IsPredeclaredGoIdentifier returns whether the given string
// is a predefined go identifier.
//
// See https://golang.org/ref/spec#Predeclared_identifiers
func IsPredeclaredGoIdentifier(str string) bool {
	_, exists := predeclaredSet[str]
	return exists
}

// IsGoIdentity checks if the given string can be used as an identity
// in the generated code like a type name or constant name.
//
// See https://golang.org/ref/spec#Identifiers
func IsGoIdentity(str string) bool {
	for i, c := range str {
		if !isValidRuneForGoID(i, c) {
			return false
		}
	}

	return IsGoKeyword(str)
}

func isValidRuneForGoID(index int, char rune) bool {
	if index == 0 && unicode.IsNumber(char) {
		return false
	}

	return unicode.IsLetter(char) || char == '_' || unicode.IsNumber(char)
}

// IsValidGoIdentity checks if the given string can be used as a
// name of variable, constant, or type.
func IsValidGoIdentity(str string) bool {
	if IsGoIdentity(str) {
		return false
	}

	return !IsPredeclaredGoIdentifier(str)
}

// SanitizeGoIdentity deletes and replaces the illegal runes in the given
// string to use the string as a valid identity.
func SanitizeGoIdentity(str string) string {
	sanitized := []rune(str)

	for i, c := range sanitized {
		if !isValidRuneForGoID(i, c) {
			sanitized[i] = '_'
		} else {
			sanitized[i] = c
		}
	}

	str = string(sanitized)

	if IsGoKeyword(str) || IsPredeclaredGoIdentifier(str) {
		str = "_" + str
	}

	if !IsValidGoIdentity(str) {
		panic("here is a bug")
	}

	return str
}

// SanitizeEnumNames fixes illegal chars in the enum names
// and removes duplicates
func SanitizeEnumNames(enumNames, enumValues []string) map[string]string {
	dupCheck := make(map[string]int, len(enumValues))
	deDup := make([][]string, 0, len(enumValues))

	for i, v := range enumValues {
		n := v
		if i < len(enumNames) {
			n = enumNames[i]
		}
		if _, dup := dupCheck[n]; !dup {
			deDup = append(deDup, []string{n, v})
		}
		dupCheck[n] = 0
	}

	dupCheck = make(map[string]int, len(deDup))
	sanitizedDeDup := make(map[string]string, len(deDup))

	for _, p := range deDup {
		n, v := p[0], p[1]
		sanitized := SanitizeGoIdentity(SchemaNameToTypeName(n))

		if _, dup := dupCheck[sanitized]; !dup {
			sanitizedDeDup[sanitized] = v
		} else {
			sanitizedDeDup[sanitized+strconv.Itoa(dupCheck[sanitized])] = v
		}
		dupCheck[sanitized]++
	}

	return sanitizedDeDup
}

func typeNamePrefix(name string) (prefix string) {
	if len(name) == 0 {
		return "Empty"
	}
	for _, r := range name {
		switch r {
		case '$':
			if len(name) == 1 {
				return "DollarSign"
			}
		case '-':
			prefix += "Minus"
		case '+':
			prefix += "Plus"
		case '&':
			prefix += "And"
		case '|':
			prefix += "Or"
		case '~':
			prefix += "Tilde"
		case '=':
			prefix += "Equal"
		case '>':
			prefix += "GreaterThan"
		case '<':
			prefix += "LessThan"
		case '#':
			prefix += "Hash"
		case '.':
			prefix += "Dot"
		case '*':
			prefix += "Asterisk"
		case '^':
			prefix += "Caret"
		case '%':
			prefix += "Percent"
		default:
			// Prepend "N" to schemas starting with a number
			if prefix == "" && unicode.IsDigit(r) {
				return "N"
			}

			// break the loop, done parsing prefix
			return
		}
	}

	return
}

// SchemaNameToTypeName converts a Schema name to a valid Go type name. It converts to camel case, and makes sure the name is
// valid in Go
func SchemaNameToTypeName(name string) string {
	return typeNamePrefix(name) + nameNormalizer(name)
}

// According to the spec, additionalProperties may be true, false, or a
// schema. If not present, true is implied. If it's a schema, true is implied.
// If it's false, no additional properties are allowed. We're going to act a little
// differently, in that if you want additionalProperties code to be generated,
// you must specify an additionalProperties type
// If additionalProperties it true/false, this field will be non-nil.
func SchemaHasAdditionalProperties(schema *openapi3.Schema) bool {
	if schema.AdditionalProperties.Has != nil && *schema.AdditionalProperties.Has {
		return true
	}

	if schema.AdditionalProperties.Schema != nil {
		return true
	}
	return false
}

// PathToTypeName converts a path, like Object/field1/nestedField into a go
// type name.
func PathToTypeName(path []string) string {
	for i, p := range path {
		path[i] = nameNormalizer(p)
	}
	return strings.Join(path, "_")
}

// StringToGoComment renders a possible multi-line string as a valid Go-Comment.
// Each line is prefixed as a comment.
func StringToGoComment(in string) string {
	return stringToGoCommentWithPrefix(in, "")
}

// StringWithTypeNameToGoComment renders a possible multi-line string as a
// valid Go-Comment, including the name of the type being referenced. Each line
// is prefixed as a comment.
func StringWithTypeNameToGoComment(in, typeName string) string {
	return stringToGoCommentWithPrefix(in, typeName)
}

func DeprecationComment(reason string) string {
	content := "Deprecated:" // The colon is required at the end even without reason
	if reason != "" {
		content += fmt.Sprintf(" %s", reason)
	}

	return stringToGoCommentWithPrefix(content, "")
}

func stringToGoCommentWithPrefix(in, prefix string) string {
	if len(in) == 0 || len(strings.TrimSpace(in)) == 0 { // ignore empty comment
		return ""
	}

	// Normalize newlines from Windows/Mac to Linux
	in = strings.ReplaceAll(in, "\r\n", "\n")
	in = strings.ReplaceAll(in, "\r", "\n")

	// Add comment to each line
	var lines []string
	for i, line := range strings.Split(in, "\n") {
		s := "//"
		if i == 0 && len(prefix) > 0 {
			s += " " + prefix
		}
		lines = append(lines, fmt.Sprintf("%s %s", s, line))
	}
	in = strings.Join(lines, "\n")

	// in case we have a multiline string which ends with \n, we would generate
	// empty-line-comments, like `// `. Therefore remove this line comment.
	in = strings.TrimSuffix(in, "\n// ")
	return in
}

// EscapePathElements breaks apart a path, and looks at each element. If it's
// not a path parameter, eg, {param}, it will URL-escape the element.
func EscapePathElements(path string) string {
	elems := strings.Split(path, "/")
	for i, e := range elems {
		if strings.HasPrefix(e, "{") && strings.HasSuffix(e, "}") {
			// This is a path parameter, we don't want to mess with its value
			continue
		}
		elems[i] = url.QueryEscape(e)
	}
	return strings.Join(elems, "/")
}

// renameSchema takes as input the name of a schema as provided in the spec,
// and the definition of the schema. If the schema overrides the name via
// x-go-name, the new name is returned, otherwise, the original name is
// returned.
func renameSchema(schemaName string, schemaRef *openapi3.SchemaRef) (string, error) {
	// References will not change type names.
	if schemaRef.Ref != "" {
		return SchemaNameToTypeName(schemaName), nil
	}
	schema := schemaRef.Value

	if extension, ok := schema.Extensions[extGoName]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return "", fmt.Errorf("invalid value for %q: %w", extPropGoType, err)
		}
		return typeName, nil
	}
	return SchemaNameToTypeName(schemaName), nil
}

// renameParameter generates the name for a parameter, taking x-go-name into
// account
func renameParameter(parameterName string, parameterRef *openapi3.ParameterRef) (string, error) {
	if parameterRef.Ref != "" {
		return SchemaNameToTypeName(parameterName), nil
	}
	parameter := parameterRef.Value

	if extension, ok := parameter.Extensions[extGoName]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return "", fmt.Errorf("invalid value for %q: %w", extPropGoType, err)
		}
		return typeName, nil
	}
	return SchemaNameToTypeName(parameterName), nil
}

// renameResponse generates the name for a parameter, taking x-go-name into
// account
func renameResponse(responseName string, responseRef *openapi3.ResponseRef) (string, error) {
	if responseRef.Ref != "" {
		return SchemaNameToTypeName(responseName), nil
	}
	response := responseRef.Value

	if extension, ok := response.Extensions[extGoName]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return "", fmt.Errorf("invalid value for %q: %w", extPropGoType, err)
		}
		return typeName, nil
	}
	return SchemaNameToTypeName(responseName), nil
}

// renameRequestBody generates the name for a parameter, taking x-go-name into
// account
func renameRequestBody(requestBodyName string, requestBodyRef *openapi3.RequestBodyRef) (string, error) {
	if requestBodyRef.Ref != "" {
		return SchemaNameToTypeName(requestBodyName), nil
	}
	requestBody := requestBodyRef.Value

	if extension, ok := requestBody.Extensions[extGoName]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return "", fmt.Errorf("invalid value for %q: %w", extPropGoType, err)
		}
		return typeName, nil
	}
	return SchemaNameToTypeName(requestBodyName), nil
}

// findSchemaByRefPath turns a $ref path into a schema. This will return ""
// if the schema wasn't found, and it'll only work successfully for schemas
// defined within the spec that we parsed.
func findSchemaNameByRefPath(refPath string, spec *openapi3.T) (string, error) {
	if spec.Components == nil {
		return "", nil
	}
	pathElements := strings.Split(refPath, "/")
	// All local references will have 4 path elements.
	if len(pathElements) != 4 {
		return "", nil
	}

	// We only support local references
	if pathElements[0] != "#" {
		return "", nil
	}
	// Only components are supported
	if pathElements[1] != "components" {
		return "", nil
	}
	propertyName := pathElements[3]
	switch pathElements[2] {
	case "schemas":
		if schema, found := spec.Components.Schemas[propertyName]; found {
			return renameSchema(propertyName, schema)
		}
	case "parameters":
		if parameter, found := spec.Components.Parameters[propertyName]; found {
			return renameParameter(propertyName, parameter)
		}
	case "responses":
		if response, found := spec.Components.Responses[propertyName]; found {
			return renameResponse(propertyName, response)
		}
	case "requestBodies":
		if requestBody, found := spec.Components.RequestBodies[propertyName]; found {
			return renameRequestBody(propertyName, requestBody)
		}
	}
	return "", nil
}

func ParseGoImportExtension(v *openapi3.SchemaRef) (*goImport, error) {
	if v.Value.Extensions[extPropGoImport] == nil || v.Value.Extensions[extPropGoType] == nil {
		return nil, nil
	}

	goTypeImportExt := v.Value.Extensions[extPropGoImport]

	importI, ok := goTypeImportExt.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert type: %T", goTypeImportExt)
	}

	gi := goImport{}
	// replicate the case-insensitive field mapping json.Unmarshal would do
	for k, v := range importI {
		if strings.EqualFold(k, "name") {
			if vs, ok := v.(string); ok {
				gi.Name = vs
			} else {
				return nil, fmt.Errorf("failed to convert type: %T", v)
			}
		} else if strings.EqualFold(k, "path") {
			if vs, ok := v.(string); ok {
				gi.Path = vs
			} else {
				return nil, fmt.Errorf("failed to convert type: %T", v)
			}
		}
	}

	return &gi, nil
}

func MergeImports(dst, src map[string]goImport) {
	for k, v := range src {
		dst[k] = v
	}
}

// TypeDefinitionsEquivalent checks for equality between two type definitions, but
// not every field is considered. We only want to know if they are fundamentally
// the same type.
func TypeDefinitionsEquivalent(t1, t2 TypeDefinition) bool {
	if t1.TypeName != t2.TypeName {
		return false
	}
	return reflect.DeepEqual(t1.Schema.OAPISchema, t2.Schema.OAPISchema)
}

// isAdditionalPropertiesExplicitFalse determines whether an openapi3.Schema is explicitly defined as `additionalProperties: false`
func isAdditionalPropertiesExplicitFalse(s *openapi3.Schema) bool {
	if s.AdditionalProperties.Has == nil {
		return false
	}

	return *s.AdditionalProperties.Has == false //nolint:gosimple
}

func sliceContains[E comparable](s []E, v E) bool {
	for _, ss := range s {
		if ss == v {
			return true
		}
	}
	return false
}
