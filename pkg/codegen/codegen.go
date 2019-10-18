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
	GenerateChiServer  bool // GenerateChiServer specifies whether to generate chi server boilerplate
	GenerateEchoServer bool // GenerateEchoServer specifies whether to generate echo server boilerplate
	GenerateClient     bool // GenerateClient specifies whether to generate client boilerplate
	GenerateTypes      bool // GenerateTypes specifies whether to generate type definitions
	EmbedSpec          bool // Whether to embed the swagger spec in the generated code
	SkipFmt            bool // Whether to skip go fmt on the generated code
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

	var echoServerOut string
	if opts.GenerateEchoServer {
		echoServerOut, err = GenerateEchoServer(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go handlers for Paths")
		}
	}

	var chiServerOut string
	if opts.GenerateChiServer {
		chiServerOut, err = GenerateChiServer(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go handlers for Paths")
		}
	}

	var clientOut string
	if opts.GenerateClient {
		clientOut, err = GenerateClient(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating client")
		}
	}

	var clientWithResponsesOut string
	if opts.GenerateClient {
		clientWithResponsesOut, err = GenerateClientWithResponses(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating client with responses")
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
	for _, str := range []string{typeDefinitions, chiServerOut, echoServerOut, clientOut, clientWithResponsesOut, inlinedSpec} {
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
		if strings.Contains(str, "errors.") {
			imports = append(imports, "github.com/pkg/errors")
		}
		if strings.Contains(str, "chi.") {
			imports = append(imports, "github.com/go-chi/chi")
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

	if opts.GenerateClient {
		_, err = w.WriteString(clientOut)
		if err != nil {
			return "", errors.Wrap(err, "error writing client")
		}
		_, err = w.WriteString(clientWithResponsesOut)
		if err != nil {
			return "", errors.Wrap(err, "error writing client")
		}
	}

	if opts.GenerateEchoServer {
		_, err = w.WriteString(echoServerOut)
		if err != nil {
			return "", errors.Wrap(err, "error writing server path handlers")
		}
	}

	if opts.GenerateChiServer {
		_, err = w.WriteString(chiServerOut)
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
	if opts.SkipFmt {
		return goCode, nil
	}
	outBytes, err := format.Source([]byte(goCode))
	if err != nil {
		fmt.Println(goCode)
		return "", errors.Wrap(err, "error formatting Go code")
	}
	return string(outBytes), nil
}

func GenerateTypeDefinitions(t *template.Template, swagger *openapi3.Swagger, ops []OperationDefinition) (string, error) {
	schemaTypes, err := GenerateTypesForSchemas(t, swagger.Components.Schemas)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component schemas")
	}

	paramTypes, err := GenerateTypesForParameters(t, swagger.Components.Parameters)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component parameters")
	}
	allTypes := append(schemaTypes, paramTypes...)

	responseTypes, err := GenerateTypesForResponses(t, swagger.Components.Responses)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component responses")
	}
	allTypes = append(allTypes, responseTypes...)

	bodyTypes, err := GenerateTypesForRequestBodies(t, swagger.Components.RequestBodies)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component request bodies")
	}
	allTypes = append(allTypes, bodyTypes...)

	paramTypesOut, err := GenerateTypesForOperations(t, ops)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for operation parameters")
	}

	typesOut, err := GenerateTypes(t, allTypes)
	if err != nil {
		return "", errors.Wrap(err, "error generating code for type definitions")
	}

	allOfBoilerplate, err := GenerateAdditionalPropertyBoilerplate(t, allTypes)
	if err != nil {
		return "", errors.Wrap(err, "error generating allOf boilerplate")
	}

	typeDefinitions := strings.Join([]string{typesOut, paramTypesOut, allOfBoilerplate}, "")
	return typeDefinitions, nil
}

// Generates type definitions for any custom types defined in the
// components/schemas section of the Swagger spec.
func GenerateTypesForSchemas(t *template.Template, schemas map[string]*openapi3.SchemaRef) ([]TypeDefinition, error) {
	types := make([]TypeDefinition, 0)
	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range SortedSchemaKeys(schemas) {
		schemaRef := schemas[schemaName]

		goSchema, err := GenerateGoSchema(schemaRef, []string{schemaName})
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error converting Schema %s to Go type", schemaName))
		}

		types = append(types, TypeDefinition{
			JsonName: schemaName,
			TypeName: SchemaNameToTypeName(schemaName),
			Schema:   goSchema,
		})

		types = append(types, goSchema.GetAdditionalTypeDefs()...)
	}
	return types, nil
}

// Generates type definitions for any custom types defined in the
// components/parameters section of the Swagger spec.
func GenerateTypesForParameters(t *template.Template, params map[string]*openapi3.ParameterRef) ([]TypeDefinition, error) {
	var types []TypeDefinition
	for _, paramName := range SortedParameterKeys(params) {
		paramOrRef := params[paramName]

		goType, err := paramToGoType(paramOrRef.Value, nil)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in parameter %s", paramName))
		}

		typeDef := TypeDefinition{
			JsonName: paramName,
			Schema:   goType,
			TypeName: SchemaNameToTypeName(paramName),
		}

		if paramOrRef.Ref != "" {
			// Generate a reference type for referenced parameters
			refType, err := RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", paramOrRef.Ref, paramName))
			}
			typeDef.TypeName = SchemaNameToTypeName(refType)
		}

		types = append(types, typeDef)
	}
	return types, nil
}

// Generates type definitions for any custom types defined in the
// components/responses section of the Swagger spec.
func GenerateTypesForResponses(t *template.Template, responses openapi3.Responses) ([]TypeDefinition, error) {
	var types []TypeDefinition

	for _, responseName := range SortedResponsesKeys(responses) {
		responseOrRef := responses[responseName]

		// We have to generate the response object. We're only going to
		// handle application/json media types here. Other responses should
		// simply be specified as strings or byte arrays.
		response := responseOrRef.Value
		jsonResponse, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(jsonResponse.Schema, []string{responseName})
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in response %s", responseName))
			}

			typeDef := TypeDefinition{
				JsonName: responseName,
				Schema:   goType,
				TypeName: SchemaNameToTypeName(responseName),
			}

			if responseOrRef.Ref != "" {
				// Generate a reference type for referenced parameters
				refType, err := RefPathToGoType(responseOrRef.Ref)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", responseOrRef.Ref, responseName))
				}
				typeDef.TypeName = SchemaNameToTypeName(refType)
			}
			types = append(types, typeDef)
		}
	}
	return types, nil
}

// Generates type definitions for any custom types defined in the
// components/requestBodies section of the Swagger spec.
func GenerateTypesForRequestBodies(t *template.Template, bodies map[string]*openapi3.RequestBodyRef) ([]TypeDefinition, error) {
	var types []TypeDefinition

	for _, bodyName := range SortedRequestBodyKeys(bodies) {
		bodyOrRef := bodies[bodyName]

		// As for responses, we will only generate Go code for JSON bodies,
		// the other body formats are up to the user.
		response := bodyOrRef.Value
		jsonBody, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(jsonBody.Schema, []string{bodyName})
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in body %s", bodyName))
			}

			typeDef := TypeDefinition{
				JsonName: bodyName,
				Schema:   goType,
				TypeName: SchemaNameToTypeName(bodyName),
			}

			if bodyOrRef.Ref != "" {
				// Generate a reference type for referenced bodies
				refType, err := RefPathToGoType(bodyOrRef.Ref)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in body %s", bodyOrRef.Ref, bodyName))
				}
				typeDef.TypeName = SchemaNameToTypeName(refType)
			}
			types = append(types, typeDef)
		}
	}
	return types, nil
}

// Helper function to pass a bunch of types to the template engine, and buffer
// its output into a string.
func GenerateTypes(t *template.Template, types []TypeDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	context := struct {
		Types []TypeDefinition
	}{
		Types: types,
	}

	err := t.ExecuteTemplate(w, "typedef.tmpl", context)
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

// Generate all the glue code which provides the API for interacting with
// additional properties and JSON-ification
func GenerateAdditionalPropertyBoilerplate(t *template.Template, typeDefs []TypeDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	var filteredTypes []TypeDefinition
	for _, t := range typeDefs {
		if t.Schema.HasAdditionalProperties {
			filteredTypes = append(filteredTypes, t)
		}
	}

	context := struct {
		Types []TypeDefinition
	}{
		Types: filteredTypes,
	}

	err := t.ExecuteTemplate(w, "additional-properties.tmpl", context)
	if err != nil {
		return "", errors.Wrap(err, "error generating additional properties code")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for additional properties")
	}
	return buf.String(), nil
}
