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
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/deepmap/oapi-codegen/pkg/codegen/templates"
)

// Options defines the optional code to generate.
type Options struct {
	GenerateServer bool // GenerateServer specifies to gen server handler.
	GenerateClient bool // GenerateClient specifies to gen client.
}

// Uses the Go templating engine to generate all of our server wrappers from
// the descriptions we've built up above from the schema objects.
// opts defines
func Generate(swagger *openapi3.Swagger, packageName string, opts Options) (string, error) {
	// This creates the golang templates text package
	t := template.New("oapi-codegen").Funcs(TemplateFunctions)
	// This parses all of our own template files into the template object
	// above
	t, err := templates.Parse(t)
	if err != nil {
		return "", fmt.Errorf("error parsing oapi-codegen templates: %s", err)
	}

	schemasOut, err := GenerateTypesForSchemas(t, swagger.Components.Schemas)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component schemas: %s", err)
	}

	paramsOut, err := GenerateTypesForParameters(t, swagger.Components.Parameters)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component parameters: %s", err)
	}

	responsesOut, err := GenerateTypesForResponses(t, swagger.Components.Responses)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component responses: %s", err)
	}

	bodiesOut, err := GenerateTypesForRquestBodies(t, swagger.Components.RequestBodies)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component request bodies: %s", err)
	}

	ops, err := OperationDefinitions(swagger)
	if err != nil {
		return "", fmt.Errorf("error creating operation definitions: %v", err)
	}

	var serverOut string
	if opts.GenerateServer {
		serverOut, err = GenerateServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %s", err)
		}
	}

	var clientOut string
	if opts.GenerateClient {
		clientOut, err = GenerateClient(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating client: %v", err)
		}
	}

	inlinedSpec, err := GenerateInlinedSpec(t, swagger)
	if err != nil {
		return "", fmt.Errorf("error generating Go handlers for Paths: %s", err)
	}

	// Imports needed for the generated code to compile
	imports := []string{
		"bytes",
		"compress/gzip",
		"encoding/base64",
		"fmt",
		"github.com/getkin/kin-openapi/openapi3",
		"github.com/deepmap/oapi-codegen/pkg/runtime",
		"strings",
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// Based on module prefixes, figure out which optional imports are required.
	for _, str := range []string{schemasOut, paramsOut, responsesOut, bodiesOut,
		serverOut, clientOut, inlinedSpec} {
		if strings.Contains(str, "time.Time") {
			imports = append(imports, "time")
		}
		if strings.Contains(str, "http.") {
			imports = append(imports, "net/http")
		}
		if strings.Contains(str, "openapi3.") {
			imports = append(imports, "github.com/getkin/kin-openapi/openapi3")
		}
		if strings.Contains(str, "json.") {
			imports = append(imports, "encoding/json")
		}
		if strings.Contains(str, "echo.") {
			imports = append(imports, "github.com/labstack/echo/v4")
		}
		if strings.Contains(str, "io.") {
			imports = append(imports, "io")
		}
		if strings.Contains(str, "url.") {
			imports = append(imports, "net/url")
		}
		if strings.Contains(str, "context.") {
			imports = append(imports, "context")
		}
	}

	importsOut, err := GenerateImports(t, imports, packageName)
	if err != nil {
		return "", fmt.Errorf("error generating imports: %s", err)
	}

	_, err = w.WriteString(importsOut)
	if err != nil {
		return "", fmt.Errorf("error writing imports: %s", err)
	}

	_, err = w.WriteString(schemasOut)
	if err != nil {
		return "", fmt.Errorf("error writing type definitions: %s", err)
	}

	_, err = w.WriteString(paramsOut)
	if err != nil {
		return "", fmt.Errorf("error writing parameter definitions: %s", err)
	}

	_, err = w.WriteString(responsesOut)
	if err != nil {
		return "", fmt.Errorf("error writing response definitions: %s", err)
	}

	_, err = w.WriteString(bodiesOut)
	if err != nil {
		return "", fmt.Errorf("error writing request body definitions: %s", err)
	}

	if opts.GenerateClient {
		_, err = w.WriteString(clientOut)
		if err != nil {
			return "", fmt.Errorf("error writing client: %v", err)
		}
	}

	if opts.GenerateServer {
		_, err = w.WriteString(serverOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %s", err)
		}
	}

	_, err = w.WriteString(inlinedSpec)
	if err != nil {
		return "", fmt.Errorf("error writing inlined spec: %s", err)
	}

	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer: %s", err)
	}

	goCode := buf.String()

	// The generation code produces unindented horrors. Use the Go formatter
	// to make it all pretty.
	outBytes, err := format.Source([]byte(goCode))
	if err != nil {
		fmt.Println(goCode)
		return "", fmt.Errorf("error formatting Go code: %s", err)
	}
	return string(outBytes), nil
}

// Generates type definitions for any custom types defined in the
// components/schemas section of the Swagger spec.
func GenerateTypesForSchemas(t *template.Template, schemas map[string]*openapi3.SchemaRef) (string, error) {
	types := make([]TypeDefinition, 0)
	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range SortedSchemaKeys(schemas) {
		schemaRef := schemas[schemaName]
		typeDef, err := schemaToGoType(schemaRef, true)
		if err != nil {
			return "", fmt.Errorf("error converting Schema %s to Go type: %s", schemaName, err)
		}

		types = append(types, TypeDefinition{
			JsonTypeName: schemaName,
			TypeName:     ToCamelCase(schemaName),
			TypeDef:      typeDef,
		})
	}

	typesOut, err := GenerateTypes(t, "schemas.tmpl", types)
	if err != nil {
		return "", fmt.Errorf("error generating type definitions: %s", err)
	}

	return typesOut, nil
}

// Generates type definitions for any custom types defined in the
// components/parameters section of the Swagger spec.
func GenerateTypesForParameters(t *template.Template, params map[string]*openapi3.ParameterRef) (string, error) {
	types := make([]TypeDefinition, 0)
	for _, paramName := range SortedParameterKeys(params) {
		paramOrRef := params[paramName]
		if paramOrRef.Ref != "" {
			// The entire definition of the parameter is an external reference
			goType, err := RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return "", fmt.Errorf("error generating Go type for (%s) in parameter %s: %s",
					paramOrRef.Ref, paramName, err)
			}
			types = append(types, TypeDefinition{
				JsonTypeName: paramName,
				TypeName:     ToCamelCase(paramName),
				TypeDef:      goType,
			})
		} else {
			// The parameter is defined inline
			goType, err := paramToGoType(paramOrRef.Value)
			if err != nil {
				return "", fmt.Errorf("error generating Go type for schema in parameter %s: %s",
					paramName, err)
			}
			types = append(types, TypeDefinition{
				JsonTypeName: paramName,
				TypeName:     ToCamelCase(paramName),
				TypeDef:      goType,
			})
		}
	}

	typesOut, err := GenerateTypes(t, "parameters.tmpl", types)
	if err != nil {
		return "", fmt.Errorf("error generating type definitions: %s", err)
	}

	return typesOut, nil
}

// Generates type definitions for any custom types defined in the
// components/responses section of the Swagger spec.
func GenerateTypesForResponses(t *template.Template, responses openapi3.Responses) (string, error) {
	types := make([]TypeDefinition, 0)
	for _, responseName := range SortedResponsesKeys(responses) {
		responseOrRef := responses[responseName]
		if responseOrRef.Ref != "" {
			// The entire response is defined as a reference
			goType, err := RefPathToGoType(responseOrRef.Ref)
			if err != nil {
				return "", fmt.Errorf("error generating Go type for (%s) in parameter %s: %s",
					responseOrRef.Ref, responseName, err)
			}
			types = append(types, TypeDefinition{
				JsonTypeName: responseName,
				TypeName:     ToCamelCase(responseName),
				TypeDef:      goType,
			})
		} else {
			// We have to generate the response object. We're only going to
			// handle application/json media types here. Other responses should
			// simply be specified as strings or byte arrays.
			response := responseOrRef.Value
			jsonResponse, found := response.Content["application/json"]
			if found {
				goType, err := schemaToGoType(jsonResponse.Schema, true)
				if err != nil {
					return "", fmt.Errorf("error generating Go type for schema in parameter %s: %s",
						responseName, err)
				}
				types = append(types, TypeDefinition{
					JsonTypeName: responseName,
					TypeName:     ToCamelCase(responseName),
					TypeDef:      goType,
				})
			}
		}
	}

	typesOut, err := GenerateTypes(t, "responses.tmpl", types)
	if err != nil {
		return "", fmt.Errorf("error generating type definitions: %s", err)
	}

	return typesOut, nil
}

// Generates type definitions for any custom types defined in the
// components/requestBodies section of the Swagger spec.
func GenerateTypesForRquestBodies(t *template.Template, bodies map[string]*openapi3.RequestBodyRef) (string, error) {
	types := make([]TypeDefinition, 0)
	for _, bodyName := range SortedRequestBodyKeys(bodies) {
		bodyOrRef := bodies[bodyName]
		if bodyOrRef.Ref != "" {
			ref := bodyOrRef.Ref
			goType, err := RefPathToGoType(ref)
			if err != nil {
				return "", fmt.Errorf("error generating Go type for (%s) in request body %s: %s",
					ref, bodyName, err)
			}
			types = append(types, TypeDefinition{
				JsonTypeName: bodyName,
				TypeName:     ToCamelCase(bodyName),
				TypeDef:      goType,
			})
		} else {
			body := bodyOrRef.Value
			// As for responses, we will only generate Go code for JSON bodies,
			// the other body formats are up to the user.
			jsonResponse, found := body.Content["application/json"]
			if found {
				goType, err := schemaToGoType(jsonResponse.Schema, true)
				if err != nil {
					return "", fmt.Errorf("error generating Go type for schema in parameter %s: %s",
						bodyName, err)
				}
				types = append(types, TypeDefinition{
					JsonTypeName: bodyName,
					TypeName:     ToCamelCase(bodyName),
					TypeDef:      goType,
				})
			}
		}
	}

	typesOut, err := GenerateTypes(t, "request-bodies.tmpl", types)
	if err != nil {
		return "", fmt.Errorf("error generating type definitions: %s", err)
	}

	return typesOut, nil
}

// Helper function to pass a bunch of types to the template engine, and buffer
// its output into a string.
func GenerateTypes(t *template.Template, templateName string, types []TypeDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	context := struct {
		Types []TypeDefinition
	}{
		Types: types,
	}

	err := t.ExecuteTemplate(w, templateName, context)
	if err != nil {
		return "", fmt.Errorf("error generating types: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for types: %s", err)
	}
	return buf.String(), nil
}

// Generate our import statements and package definition.
func GenerateImports(t *template.Template, imports []string, packageName string) (string, error) {
	if len(imports) == 0 {
		return "", nil
	}

	sort.Strings(imports)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	context := struct {
		Imports     []string
		PackageName string
	}{
		Imports:     imports,
		PackageName: packageName,
	}
	err := t.ExecuteTemplate(w, "imports.tmpl", context)
	if err != nil {
		return "", fmt.Errorf("error generating imports: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for imports: %s", err)
	}
	return buf.String(), nil
}
