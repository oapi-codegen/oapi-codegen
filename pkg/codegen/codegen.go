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
	"github.com/pkg/errors"

	"github.com/deepmap/oapi-codegen/pkg/codegen/templates"
)

// Options defines the optional code to generate.
type Options struct {
	GenerateServer              bool // GenerateServer specifies whether to generate server boilerplate
	GenerateClient              bool // GenerateClient specifies whether to generate client boilerplate
	GenerateClientWithResponses bool // GenerateClientWithResponses specifies whether to generate client boilerplate including responses (overrides GenerateClient)
	GenerateTypes               bool // GenerateTypes specifies whether to generate type definitions
	EmbedSpec                   bool // Whether to embed the swagger spec in the generated code
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
		return "", errors.Wrap(err, "error parsing oapi-codegen templates")
	}

	ops, err := OperationDefinitions(swagger)
	if err != nil {
		return "", errors.Wrap(err, "error creating operation definitions")
	}

	var typeDefinitions string
	if opts.GenerateTypes {
		typeDefinitions, err = GenerateTypeDefinitions(t, swagger, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating type definitions")
		}
	}

	var serverOut string
	if opts.GenerateServer {
		serverOut, err = GenerateServer(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go handlers for Paths")
		}
	}

	var clientOut string
	if opts.GenerateClientWithResponses {
		clientOut, err = GenerateClientWithResponses(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating client")
		}
	} else if opts.GenerateClient {
		clientOut, err = GenerateClient(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating client")
		}
	}

	var inlinedSpec string
	if opts.EmbedSpec {
		inlinedSpec, err = GenerateInlinedSpec(t, swagger)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go handlers for Paths")
		}
	}

	// Imports needed for the generated code to compile
	var imports []string

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// Based on module prefixes, figure out which optional imports are required.
	// TODO: this is error prone, use tighter matches
	for _, str := range []string{typeDefinitions, serverOut, clientOut, inlinedSpec} {
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
		if strings.Contains(str, "ioutil.") {
			imports = append(imports, "io/ioutil")
		}
		if strings.Contains(str, "url.") {
			imports = append(imports, "net/url")
		}
		if strings.Contains(str, "context.") {
			imports = append(imports, "context")
		}
		if strings.Contains(str, "runtime.") {
			imports = append(imports, "github.com/deepmap/oapi-codegen/pkg/runtime")
		}
		if strings.Contains(str, "bytes.") {
			imports = append(imports, "bytes")
		}
		if strings.Contains(str, "gzip.") {
			imports = append(imports, "compress/gzip")
		}
		if strings.Contains(str, "base64.") {
			imports = append(imports, "encoding/base64")
		}
		if strings.Contains(str, "openapi3.") {
			imports = append(imports, "github.com/getkin/kin-openapi/openapi3")
		}
		if strings.Contains(str, "strings.") {
			imports = append(imports, "strings")
		}
		if strings.Contains(str, "fmt.") {
			imports = append(imports, "fmt")
		}
		if strings.Contains(str, "yaml.") {
			imports = append(imports, "gopkg.in/yaml.v2")
		}
		if strings.Contains(str, "xml.") {
			imports = append(imports, "encoding/xml")
		}
	}

	importsOut, err := GenerateImports(t, imports, packageName)
	if err != nil {
		return "", errors.Wrap(err, "error generating imports")
	}

	_, err = w.WriteString(importsOut)
	if err != nil {
		return "", errors.Wrap(err, "error writing imports")
	}

	_, err = w.WriteString(typeDefinitions)
	if err != nil {
		return "", errors.Wrap(err, "error writing type definitions")

	}

	if opts.GenerateClient || opts.GenerateClientWithResponses {
		_, err = w.WriteString(clientOut)
		if err != nil {
			return "", errors.Wrap(err, "error writing client")
		}
	}

	if opts.GenerateServer {
		_, err = w.WriteString(serverOut)
		if err != nil {
			return "", errors.Wrap(err, "error writing server path handlers")
		}
	}

	if opts.EmbedSpec {
		_, err = w.WriteString(inlinedSpec)
		if err != nil {
			return "", errors.Wrap(err, "error writing inlined spec")
		}
	}

	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer")
	}

	goCode := buf.String()

	// The generation code produces unindented horrors. Use the Go formatter
	// to make it all pretty.
	outBytes, err := format.Source([]byte(goCode))
	if err != nil {
		fmt.Println(goCode)
		return "", errors.Wrap(err, "error formatting Go code")
	}
	return string(outBytes), nil
}

func GenerateTypeDefinitions(t *template.Template, swagger *openapi3.Swagger, ops []OperationDefinition) (string, error) {
	schemasOut, err := GenerateTypesForSchemas(t, swagger.Components.Schemas)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component schemas")
	}

	paramsOut, err := GenerateTypesForParameters(t, swagger.Components.Parameters)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component parameters")
	}

	responsesOut, err := GenerateTypesForResponses(t, swagger.Components.Responses)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component responses")
	}

	bodiesOut, err := GenerateTypesForRequestBodies(t, swagger.Components.RequestBodies)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component request bodies")
	}

	paramsTypesOut, err := GenerateTypesForParams(t, ops)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component request bodies")
	}

	typeDefinitions := strings.Join([]string{schemasOut, paramsOut, responsesOut, bodiesOut, paramsTypesOut}, "")
	return typeDefinitions, nil
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
			return "", errors.Wrap(err, fmt.Sprintf("error converting Schema %s to Go type", schemaName))
		}

		types = append(types, TypeDefinition{
			JsonTypeName: schemaName,
			TypeName:     ToCamelCase(schemaName),
			TypeDef:      typeDef,
		})
	}

	typesOut, err := GenerateTypes(t, "schemas.tmpl", types)
	if err != nil {
		return "", errors.Wrap(err, "error generating type definitions")
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
				return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", paramOrRef.Ref, paramName))
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
				return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in parameter %s", paramName))
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
		return "", errors.Wrap(err, "error generating type definitions")
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
				return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", responseOrRef.Ref, responseName))
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
					return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in parameter %s", responseName))
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
		return "", errors.Wrap(err, "error generating response type definitions")
	}

	return typesOut, nil
}

// Generates type definitions for any custom types defined in the
// components/requestBodies section of the Swagger spec.
func GenerateTypesForRequestBodies(t *template.Template, bodies map[string]*openapi3.RequestBodyRef) (string, error) {
	types := make([]TypeDefinition, 0)
	for _, bodyName := range SortedRequestBodyKeys(bodies) {
		bodyOrRef := bodies[bodyName]
		if bodyOrRef.Ref != "" {
			ref := bodyOrRef.Ref
			goType, err := RefPathToGoType(ref)
			if err != nil {
				return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in request body %s", ref, bodyName))
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
					return "", errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in parameter %s", bodyName))
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
		return "", errors.Wrap(err, "error generating types")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for types")
	}
	return buf.String(), nil
}

// Generate our import statements and package definition.
func GenerateImports(t *template.Template, imports []string, packageName string) (string, error) {
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
		return "", errors.Wrap(err, "error generating imports")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for imports")
	}
	return buf.String(), nil
}
