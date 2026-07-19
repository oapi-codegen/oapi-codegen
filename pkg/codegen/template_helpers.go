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
	"os"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

const (
	// These allow the case statements to be sorted later:
	prefixLeastSpecific = "9"

	defaultClientTypeName = "Client"
)

var (
	contentTypesJSON    = []string{"application/json", "text/x-json", "application/problem+json"}
	contentTypesHalJSON = []string{"application/hal+json"}
	contentTypesYAML    = []string{"application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml"}
	contentTypesXML     = []string{"application/xml", "text/xml", "application/problems+xml"}

	responseTypeSuffix = "Response"

	titleCaser = cases.Title(language.English)
)

// genParamArgs takes an array of Parameter definition, and generates a valid
// Go parameter declaration from them, eg:
// ", foo int, bar string, baz float32". The preceding comma is there to save
// a lot of work in the template engine.
func genParamArgs(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		paramName := p.GoVariableName()
		parts[i] = fmt.Sprintf("%s %s", paramName, p.TypeDef())
	}
	return ", " + strings.Join(parts, ", ")
}

// genParamTypes is much like the one above, except it only produces the
// types of the parameters for a type declaration. It would produce this
// from the same input as above:
// ", int, string, float32".
func genParamTypes(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.TypeDef()
	}
	return ", " + strings.Join(parts, ", ")
}

// This is another variation of the function above which generates only the
// parameter names:
// ", foo, bar, baz"
func genParamNames(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = p.GoVariableName()
	}
	return ", " + strings.Join(parts, ", ")
}

// genResponsePayload generates the payload returned at the end of each client request function
func genResponsePayload(operationID string) string {
	var buffer = bytes.NewBufferString("")

	// Here is where we build up a response:
	fmt.Fprintf(buffer, "&%s{\n", genResponseTypeName(operationID))
	fmt.Fprintf(buffer, "Body: bodyBytes,\n")
	fmt.Fprintf(buffer, "HTTPResponse: rsp,\n")
	fmt.Fprintf(buffer, "}")

	return buffer.String()
}

// genResponseUnmarshal generates unmarshaling steps for structured response payloads
func genResponseUnmarshal(op *OperationDefinition) string {
	var handledCaseClauses = make(map[string]string)
	var unhandledCaseClauses = make(map[string]string)

	// Get the type definitions from the operation:
	typeDefinitions, err := op.GetResponseTypeDefinitions()
	if err != nil {
		panic(err)
	}

	if len(typeDefinitions) == 0 {
		// No types.
		return ""
	}

	// Add a case for each possible response:
	buffer := new(bytes.Buffer)
	responses := op.Spec.Responses
	handledResponseNames := make(map[string]struct{})
	for _, typeDefinition := range typeDefinitions {

		responseRef := responses.Value(typeDefinition.ResponseName)
		if responseRef == nil {
			continue
		}

		// We can't do much without a value:
		if responseRef.Value == nil {
			fmt.Fprintf(os.Stderr, "Response %s.%s has nil value\n", op.OperationId, typeDefinition.ResponseName)
			continue
		}

		handledResponseNames[typeDefinition.ResponseName] = struct{}{}

		// If there is no content-type then we have no unmarshaling to do:
		if len(responseRef.Value.Content) == 0 {
			caseAction := "break // No content-type"
			caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName) + ":"
			unhandledCaseClauses[prefixLeastSpecific+caseClauseKey] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			continue
		}

		// If we made it this far then we need to handle unmarshaling for each content-type:
		SortedMapKeys := SortedMapKeys(responseRef.Value.Content)
		jsonCount := 0
		for _, contentTypeName := range SortedMapKeys {
			if slices.Contains(contentTypesJSON, contentTypeName) || util.IsMediaTypeJson(contentTypeName) {
				jsonCount++
			}
		}

		for _, contentTypeName := range SortedMapKeys {

			// We get "interface{}" when using "anyOf" or "oneOf" (which doesn't work with Go types):
			if typeDefinition.TypeName == "interface{}" {
				// Unable to unmarshal this, so we leave it out:
				continue
			}

			// Add content-types here (json / yaml / xml etc):
			switch {

			// JSON:
			case slices.Contains(contentTypesJSON, contentTypeName) || util.IsMediaTypeJson(contentTypeName):
				if typeDefinition.ContentTypeName == contentTypeName {
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := json.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)

					if jsonCount > 1 {
						caseKey, caseClause := buildUnmarshalCaseStrict(typeDefinition, caseAction, contentTypeName)
						handledCaseClauses[caseKey] = caseClause
					} else {
						caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "json")
						handledCaseClauses[caseKey] = caseClause
					}
				}

			// YAML:
			case slices.Contains(contentTypesYAML, contentTypeName):
				if typeDefinition.ContentTypeName == contentTypeName {
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := yaml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)
					caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "yaml")
					handledCaseClauses[caseKey] = caseClause
				}

			// XML:
			case slices.Contains(contentTypesXML, contentTypeName):
				if typeDefinition.ContentTypeName == contentTypeName {
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := xml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)
					caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "xml")
					handledCaseClauses[caseKey] = caseClause
				}

			// Everything else:
			default:
				caseAction := fmt.Sprintf("// Content-type (%s) unsupported", contentTypeName)
				caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName) + ":"
				unhandledCaseClauses[prefixLeastSpecific+caseClauseKey] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			}
		}
	}

	// Emit explicit case clauses for responses declared without content (e.g.
	// "204 No Content"). These responses have no type definitions, so the loop
	// above never visits them. Without an explicit case clause, the "default"
	// catch-all (which matches with "&& true") would attempt to unmarshal an
	// empty body and fail.
	//
	// Keys follow the same "<responseName>.<detail>" scheme as
	// buildUnmarshalCase, so an exact bodyless code (e.g. "204") sorts among
	// its numeric peers, a range wildcard (e.g. "2XX") sorts after every
	// explicit content case it covers (preventing it from shadowing them) but
	// before the default catch-all, which sorts last.
	//
	// "default" itself is skipped — a bodyless default would emit "case true:"
	// which shadows everything. Such a spec is degenerate and the existing
	// behaviour (no guard) is acceptable.
	if responses != nil {
		for _, responseName := range SortedMapKeys(responses.Map()) {
			if _, handled := handledResponseNames[responseName]; handled {
				continue
			}
			if responseName == "default" {
				continue
			}
			responseRef := responses.Value(responseName)
			if responseRef == nil || responseRef.Value == nil {
				continue
			}
			if len(responseRef.Value.Content) == 0 {
				caseCondition := getConditionOfResponseName("rsp.StatusCode", responseName)
				caseClause := fmt.Sprintf("case %s:\nbreak // No content-type\n", caseCondition)
				caseKey := fmt.Sprintf("%s.%s.nocontent", prefixLeastSpecific, responseName)
				handledCaseClauses[caseKey] = caseClause
			}
		}
	}

	if len(handledCaseClauses)+len(unhandledCaseClauses) == 0 {
		// switch would be empty.
		return ""
	}

	// Now build the switch statement in order of most-to-least specific:
	// See: https://github.com/oapi-codegen/oapi-codegen/issues/127 for why we handle this in two separate
	// groups.
	fmt.Fprintf(buffer, "switch {\n")
	for _, caseClauseKey := range SortedMapKeys(handledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", handledCaseClauses[caseClauseKey])
	}
	for _, caseClauseKey := range SortedMapKeys(unhandledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", unhandledCaseClauses[caseClauseKey])
	}
	fmt.Fprintf(buffer, "}\n")

	return buffer.String()
}

// buildUnmarshalCase builds an unmarshaling case clause for different content-types.
//
// The sort key puts the response name before the content type so that
// lexicographic ordering of the keys yields most-specific-first case clauses:
// exact codes sort numerically ("200" < "204"), range wildcards sort after the
// exact codes they cover ("204" < "2XX" since 'X' > '9') and before the next
// tier ("2XX" < "300"), and "default" sorts after everything ('d' > 'X').
func buildUnmarshalCase(typeDefinition ResponseTypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf("%s.%s.%s", prefixLeastSpecific, typeDefinition.ResponseName, contentType)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	contentTypeLiteral := StringToGoString(contentType)
	caseClause = fmt.Sprintf("case strings.Contains(rsp.Header.Get(\"%s\"), %s) && %s:\n%s\n", "Content-Type", contentTypeLiteral, caseClauseKey, caseAction)
	return caseKey, caseClause
}

// buildUnmarshalCaseStrict is buildUnmarshalCase with an exact Content-Type
// match; it uses the same "<responseName>.<contentType>" key scheme.
func buildUnmarshalCaseStrict(typeDefinition ResponseTypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf("%s.%s.%s", prefixLeastSpecific, typeDefinition.ResponseName, contentType)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	contentTypeLiteral := StringToGoString(contentType)
	caseClause = fmt.Sprintf("case rsp.Header.Get(\"%s\") == %s && %s:\n%s\n", "Content-Type", contentTypeLiteral, caseClauseKey, caseAction)
	return caseKey, caseClause
}

// genResponseTypeName creates the name of generated response types (given the operationID).
// It first checks if the multi-pass name resolver has assigned a name for this
// wrapper type (which would happen if the default name collides with a schema type).
func genResponseTypeName(operationID string) string {
	if name, ok := globalState.resolvedClientWrapperNames[operationID]; ok {
		return name
	}
	return fmt.Sprintf("%s%s", UppercaseFirstCharacter(operationID), responseTypeSuffix)
}

func getResponseTypeDefinitions(op *OperationDefinition) []ResponseTypeDefinition {
	td, err := op.GetResponseTypeDefinitions()
	if err != nil {
		panic(err)
	}
	return td
}

// Return the statusCode comparison clause from the response name.
func getConditionOfResponseName(statusCodeVar, responseName string) string {
	switch responseName {
	case "default":
		return "true"
	case "1XX", "2XX", "3XX", "4XX", "5XX":
		return fmt.Sprintf("%s / 100 == %s", statusCodeVar, responseName[:1])
	default:
		return fmt.Sprintf("%s == %s", statusCodeVar, responseName)
	}
}

// responsesWithHeaders returns the subset of responses that declare headers,
// ordered most-specific-first for emission as switch case clauses: exact
// status codes sort numerically, range wildcards after the exact codes they
// cover ("204" < "2XX" since 'X' > '9'), and "default" (which compiles to
// `case true:`) after everything ('d' > 'X').
func responsesWithHeaders(responses []ResponseDefinition) []ResponseDefinition {
	var out []ResponseDefinition
	for _, response := range responses {
		if len(response.Headers) > 0 {
			out = append(out, response)
		}
	}
	slices.SortFunc(out, func(a, b ResponseDefinition) int {
		return strings.Compare(a.StatusCode, b.StatusCode)
	})
	return out
}

// This outputs a string array
func toStringArray(sarr []string) string {
	s := strings.Join(sarr, `","`)
	if len(s) > 0 {
		s = `"` + s + `"`
	}
	return `[]string{` + s + `}`
}

// stripNewLines removes newlines so untrusted spec text stays inside a single
// generated `//` line comment instead of breaking out into real Go source.
func stripNewLines(s string) string {
	r := strings.NewReplacer("\n", "", "\r", "")
	return r.Replace(s)
}

// genServerURLWithVariablesFunctionParams is a template helper method to generate the function parameters for the generated function for a Server object that contains `variables` (https://spec.openapis.org/oas/v3.0.3#server-object)
//
// goTypePrefix is the prefix being used to create underlying types in the template (likely the `ServerObjectDefinition.GoName`)
// variables are this `ServerObjectDefinition`'s variables for the Server object (likely the `ServerObjectDefinition.OAPISchema`)
//
// Undeclared `{name}` placeholders that appear in the URL but have no
// entry in `variables` are NOT handled here; they're emitted as plain
// `string` parameters by `ServerObjectDefinition.NewFunctionParams`,
// which the template calls instead of this helper. Custom
// `server-urls.tmpl` overrides that still call this helper directly
// keep their pre-existing two-argument signature.
func genServerURLWithVariablesFunctionParams(goTypePrefix string, variables map[string]*openapi3.ServerVariable) string {
	keys := SortedMapKeys(variables)

	if len(variables) == 0 {
		return ""
	}
	parts := make([]string, len(variables))

	for i := range keys {
		k := keys[i]
		variableDefinitionPrefix := goTypePrefix + UppercaseFirstCharacter(k) + "Variable"
		parts[i] = k + " " + variableDefinitionPrefix
	}
	return strings.Join(parts, ", ")
}

// httpMethodConstant converts an HTTP method string (e.g. "GET") to the
// corresponding Go net/http constant (e.g. "http.MethodGet").
func httpMethodConstant(method string) string {
	switch method {
	case "GET":
		return "http.MethodGet"
	case "POST":
		return "http.MethodPost"
	case "PUT":
		return "http.MethodPut"
	case "DELETE":
		return "http.MethodDelete"
	case "PATCH":
		return "http.MethodPatch"
	case "HEAD":
		return "http.MethodHead"
	case "OPTIONS":
		return "http.MethodOptions"
	case "TRACE":
		return "http.MethodTrace"
	default:
		return fmt.Sprintf("%q", method)
	}
}

// dict builds a map[string]any from an even-length list of alternating
// key/value arguments. It lets a template pass more than one named value to a
// {{template}} invocation (text/template only accepts a single data argument),
// e.g. {{template "partial" (dict "Headers" $headers "Setter" "w.Header().Set")}}.
func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments, got %d", len(values))
	}
	m := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings, got %T", values[i])
		}
		m[key] = values[i+1]
	}
	return m, nil
}

// TemplateFunctions is passed to the template engine, and we can call each
// function here by keyName from the template code.
var TemplateFunctions = template.FuncMap{
	"dict":                       dict,
	"genParamArgs":               genParamArgs,
	"genParamTypes":              genParamTypes,
	"genParamNames":              genParamNames,
	"genParamFmtString":          ReplacePathParamsWithStr,
	"swaggerUriToIrisUri":        SwaggerUriToIrisUri,
	"swaggerUriToEchoUri":        SwaggerUriToEchoUri,
	"swaggerUriToFiberUri":       SwaggerUriToFiberUri,
	"swaggerUriToChiUri":         SwaggerUriToChiUri,
	"swaggerUriToGinUri":         SwaggerUriToGinUri,
	"swaggerUriToGorillaUri":     SwaggerUriToGorillaUri,
	"swaggerUriToStdHttpUri":     SwaggerUriToStdHttpUri,
	"lcFirst":                    LowercaseFirstCharacter,
	"ucFirst":                    UppercaseFirstCharacter,
	"ucFirstWithPkgName":         UppercaseFirstCharacterWithPkgName,
	"camelCase":                  ToCamelCase,
	"genResponsePayload":         genResponsePayload,
	"genResponseTypeName":        genResponseTypeName,
	"genResponseUnmarshal":       genResponseUnmarshal,
	"getConditionOfResponseName": getConditionOfResponseName,
	"responsesWithHeaders":       responsesWithHeaders,
	"getResponseTypeDefinitions": getResponseTypeDefinitions,
	"toStringArray":              toStringArray,
	"lower":                      strings.ToLower,
	"title":                      titleCaser.String,
	"stripNewLines":              stripNewLines,
	"sanitizeGoIdentity":         SanitizeGoIdentity,
	"schemaNameToTypeName":       SchemaNameToTypeName,
	"toGoString":                 StringToGoString,
	"toGoComment":                StringWithTypeNameToGoComment,

	"genServerURLWithVariablesFunctionParams": genServerURLWithVariablesFunctionParams,
	"httpMethodConstant":                      httpMethodConstant,
}
