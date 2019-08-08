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
	GenerateServer bool              // GenerateServer specifies whether to generate server boilerplate
	GenerateClient bool              // GenerateClient specifies whether to generate client boilerplate
	GenerateTypes  bool              // GenerateTypes specifies whether to generate type definitions
	EmbedSpec      bool              // Whether to embed the swagger spec in the generated code
	TypeImports    map[string]string // Additional imports added when using remote references
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

	ops, err := OperationDefinitions(swagger, opts.TypeImports)
	if err != nil {
		return "", errors.Wrap(err, "error creating operation definitions")
	}

	//if opts.TypeImports != nil {
	// TODO: we need to add package name to external resource type names.
	// this affects all defines schemas here
	//}

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
	allCode := strings.Join([]string{typeDefinitions, serverOut, clientOut, clientWithResponsesOut, inlinedSpec}, "\n")

	knownTypeImports := map[string]string{
		"time.Time": "time",
		"http.":     "net/http",
		"openapi3.": "github.com/getkin/kin-openapi/openapi3",
		"json.":     "encoding/json",
		"echo.":     "github.com/labstack/echo/v4",
		"io.":       "io",
		"ioutil.":   "io/ioutil",
		"url.":      "net/url",
		"context.":  "context",
		"runtime.":  "github.com/deepmap/oapi-codegen/pkg/runtime",
		"bytes.":    "bytes",
		"gzip.":     "compress/gzip",
		"base64.":   "encoding/base64",
		"strings.":  "strings",
		"fmt.":      "fmt",
		"yaml.":     "gopkg.in/yaml.v2",
		"xml.":      "encoding/xml",
		"errors.":   "github.com/pkg/errors",
	}

	for t, i := range knownTypeImports {
		if strings.Contains(allCode, t) {
			imports = append(imports, i)
		}
	}

	for t, i := range opts.TypeImports {
		if strings.Contains(allCode, t) {
			imports = append(imports, i)
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
	var typeMap map[string]string
	if len(ops) > 0 {
		typeMap = ops[0].TypeMap // they are all the same
	}

	schemaTypes, err := GenerateTypesForSchemas(t, swagger.Components.Schemas, typeMap)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component schemas")
	}

	paramTypes, err := GenerateTypesForParameters(t, swagger.Components.Parameters, typeMap)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component parameters")
	}
	allTypes := append(schemaTypes, paramTypes...)

	responseTypes, err := GenerateTypesForResponses(t, swagger.Components.Responses, typeMap)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component responses")
	}
	allTypes = append(allTypes, responseTypes...)

	bodyTypes, err := GenerateTypesForRequestBodies(t, swagger.Components.RequestBodies, typeMap)
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
func GenerateTypesForSchemas(t *template.Template, schemas map[string]*openapi3.SchemaRef, typeMap map[string]string) ([]TypeDefinition, error) {
	types := make([]TypeDefinition, 0)
	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range SortedSchemaKeys(schemas) {
		schemaRef := schemas[schemaName]

		goSchema, err := GenerateGoSchema(schemaRef, []string{schemaName}, typeMap)
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
func GenerateTypesForParameters(t *template.Template, params map[string]*openapi3.ParameterRef, typeMap map[string]string) ([]TypeDefinition, error) {
	var types []TypeDefinition
	for _, paramName := range SortedParameterKeys(params) {
		paramOrRef := params[paramName]

		goType, err := paramToGoType(paramOrRef.Value, nil, typeMap)
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
			refType, err := RefPathToGoType(paramOrRef.Ref, typeMap)
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
func GenerateTypesForResponses(t *template.Template, responses openapi3.Responses, typeMap map[string]string) ([]TypeDefinition, error) {
	var types []TypeDefinition

	for _, responseName := range SortedResponsesKeys(responses) {
		responseOrRef := responses[responseName]

		// We have to generate the response object. We're only going to
		// handle application/json media types here. Other responses should
		// simply be specified as strings or byte arrays.
		response := responseOrRef.Value
		jsonResponse, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(jsonResponse.Schema, []string{responseName}, typeMap)
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
				refType, err := RefPathToGoType(responseOrRef.Ref, typeMap)
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
func GenerateTypesForRequestBodies(t *template.Template, bodies map[string]*openapi3.RequestBodyRef, typeMap map[string]string) ([]TypeDefinition, error) {
	var types []TypeDefinition

	for _, bodyName := range SortedRequestBodyKeys(bodies) {
		bodyOrRef := bodies[bodyName]

		// As for responses, we will only generate Go code for JSON bodies,
		// the other body formats are up to the user.
		response := bodyOrRef.Value
		jsonBody, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(jsonBody.Schema, []string{bodyName}, typeMap)
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
				refType, err := RefPathToGoType(bodyOrRef.Ref, typeMap)
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
