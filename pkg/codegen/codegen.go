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
	"runtime/debug"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/imports"

	"github.com/deepmap/oapi-codegen/pkg/codegen/templates"
)

// Options defines the optional code to generate.
type Options struct {
	GenerateChiServer  bool              // GenerateChiServer specifies whether to generate chi server boilerplate
	GenerateEchoServer bool              // GenerateEchoServer specifies whether to generate echo server boilerplate
	GenerateClient     bool              // GenerateClient specifies whether to generate client boilerplate
	GenerateTypes      bool              // GenerateTypes specifies whether to generate type definitions
	EmbedSpec          bool              // Whether to embed the swagger spec in the generated code
	SkipFmt            bool              // Whether to skip go imports on the generated code
	SkipPrune          bool              // Whether to skip pruning unused components on the generated code
	AliasTypes         bool              // Whether to alias types if possible
	IncludeTags        []string          // Only include operations that have one of these tags. Ignored when empty.
	ExcludeTags        []string          // Exclude operations that have one of these tags. Ignored when empty.
	UserTemplates      map[string]string // Override built-in templates from user-provided files
	ImportMapping      map[string]string // ImportMapping specifies the golang package path for each external reference
	ExcludeSchemas     []string          // Exclude from generation schemas with given names. Ignored when empty.
}

// goImport represents a go package to be imported in the generated code
type goImport struct {
	Name string // package name
	Path string // package path
}

// String returns a go import statement
func (gi goImport) String() string {
	if gi.Name != "" {
		return fmt.Sprintf("%s %q", gi.Name, gi.Path)
	}
	return fmt.Sprintf("%q", gi.Path)
}

// importMap maps external OpenAPI specifications files/urls to external go packages
type importMap map[string]goImport

// GoImports returns a slice of go import statements
func (im importMap) GoImports() []string {
	goImports := make([]string, 0, len(im))
	for _, v := range im {
		goImports = append(goImports, v.String())
	}
	return goImports
}

var importMapping importMap

func constructImportMapping(input map[string]string) importMap {
	var (
		pathToName = map[string]string{}
		result     = importMap{}
	)

	{
		var packagePaths []string
		for _, packageName := range input {
			packagePaths = append(packagePaths, packageName)
		}
		sort.Strings(packagePaths)

		for _, packagePath := range packagePaths {
			if _, ok := pathToName[packagePath]; !ok {
				pathToName[packagePath] = fmt.Sprintf("externalRef%d", len(pathToName))
			}
		}
	}
	for specPath, packagePath := range input {
		result[specPath] = goImport{Name: pathToName[packagePath], Path: packagePath}
	}
	return result
}

// Uses the Go templating engine to generate all of our server wrappers from
// the descriptions we've built up above from the schema objects.
// opts defines
func Generate(swagger *openapi3.T, packageName string, opts Options) (string, error) {
	importMapping = constructImportMapping(opts.ImportMapping)

	filterOperationsByTag(swagger, opts)
	if !opts.SkipPrune {
		pruneUnusedComponents(swagger)
	}

	// This creates the golang templates text package
	TemplateFunctions["opts"] = func() Options { return opts }
	t := template.New("oapi-codegen").Funcs(TemplateFunctions)
	// This parses all of our own template files into the template object
	// above
	t, err := templates.Parse(t)
	if err != nil {
		return "", fmt.Errorf("error parsing oapi-codegen templates: %w", err)
	}

	// Override built-in templates with user-provided versions
	for _, tpl := range t.Templates() {
		if _, ok := opts.UserTemplates[tpl.Name()]; ok {
			utpl := t.New(tpl.Name())
			if _, err := utpl.Parse(opts.UserTemplates[tpl.Name()]); err != nil {
				return "", fmt.Errorf("error parsing user-provided template %q: %w", tpl.Name(), err)
			}
		}
	}

	ops, err := OperationDefinitions(swagger)
	if err != nil {
		return "", fmt.Errorf("error creating operation definitions: %w", err)
	}

	var typeDefinitions, constantDefinitions string
	if opts.GenerateTypes {
		typeDefinitions, err = GenerateTypeDefinitions(t, swagger, ops, opts.ExcludeSchemas)
		if err != nil {
			return "", fmt.Errorf("error generating type definitions: %w", err)
		}

		constantDefinitions, err = GenerateConstants(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating constants: %w", err)
		}

	}

	var echoServerOut string
	if opts.GenerateEchoServer {
		echoServerOut, err = GenerateEchoServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var chiServerOut string
	if opts.GenerateChiServer {
		chiServerOut, err = GenerateChiServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var clientOut string
	if opts.GenerateClient {
		clientOut, err = GenerateClient(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating client: %w", err)
		}
	}

	var clientWithResponsesOut string
	if opts.GenerateClient {
		clientWithResponsesOut, err = GenerateClientWithResponses(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating client with responses: %w", err)
		}
	}

	var inlinedSpec string
	if opts.EmbedSpec {
		inlinedSpec, err = GenerateInlinedSpec(t, importMapping, swagger)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	externalImports := importMapping.GoImports()
	importsOut, err := GenerateImports(t, externalImports, packageName)
	if err != nil {
		return "", fmt.Errorf("error generating imports: %w", err)
	}

	_, err = w.WriteString(importsOut)
	if err != nil {
		return "", fmt.Errorf("error writing imports: %w", err)
	}

	_, err = w.WriteString(constantDefinitions)
	if err != nil {
		return "", fmt.Errorf("error writing constants: %w", err)
	}

	_, err = w.WriteString(typeDefinitions)
	if err != nil {
		return "", fmt.Errorf("error writing type definitions: %w", err)

	}

	if opts.GenerateClient {
		_, err = w.WriteString(clientOut)
		if err != nil {
			return "", fmt.Errorf("error writing client: %w", err)
		}
		_, err = w.WriteString(clientWithResponsesOut)
		if err != nil {
			return "", fmt.Errorf("error writing client: %w", err)
		}
	}

	if opts.GenerateEchoServer {
		_, err = w.WriteString(echoServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.GenerateChiServer {
		_, err = w.WriteString(chiServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.EmbedSpec {
		_, err = w.WriteString(inlinedSpec)
		if err != nil {
			return "", fmt.Errorf("error writing inlined spec: %w", err)
		}
	}

	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer: %w", err)
	}

	// remove any byte-order-marks which break Go-Code
	goCode := SanitizeCode(buf.String())

	// The generation code produces unindented horrors. Use the Go Imports
	// to make it all pretty.
	if opts.SkipFmt {
		return goCode, nil
	}

	outBytes, err := imports.Process(packageName+".go", []byte(goCode), nil)
	if err != nil {
		fmt.Println(goCode)
		return "", fmt.Errorf("error formatting Go code: %w", err)
	}
	return string(outBytes), nil
}

func GenerateTypeDefinitions(t *template.Template, swagger *openapi3.T, ops []OperationDefinition, excludeSchemas []string) (string, error) {
	schemaTypes, err := GenerateTypesForSchemas(t, swagger.Components.Schemas, excludeSchemas)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component schemas: %w", err)
	}

	paramTypes, err := GenerateTypesForParameters(t, swagger.Components.Parameters)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component parameters: %w", err)
	}
	allTypes := append(schemaTypes, paramTypes...)

	responseTypes, err := GenerateTypesForResponses(t, swagger.Components.Responses)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component responses: %w", err)
	}
	allTypes = append(allTypes, responseTypes...)

	bodyTypes, err := GenerateTypesForRequestBodies(t, swagger.Components.RequestBodies)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component request bodies: %w", err)
	}
	allTypes = append(allTypes, bodyTypes...)

	paramTypesOut, err := GenerateTypesForOperations(t, ops)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component request bodies: %w", err)
	}

	enumsOut, err := GenerateEnums(t, allTypes)
	if err != nil {
		return "", fmt.Errorf("error generating code for type enums: %w", err)
	}

	typesOut, err := GenerateTypes(t, allTypes)
	if err != nil {
		return "", fmt.Errorf("error generating code for type definitions: %w", err)
	}

	allOfBoilerplate, err := GenerateAdditionalPropertyBoilerplate(t, allTypes)
	if err != nil {
		return "", fmt.Errorf("error generating allOf boilerplate: %w", err)
	}

	typeDefinitions := strings.Join([]string{enumsOut, typesOut, paramTypesOut, allOfBoilerplate}, "")
	return typeDefinitions, nil
}

// Generates operation ids, context keys, paths, etc. to be exported as constants
func GenerateConstants(t *template.Template, ops []OperationDefinition) (string, error) {
	constants := Constants{
		SecuritySchemeProviderNames: []string{},
	}

	providerNameMap := map[string]struct{}{}
	for _, op := range ops {
		for _, def := range op.SecurityDefinitions {
			providerName := SanitizeGoIdentity(def.ProviderName)
			providerNameMap[providerName] = struct{}{}
		}
	}

	var providerNames []string
	for providerName := range providerNameMap {
		providerNames = append(providerNames, providerName)
	}

	sort.Strings(providerNames)

	constants.SecuritySchemeProviderNames = append(constants.SecuritySchemeProviderNames, providerNames...)

	return GenerateTemplates([]string{"constants.tmpl"}, t, constants)
}

// Generates type definitions for any custom types defined in the
// components/schemas section of the Swagger spec.
func GenerateTypesForSchemas(t *template.Template, schemas map[string]*openapi3.SchemaRef, excludeSchemas []string) ([]TypeDefinition, error) {
	var excludeSchemasMap = make(map[string]bool)
	for _, schema := range excludeSchemas {
		excludeSchemasMap[schema] = true
	}
	types := make([]TypeDefinition, 0)
	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range SortedSchemaKeys(schemas) {
		if _, ok := excludeSchemasMap[schemaName]; ok {
			continue
		}
		schemaRef := schemas[schemaName]

		goSchema, err := GenerateGoSchema(schemaRef, []string{schemaName})
		if err != nil {
			return nil, fmt.Errorf("error converting Schema %s to Go type: %w", schemaName, err)
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
			return nil, fmt.Errorf("error generating Go type for schema in parameter %s: %w", paramName, err)
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
				return nil, fmt.Errorf("error generating Go type for (%s) in parameter %s: %w", paramOrRef.Ref, paramName, err)
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
				return nil, fmt.Errorf("error generating Go type for schema in response %s: %w", responseName, err)
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
					return nil, fmt.Errorf("error generating Go type for (%s) in parameter %s: %w", responseOrRef.Ref, responseName, err)
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
				return nil, fmt.Errorf("error generating Go type for schema in body %s: %w", bodyName, err)
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
					return nil, fmt.Errorf("error generating Go type for (%s) in body %s: %w", bodyOrRef.Ref, bodyName, err)
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
	m := map[string]bool{}
	ts := []TypeDefinition{}

	for _, t := range types {
		if found := m[t.TypeName]; found {
			continue
		}

		m[t.TypeName] = true

		ts = append(ts, t)
	}

	context := struct {
		Types []TypeDefinition
	}{
		Types: ts,
	}

	return GenerateTemplates([]string{"typedef.tmpl"}, t, context)
}

func GenerateEnums(t *template.Template, types []TypeDefinition) (string, error) {
	c := Constants{
		EnumDefinitions: []EnumDefinition{},
	}

	m := map[string]bool{}

	for _, tp := range types {
		if found := m[tp.TypeName]; found {
			continue
		}

		m[tp.TypeName] = true

		if len(tp.Schema.EnumValues) > 0 {
			wrapper := ""
			if tp.Schema.GoType == "string" {
				wrapper = `"`
			}
			c.EnumDefinitions = append(c.EnumDefinitions, EnumDefinition{
				Schema:       tp.Schema,
				TypeName:     tp.TypeName,
				ValueWrapper: wrapper,
			})
		}
	}

	return GenerateTemplates([]string{"constants.tmpl"}, t, c)

}

// Generate our import statements and package definition.
func GenerateImports(t *template.Template, externalImports []string, packageName string) (string, error) {
	// Read build version for incorporating into generated files
	var modulePath string
	var moduleVersion string
	if bi, ok := debug.ReadBuildInfo(); ok {
		modulePath = bi.Main.Path
		moduleVersion = bi.Main.Version
	} else {
		// Unit tests have ok=false, so we'll just use "unknown" for the
		// version if we can't read this.
		modulePath = "unknown module path"
		moduleVersion = "unknown version"
	}

	context := struct {
		ExternalImports []string
		PackageName     string
		ModuleName      string
		Version         string
	}{
		ExternalImports: externalImports,
		PackageName:     packageName,
		ModuleName:      modulePath,
		Version:         moduleVersion,
	}

	return GenerateTemplates([]string{"imports.tmpl"}, t, context)
}

// Generate all the glue code which provides the API for interacting with
// additional properties and JSON-ification
func GenerateAdditionalPropertyBoilerplate(t *template.Template, typeDefs []TypeDefinition) (string, error) {
	var filteredTypes []TypeDefinition

	m := map[string]bool{}

	for _, t := range typeDefs {
		if found := m[t.TypeName]; found {
			continue
		}

		m[t.TypeName] = true

		if t.Schema.HasAdditionalProperties {
			filteredTypes = append(filteredTypes, t)
		}
	}

	context := struct {
		Types []TypeDefinition
	}{
		Types: filteredTypes,
	}

	return GenerateTemplates([]string{"additional-properties.tmpl"}, t, context)

}

// SanitizeCode runs sanitizers across the generated Go code to ensure the
// generated code will be able to compile.
func SanitizeCode(goCode string) string {
	// remove any byte-order-marks which break Go-Code
	// See: https://groups.google.com/forum/#!topic/golang-nuts/OToNIPdfkks
	return strings.Replace(goCode, "\uFEFF", "", -1)
}
