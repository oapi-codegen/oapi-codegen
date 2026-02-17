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
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

const (
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
	var caseClauses = make(map[string]string)

	// Get the type definitions from the operation:
	typeDefinitions, err := op.GetResponseTypeDefinitions()
	if err != nil {
		panic(err)
	}

	typeDefTable := make(map[string]map[string]*ResponseTypeDefinition)
	for i := range typeDefinitions {
		t := &typeDefinitions[i]
		if typeDefTable[t.ResponseName] == nil {
			typeDefTable[t.ResponseName] = make(map[string]*ResponseTypeDefinition)
		}
		typeDefTable[t.ResponseName][t.ContentTypeName] = t
	}

	// Add a case for each possible response:
	buffer := new(bytes.Buffer)
	responses := op.Spec.Responses

	hasWildcardResponse := false
testWildcardResponseLoop:
	for _, responseName := range []string{"default", "1XX", "2XX", "3XX", "4XX", "5XX"} {
		if responseRef := responses.Value(responseName); responseRef != nil &&
			responseRef.Value != nil {
			for contentTypeName := range responseRef.Value.Content {
				if typeDefinition := typeDefTable[responseName][contentTypeName]; typeDefinition != nil {
					hasWildcardResponse = true
					break testWildcardResponseLoop
				}
			}
		}
	}

	for responseName, responseRef := range responses.Map() {
		keyPrefix := ""
		switch responseName {
		case "default":
			keyPrefix = "2XXX"
		case "1XX", "2XX", "3XX", "4XX", "5XX":
			keyPrefix = "1" + responseName
		default:
			keyPrefix = "0" + responseName
		}

		// We can't do much without a value:
		if responseRef.Value == nil {
			fmt.Fprintf(os.Stderr, "Response %s.%s has nil value\n", op.OperationId, responseName)
			continue
		}

		// If there is no content-type then we have no unmarshaling to do:
		if len(responseRef.Value.Content) == 0 {
			if hasWildcardResponse {
				caseAction := "break // No content-type"
				caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", responseName) + ":"
				caseClauses[keyPrefix+caseClauseKey+"1"] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			}
			continue
		}

		// If we made it this far then we need to handle unmarshaling for each content-type:
		jsonCount := 0
		for contentTypeName := range responseRef.Value.Content {
			if StringInArray(contentTypeName, contentTypesJSON) || util.IsMediaTypeJson(contentTypeName) {
				jsonCount++
			}
		}

		for contentTypeName := range responseRef.Value.Content {
			if typeDefinition := typeDefTable[responseName][contentTypeName]; typeDefinition == nil {
				// no type definition
				if hasWildcardResponse {
					caseAction := fmt.Sprintf("// Content-type (%s) unsupported", contentTypeName)
					caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", responseName) + ":"
					caseClauses[keyPrefix+caseClauseKey+"1"] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
				}
			} else {
				// We get "interface{}" when using "anyOf" or "oneOf" (which doesn't work with Go types):
				if typeDefinition.TypeName == "interface{}" {
					// Unable to unmarshal this, so we leave it out:
					continue
				}

				// Add content-types here (json / yaml / xml etc):
				switch {

				// JSON:
				case StringInArray(contentTypeName, contentTypesJSON) || util.IsMediaTypeJson(contentTypeName):
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := json.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)

					if jsonCount > 1 {
						caseKey, caseClause := buildUnmarshalCaseStrict(*typeDefinition, caseAction, contentTypeName)
						caseClauses[keyPrefix+"0"+caseKey] = caseClause
					} else {
						caseKey, caseClause := buildUnmarshalCase(*typeDefinition, caseAction, "json")
						caseClauses[keyPrefix+"0"+caseKey] = caseClause
					}

				// YAML:
				case StringInArray(contentTypeName, contentTypesYAML):
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := yaml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)
					caseKey, caseClause := buildUnmarshalCase(*typeDefinition, caseAction, "yaml")
					caseClauses[keyPrefix+"0"+caseKey] = caseClause

				// XML:
				case StringInArray(contentTypeName, contentTypesXML):
					caseAction := fmt.Sprintf("var dest %s\n"+
						"if err := xml.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)
					caseKey, caseClause := buildUnmarshalCase(*typeDefinition, caseAction, "xml")
					caseClauses[keyPrefix+"0"+caseKey] = caseClause
				}
			}
		}
	}

	if len(caseClauses) == 0 {
		// switch would be empty.
		return ""
	}

	// Now build the switch statement in order of most-to-least specific:
	// See: https://github.com/oapi-codegen/oapi-codegen/issues/127 for why we handle this in two separate
	// groups.
	fmt.Fprintf(buffer, "switch {\n")
	for _, caseClauseKey := range SortedMapKeys(caseClauses) {

		fmt.Fprintf(buffer, "%s\n", caseClauses[caseClauseKey])
	}
	fmt.Fprintf(buffer, "}\n")

	return buffer.String()
}

// buildUnmarshalCase builds an unmarshaling case clause for different content-types:
func buildUnmarshalCase(typeDefinition ResponseTypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf(".%s.%s", contentType, typeDefinition.ResponseName)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	contentTypeLiteral := StringToGoString(contentType)
	caseClause = fmt.Sprintf("case strings.Contains(rsp.Header.Get(\"%s\"), %s) && %s:\n%s\n", "Content-Type", contentTypeLiteral, caseClauseKey, caseAction)
	return caseKey, caseClause
}

func buildUnmarshalCaseStrict(typeDefinition ResponseTypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf(".%s.%s", contentType, typeDefinition.ResponseName)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	contentTypeLiteral := StringToGoString(contentType)
	caseClause = fmt.Sprintf("case rsp.Header.Get(\"%s\") == %s && %s:\n%s\n", "Content-Type", contentTypeLiteral, caseClauseKey, caseAction)
	return caseKey, caseClause
}

// genResponseTypeName creates the name of generated response types (given the operationID):
func genResponseTypeName(operationID string) string {
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

// This outputs a string array
func toStringArray(sarr []string) string {
	s := strings.Join(sarr, `","`)
	if len(s) > 0 {
		s = `"` + s + `"`
	}
	return `[]string{` + s + `}`
}

func stripNewLines(s string) string {
	r := strings.NewReplacer("\n", "")
	return r.Replace(s)
}

// genServerURLWithVariablesFunctionParams is a template helper method to generate the function parameters for the generated function for a Server object that contains `variables` (https://spec.openapis.org/oas/v3.0.3#server-object)
//
// goTypePrefix is the prefix being used to create underlying types in the template (likely the `ServerObjectDefinition.GoName`)
// variables are this `ServerObjectDefinition`'s variables for the Server object (likely the `ServerObjectDefinition.OAPISchema`)
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

// TemplateFunctions is passed to the template engine, and we can call each
// function here by keyName from the template code.
var TemplateFunctions = template.FuncMap{
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
	"getResponseTypeDefinitions": getResponseTypeDefinitions,
	"toStringArray":              toStringArray,
	"lower":                      strings.ToLower,
	"title":                      titleCaser.String,
	"stripNewLines":              stripNewLines,
	"sanitizeGoIdentity":         SanitizeGoIdentity,
	"toGoString":                 StringToGoString,
	"toGoComment":                StringWithTypeNameToGoComment,

	"genServerURLWithVariablesFunctionParams": genServerURLWithVariablesFunctionParams,
}
