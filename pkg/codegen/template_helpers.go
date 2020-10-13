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

	"github.com/labstack/echo/v4"
)

const (
	// These allow the case statements to be sorted later:
	prefixMostSpecific, prefixLessSpecific, prefixLeastSpecific = "3", "6", "9"
	responseTypeSuffix                                          = "Response"
)

var (
	contentTypesJSON = []string{echo.MIMEApplicationJSON, "text/x-json"}
	contentTypesYAML = []string{"application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml"}
	contentTypesXML  = []string{echo.MIMEApplicationXML, echo.MIMETextXML}
)

// This function takes an array of Parameter definition, and generates a valid
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

// This function is much like the one above, except it only produces the
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

func genOperationContext(op *OperationDefinition) string {
	var buffer = bytes.NewBufferString("")

	buffer.WriteString(fmt.Sprintf(`type %sContext struct {
	echo.Context
}`, op.OperationId))

	buffer.WriteString(genResponseHelpers(op))
	buffer.WriteString(genRequestHelpers(op))

	return buffer.String()
}

func genResponseHelpers(op *OperationDefinition) string {
	var buffer = bytes.NewBufferString("")

	responses := op.Spec.Responses
	sortedResponsesKeys := SortedResponsesKeys(responses)
	for _, responseName := range sortedResponsesKeys {
		responseRef := responses[responseName]

		if responseName == "default" {
			continue
		}

		// We can only generate a type if we have a value:
		if responseRef.Value != nil {
			sortedContentKeys := SortedContentKeys(responseRef.Value.Content)
			for _, contentTypeName := range sortedContentKeys {
				contentType := responseRef.Value.Content[contentTypeName]
				// We can only generate a type if we have a schema:
				if contentType.Schema != nil {
					responseSchema, err := GenerateGoSchema(contentType.Schema, []string{responseName})
					if err != nil {
						err = fmt.Errorf("Unable to determine Go type for %s.%s %w", op.OperationId, contentTypeName, err)
						panic(err)
					}

					var typeName string
					switch {

					case StringInArray(contentTypeName, contentTypesJSON):
						typeName = "JSON"

					case StringInArray(contentTypeName, contentTypesXML):
						typeName = "XML"

					case StringInArray(contentTypeName, contentTypesYAML):
						buffer.WriteString(fmt.Sprintf(`
func (c *%sContext) YAML%s(resp %s) error {
	err := c.Validate(resp)
	if err != nil {
		return err
	}
	var out []byte
	var err error
	out, err = yaml.Marshal(resp)
	if err != nil {
		return err
	}
	return c.Blob(%s, "%s", out)
}
`, op.OperationId, responseName, responseSchema.TypeDecl(), responseName, contentTypeName))
						continue

					default:
						continue
					}

					buffer.WriteString(fmt.Sprintf(`
func (c *%sContext) %s(resp %s) error {
	err := c.Validate(resp)
	if err != nil {
		return err
	}
	return c.%s(%s, resp)
}
`, op.OperationId, typeName+responseName, responseSchema.TypeDecl(), typeName, responseName))

				}
			}
		}
	}
	return buffer.String()
}

func genRequestHelpers(op *OperationDefinition) string {
	var buffer = bytes.NewBufferString("")

	requestBody := op.Spec.RequestBody
	if requestBody == nil || requestBody.Value == nil {
		return ""
	}

	included := map[string]string{}

	requestBodyDefinitions, _, err := GenerateBodyDefinitions(op.OperationId, requestBody)
	if err != nil {
		panic(err)
	}

	for _, rbd := range requestBodyDefinitions {
		contentTypeName := rbd.ContentType
		var typeName string
		switch {

		// JSON:
		case StringInArray(contentTypeName, contentTypesJSON):
			typeName = "JSON"
		case StringInArray(contentTypeName, contentTypesXML):
			typeName = "XML"
		default:
			continue
		}
		if previousCT, ok := included[typeName]; ok {
			panic(fmt.Sprintf("%s: failed to add %s because of %s", typeName, contentTypeName, previousCT))
		} else {
			included[typeName] = contentTypeName
		}

		fmt.Fprintf(buffer, `
func (c *%sContext) Bind%s() (*%s, error) {
	var err error
`, op.OperationId, typeName, rbd.Schema.TypeDecl())
		buffer.WriteString(`
	// optional
	if c.Request().ContentLength == 0 {`)
		if !requestBody.Value.Required {
			buffer.WriteString(`return nil, nil`)
		} else {
			buffer.WriteString(`return nil, errors.New("the request body should not be empty")`)
		}
		buffer.WriteString("\t}\n")

		fmt.Fprintf(buffer, `
	ctype := c.Request().Header.Get(echo.HeaderContentType)
	if ctype != "%s" {
		err = errors.New(fmt.Sprintf("incorrect content type: %%s", ctype))
		return nil, err
	}

	var result %s
	if err = c.Bind(&result); err != nil {
		return nil, err
	}

	if err = c.Validate(result); err != nil {
		return nil, err
	}

	return &result, nil
}
`, contentTypeName, rbd.Schema.TypeDecl())

	}

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
	for _, typeDefinition := range typeDefinitions {

		responseRef, ok := responses[typeDefinition.ResponseName]
		if !ok {
			continue
		}

		// We can't do much without a value:
		if responseRef.Value == nil {
			fmt.Fprintf(os.Stderr, "Response %s.%s has nil value\n", op.OperationId, typeDefinition.ResponseName)
			continue
		}

		// If there is no content-type then we have no unmarshaling to do:
		if len(responseRef.Value.Content) == 0 {
			caseAction := "break // No content-type"
			caseClauseKey := "case " + getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName) + ":"
			unhandledCaseClauses[prefixLeastSpecific+caseClauseKey] = fmt.Sprintf("%s\n%s\n", caseClauseKey, caseAction)
			continue
		}

		// If we made it this far then we need to handle unmarshaling for each content-type:
		sortedContentKeys := SortedContentKeys(responseRef.Value.Content)
		for _, contentTypeName := range sortedContentKeys {

			// We get "interface{}" when using "anyOf" or "oneOf" (which doesn't work with Go types):
			if typeDefinition.TypeName == "interface{}" {
				// Unable to unmarshal this, so we leave it out:
				continue
			}

			// Add content-types here (json / yaml / xml etc):
			switch {

			// JSON:
			case StringInArray(contentTypeName, contentTypesJSON):
				if typeDefinition.ContentTypeName == contentTypeName {
					var caseAction string

					caseAction = fmt.Sprintf("var dest %s\n"+
						"if err := json.Unmarshal(bodyBytes, &dest); err != nil { \n"+
						" return nil, err \n"+
						"}\n"+
						"response.%s = &dest",
						typeDefinition.Schema.TypeDecl(),
						typeDefinition.TypeName)

					caseKey, caseClause := buildUnmarshalCase(typeDefinition, caseAction, "json")
					handledCaseClauses[caseKey] = caseClause
				}

			// YAML:
			case StringInArray(contentTypeName, contentTypesYAML):
				if typeDefinition.ContentTypeName == contentTypeName {
					var caseAction string
					caseAction = fmt.Sprintf("var dest %s\n"+
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
			case StringInArray(contentTypeName, contentTypesXML):
				if typeDefinition.ContentTypeName == contentTypeName {
					var caseAction string
					caseAction = fmt.Sprintf("var dest %s\n"+
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

	if len(handledCaseClauses)+len(unhandledCaseClauses) == 0 {
		// switch would be empty.
		return ""
	}

	// Now build the switch statement in order of most-to-least specific:
	// See: https://github.com/deepmap/oapi-codegen/issues/127 for why we handle this in two separate
	// groups.
	fmt.Fprintf(buffer, "switch {\n")
	for _, caseClauseKey := range SortedStringKeys(handledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", handledCaseClauses[caseClauseKey])
	}
	for _, caseClauseKey := range SortedStringKeys(unhandledCaseClauses) {

		fmt.Fprintf(buffer, "%s\n", unhandledCaseClauses[caseClauseKey])
	}
	fmt.Fprintf(buffer, "}\n")

	return buffer.String()
}

// buildUnmarshalCase builds an unmarshalling case clause for different content-types:
func buildUnmarshalCase(typeDefinition ResponseTypeDefinition, caseAction string, contentType string) (caseKey string, caseClause string) {
	caseKey = fmt.Sprintf("%s.%s.%s", prefixLeastSpecific, contentType, typeDefinition.ResponseName)
	caseClauseKey := getConditionOfResponseName("rsp.StatusCode", typeDefinition.ResponseName)
	caseClause = fmt.Sprintf("case strings.Contains(rsp.Header.Get(\"%s\"), \"%s\") && %s:\n%s\n", echo.HeaderContentType, contentType, caseClauseKey, caseAction)
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
	return `[]string{"` + strings.Join(sarr, `","`) + `"}`
}

func stripNewLines(s string) string {
	r := strings.NewReplacer("\n", "")
	return r.Replace(s)
}

// This function map is passed to the template engine, and we can call each
// function here by keyName from the template code.
var TemplateFunctions = template.FuncMap{
	"genParamArgs":               genParamArgs,
	"genParamTypes":              genParamTypes,
	"genParamNames":              genParamNames,
	"genParamFmtString":          ReplacePathParamsWithStr,
	"swaggerUriToEchoUri":        SwaggerUriToEchoUri,
	"swaggerUriToChiUri":         SwaggerUriToChiUri,
	"swaggerUriToGinUri":         SwaggerUriToGinUri,
	"lcFirst":                    LowercaseFirstCharacter,
	"ucFirst":                    UppercaseFirstCharacter,
	"camelCase":                  ToCamelCase,
	"genResponsePayload":         genResponsePayload,
	"genResponseTypeName":        genResponseTypeName,
	"genResponseUnmarshal":       genResponseUnmarshal,
	"genOperationContext":        genOperationContext,
	"getResponseTypeDefinitions": getResponseTypeDefinitions,
	"toStringArray":              toStringArray,
	"lower":                      strings.ToLower,
	"title":                      strings.Title,
	"stripNewLines":              stripNewLines,
	"sanitizeGoIdentity":         SanitizeGoIdentity,
}
