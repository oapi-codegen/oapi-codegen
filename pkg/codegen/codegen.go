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
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"

	"github.com/weberr13/oapi-codegen/pkg/codegen/templates"
)

// Options defines the optional code to generate.
type Options struct {
	GenerateChiServer  bool                      // GenerateChiServer specifies whether to generate chi server boilerplate
	GenerateEchoServer bool                      // GenerateEchoServer specifies whether to generate echo server boilerplate
	GenerateClient     bool                      // GenerateClient specifies whether to generate client boilerplate
	GenerateTypes      bool                      // GenerateTypes specifies whether to generate type definitions
	EmbedSpec          bool                      // Whether to embed the swagger spec in the generated code
	SkipFmt            bool                      // Whether to skip go fmt on the generated code
	IncludeTags        []string                  // Only include operations that have one of these tags. Ignored when empty.
	ExcludeTags        []string                  // Exclude operations that have one of these tags. Ignored when empty.
	ClearRefsSpec      bool                      // Whether to resolve all references in the included swagger spec
	ImportedTypes      map[string]TypeImportSpec // Additional imports added when using remote references
}

type goImport struct {
	lookFor     string
	alias       string
	packageName string
}

func (i goImport) String() string {
	if i.alias != "" {
		return fmt.Sprintf("%s %q", i.alias, i.packageName)
	}
	return fmt.Sprintf("%q", i.packageName)
}

type goImports []goImport

var (
	allGoImports = goImports{
		{lookFor: "base64\\.", packageName: "encoding/base64"},
		{lookFor: "bytes\\.", packageName: "bytes"},
		{lookFor: "chi\\.", packageName: "github.com/go-chi/chi"},
		{lookFor: "context\\.", packageName: "context"},
		{lookFor: "echo\\.", packageName: "github.com/labstack/echo/v4"},
		{lookFor: "errors\\.", packageName: "github.com/pkg/errors"},
		{lookFor: "fmt\\.", packageName: "fmt"},
		{lookFor: "gzip\\.", packageName: "compress/gzip"},
		{lookFor: "http\\.", packageName: "net/http"},
		{lookFor: "io\\.", packageName: "io"},
		{lookFor: "ioutil\\.", packageName: "io/ioutil"},
		{lookFor: "json\\.", packageName: "encoding/json"},
		{lookFor: "openapi3\\.", packageName: "github.com/getkin/kin-openapi/openapi3"},
		{lookFor: "openapi_types\\.", alias: "openapi_types", packageName: "github.com/weberr13/oapi-codegen/pkg/types"},
		{lookFor: "path\\.", packageName: "path"},
		{lookFor: "runtime\\.", packageName: "github.com/weberr13/oapi-codegen/pkg/runtime"},
		{lookFor: "strings\\.", packageName: "strings"},
		{lookFor: "time\\.Duration", packageName: "time"},
		{lookFor: "time\\.Time", packageName: "time"},
		{lookFor: "url\\.", packageName: "net/url"},
		{lookFor: "xml\\.", packageName: "encoding/xml"},
		{lookFor: "yaml\\.", packageName: "gopkg.in/yaml.v2"},
		{lookFor: "decode\\.", packageName: "github.com/weberr13/go-decode/decode"},
	}
)

// Uses the Go templating engine to generate all of our server wrappers from

type genCtx struct {
	PackageName   string
	TypeDefs      []TypeDefinition
	OpDefs        []OperationDefinition
	GoSchemaMap   map[string]Schema
	ImportedTypes map[string]TypeImportSpec
	Swagger       *openapi3.Swagger
	HasDecorators bool
}

func newGenCtx(swagger *openapi3.Swagger, packageName string, opts Options) *genCtx {
	return &genCtx{Swagger: swagger, PackageName: packageName,
		TypeDefs: []TypeDefinition{}, GoSchemaMap: map[string]Schema{},
		ImportedTypes: opts.ImportedTypes,
	}
}

func (ctx *genCtx) getAllOpTypes() []TypeDefinition {
	var td []TypeDefinition
	for _, op := range ctx.OpDefs {
		td = append(td, op.TypeDefinitions...)
	}
	return td
}

// Generate uses the Go templating engine to generate all of our server wrappers from
// the descriptions we've built up above from the schema objects.
// opts defines
func Generate(swagger *openapi3.Swagger, packageName string, opts Options) (string, error) {
	filterOperationsByTag(swagger, opts)

	// This creates the golang templates text package
	t := template.New("oapi-codegen").Funcs(TemplateFunctions)
	// This parses all of our own template files into the template object
	// above
	t, err := templates.Parse(t)
	if err != nil {
		return "", errors.Wrap(err, "error parsing oapi-codegen templates")
	}

	ctx := newGenCtx(swagger, packageName, opts)

	// only generate operation typedefs if not producing resolved spec
	if !opts.ClearRefsSpec {
		err = OperationDefinitions(ctx, swagger)
		if err != nil {
			return "", errors.Wrap(err, "error creating operation definitions")
		}
	}

	var typeDefinitions string
	if opts.GenerateTypes {
		typeDefinitions, err = GenerateTypeDefinitions(ctx, t, swagger)
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
		clientOut, err = GenerateClient(ctx, t)
		if err != nil {
			return "", errors.Wrap(err, "error generating client")
		}
	}

	var clientWithResponsesOut string
	if opts.GenerateClient {
		clientWithResponsesOut, err = GenerateClientWithResponses(ctx, t)
		if err != nil {
			return "", errors.Wrap(err, "error generating client with responses")
		}
	}

	var inlinedSpec string
	if opts.EmbedSpec {
		inlinedSpec, err = GenerateInlinedSpec(ctx, t, swagger)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go handlers for Paths")
		}
	}

	// Imports needed for the generated code to compile
	var imports []TypeImportSpec

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// TODO: this is error prone, use tighter matches
	allCode := strings.Join([]string{typeDefinitions, echoServerOut, chiServerOut, clientOut, clientWithResponsesOut, inlinedSpec}, "\n")
	for t, i := range ctx.ImportedTypes {
		if strings.Contains(allCode, t) {
			allGoImports = append(allGoImports, goImport{lookFor: `\` + i.Name, packageName: i.PackageName})
		}
	}
	// Based on module prefixes, figure out which optional imports are required.
	for _, str := range []string{typeDefinitions, chiServerOut, echoServerOut, clientOut, clientWithResponsesOut, inlinedSpec} {
		for _, goImport := range allGoImports {
			match, err := regexp.MatchString(fmt.Sprintf("[^a-zA-Z0-9_]%s", goImport.lookFor), str)
			if err != nil {
				return "", errors.Wrap(err, "error figuring out imports")
			}
			if match {
				imports = append(imports, goImport.String())
			}
		}
	}

	importsOut, err := GenerateImports(ctx, t, imports)
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

	// remove any byte-order-marks which break Go-Code
	goCode := SanitizeCode(buf.String())

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

func GenerateTypeDefinitions(ctx *genCtx, t *template.Template, swagger *openapi3.Swagger) (string, error) {
	var err error

	err = GenerateTypesForSchemas(ctx, t, swagger.Components.Schemas)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component schemas")
	}

	err = GenerateTypesForParameters(ctx, t, swagger.Components.Parameters)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component parameters")
	}

	err = GenerateTypesForResponses(ctx, t, swagger.Components.Responses)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component responses")
	}

	err = GenerateTypesForRequestBodies(ctx, t, swagger.Components.RequestBodies)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for component request bodies")
	}

	paramTypesOut, err := GenerateTypesForOperations(ctx, t)
	if err != nil {
		return "", errors.Wrap(err, "error generating Go types for operation parameters")
	}

	typesOut, err := GenerateTypes(ctx, t)
	if err != nil {
		return "", errors.Wrap(err, "error generating code for type definitions")
	}
	allOfBoilerplate, err := GenerateAdditionalPropertyBoilerplate(ctx, t, ctx.TypeDefs)
	if err != nil {
		return "", errors.Wrap(err, "error generating allOf boilerplate")
	}

	typeDefinitions := strings.Join([]string{typesOut, paramTypesOut, allOfBoilerplate}, "")
	return typeDefinitions, nil
}

// Generates type definitions for any custom types defined in the
// components/schemas section of the Swagger spec.
func GenerateTypesForSchemas(ctx *genCtx, t *template.Template, schemas map[string]*openapi3.SchemaRef) error {

	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range SortedSchemaKeys(schemas) {
		schemaRef := schemas[schemaName]

		goSchema, err := GenerateGoSchema(ctx, schemaRef, []string{schemaName})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error converting Schema %s to Go type", schemaName))
		}

		ctx.TypeDefs = append(ctx.TypeDefs, TypeDefinition{
			JsonName: schemaName,
			TypeName: SchemaNameToTypeName(schemaName),
			Schema:   goSchema,
		})

		ctx.TypeDefs = append(ctx.TypeDefs, goSchema.GetAdditionalTypeDefs()...)
	}
	return nil
}

// Generates type definitions for any custom types defined in the
// components/parameters section of the Swagger spec.
func GenerateTypesForParameters(ctx *genCtx, t *template.Template, params map[string]*openapi3.ParameterRef) error {
	for _, paramName := range SortedParameterKeys(params) {
		paramOrRef := params[paramName]

		goType, err := paramToGoType(ctx, paramOrRef.Value, nil)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in parameter %s", paramName))
		}

		typeDef := TypeDefinition{
			JsonName: paramName,
			Schema:   goType,
			TypeName: SchemaNameToTypeName(paramName),
		}

		if paramOrRef.Ref != "" {
			// Generate a reference type for referenced parameters
			refType, err := RefPathToGoType(paramOrRef.Ref, ctx.ImportedTypes)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", paramOrRef.Ref, paramName))
			}
			typeDef.TypeName = SchemaNameToTypeName(refType)
		}

		ctx.TypeDefs = append(ctx.TypeDefs, typeDef)
	}
	return nil
}

// Generates type definitions for any custom types defined in the
// components/responses section of the Swagger spec.
func GenerateTypesForResponses(ctx *genCtx, t *template.Template, responses openapi3.Responses) error {

	for _, responseName := range SortedResponsesKeys(responses) {
		responseOrRef := responses[responseName]

		// We have to generate the response object. We're only going to
		// handle application/json media types here. Other responses should
		// simply be specified as strings or byte arrays.
		response := responseOrRef.Value
		jsonResponse, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(ctx, jsonResponse.Schema, []string{responseName})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in response %s", responseName))
			}

			typeDef := TypeDefinition{
				JsonName: responseName,
				Schema:   goType,
				TypeName: SchemaNameToTypeName(responseName),
			}

			if responseOrRef.Ref != "" {
				// Generate a reference type for referenced parameters
				refType, err := RefPathToGoType(responseOrRef.Ref, ctx.ImportedTypes)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in parameter %s", responseOrRef.Ref, responseName))
				}
				typeDef.TypeName = SchemaNameToTypeName(refType)
			}
			ctx.TypeDefs = append(ctx.TypeDefs, typeDef)
		}
	}
	return nil
}

// Generates type definitions for any custom types defined in the
// components/requestBodies section of the Swagger spec.
func GenerateTypesForRequestBodies(ctx *genCtx, t *template.Template, bodies map[string]*openapi3.RequestBodyRef) error {
	var types []TypeDefinition

	for _, bodyName := range SortedRequestBodyKeys(bodies) {
		bodyOrRef := bodies[bodyName]

		// As for responses, we will only generate Go code for JSON bodies,
		// the other body formats are up to the user.
		response := bodyOrRef.Value
		jsonBody, found := response.Content["application/json"]
		if found {
			goType, err := GenerateGoSchema(ctx, jsonBody.Schema, []string{bodyName})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error generating Go type for schema in body %s", bodyName))
			}

			typeDef := TypeDefinition{
				JsonName: bodyName,
				Schema:   goType,
				TypeName: SchemaNameToTypeName(bodyName),
			}

			if bodyOrRef.Ref != "" {
				// Generate a reference type for referenced bodies
				refType, err := RefPathToGoType(bodyOrRef.Ref, ctx.ImportedTypes)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("error generating Go type for (%s) in body %s", bodyOrRef.Ref, bodyName))
				}
				typeDef.TypeName = SchemaNameToTypeName(refType)
			}
			types = append(types, typeDef)
		}
	}
	ctx.TypeDefs = append(ctx.TypeDefs, types...)
	return nil
}

// Helper function to pass a bunch of types to the template engine, and buffer
// its output into a string.
func GenerateTypes(ctx *genCtx, t *template.Template) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := t.ExecuteTemplate(w, "typedef.tmpl", ctx)
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
func GenerateImports(ctx *genCtx, t *template.Template, imports []TypeImportSpec) (string, error) {
	sort.Slice(imports, func(i, j int) bool { return imports[i].ImportPath < imports[j].ImportPath })

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	context := struct {
		Imports     []TypeImportSpec
		PackageName string
	}{
		Imports:     imports,
		PackageName: ctx.PackageName,
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

func getDecorators(decorators map[string]Decorator, s Schema) map[string]Decorator {
	if len(s.Decorators) > 0 {
		for k, v := range s.Decorators {
			decorators[k] = v
		}
	}
	if s.AdditionalPropertiesType != nil {
		m := getDecorators(decorators, *s.AdditionalPropertiesType)
		for k, v := range m {
			decorators[k] = v
		}
	}
	for _, p := range s.Properties {
		m := getDecorators(decorators, p.Schema)
		for k, v := range m {
			decorators[k] = v
		}
	}
	for _, p := range s.AdditionalTypes {
		m := getDecorators(decorators, p.Schema)
		for k, v := range m {
			decorators[k] = v
		}
	}
	return decorators
}

// Generate all the glue code which provides the API for interacting with
// additional properties and JSON-ification
func GenerateAdditionalPropertyBoilerplate(ctx *genCtx, t *template.Template, td []TypeDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	var filteredTypes []TypeDefinition
	decorators := make(map[string]Decorator)
	for _, t := range td {
		if t.Schema.HasAdditionalProperties {
			filteredTypes = append(filteredTypes, t)
		}
		// If the type is an empty interface, can assume it's decorators are held by parent that references it
		if t.Schema.GoType != "interface{}" {
			decorators = getDecorators(decorators, t.Schema)
		}
	}

	// filter out all decorators with a schema path that is deeper than SchemaName.property
	// since the decode library always operates on an object of the current schema, iterating over its properties
	for k, _ := range decorators {
		if strings.Count(k, ".") > 1 {
			delete(decorators, k)
		}
	}

	// todo: this has side effects that should be computed elsewhere
	ctx.HasDecorators = ctx.HasDecorators || len(decorators) > 0

	type factoryMap struct {
		Discriminator string
		Factories     []Decorator
	}

	dsm := map[string]Decorator{}
	dsp := map[string]factoryMap{}
	for _, v := range decorators {
		dsm[v.SchemaName] = v

		fm := dsp[v.SchemaPath]
		fm.Discriminator = v.Discriminator
		fm.Factories = append(fm.Factories, v)
		dsp[v.SchemaPath] = fm

	}
	for _, v := range dsp {
		sort.Slice(v.Factories, func(i, j int) bool { return v.Factories[i].JSONName < v.Factories[j].JSONName })
	}

	// issue a warning if dsm contains schemas that have duplicate  JSONName (this will cause TypeFactory to have
	// duplicate map entries and will therefore not compile. TypeFactory is intended to be removed. This is a temporary measure
	// todo: remove once we no longer use type factory
	djn := map[string]Decorator{}
	for _, v := range dsm {
		if d, ok := djn[v.JSONName]; ok {
			fmt.Fprintf(os.Stderr, "Duplicate Discriminator JSON value `%s` (schema1: %s, schema2: %s). Code will not compile\n", v.JSONName, d.SchemaName, v.SchemaName)
		}
		djn[v.JSONName] = v
	}

	context := struct {
		Types            []TypeDefinition
		Decorators       map[string]Decorator
		DecoratedSchemas map[string]Decorator
		DecoratedPaths   map[string]factoryMap
		HasDecorators    bool
	}{
		Types:            filteredTypes,
		Decorators:       decorators,
		DecoratedSchemas: dsm,
		DecoratedPaths:   dsp,
		HasDecorators:    ctx.HasDecorators,
	}

	//debug.PrintStack()
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

// SanitizeCode runs sanitizers across the generated Go code to ensure the
// generated code will be able to compile.
func SanitizeCode(goCode string) string {
	// remove any byte-order-marks which break Go-Code
	// See: https://groups.google.com/forum/#!topic/golang-nuts/OToNIPdfkks
	return strings.Replace(goCode, "\uFEFF", "", -1)
}

func filterOperationsByTag(swagger *openapi3.Swagger, opts Options) {
	if len(opts.ExcludeTags) > 0 {
		excludeOperationsWithTags(swagger.Paths, opts.ExcludeTags)
	}
	if len(opts.IncludeTags) > 0 {
		includeOperationsWithTags(swagger.Paths, opts.IncludeTags, false)
	}
}

func excludeOperationsWithTags(paths openapi3.Paths, tags []string) {
	includeOperationsWithTags(paths, tags, true)
}

func includeOperationsWithTags(paths openapi3.Paths, tags []string, exclude bool) {
	for _, pathItem := range paths {
		ops := pathItem.Operations()
		names := make([]string, 0, len(ops))
		for name, op := range ops {
			if operationHasTag(op, tags) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			pathItem.SetOperation(name, nil)
		}
	}
}

//operationHasTag returns true if the operation is tagged with any of tags
func operationHasTag(op *openapi3.Operation, tags []string) bool {
	if op == nil {
		return false
	}
	for _, hasTag := range op.Tags {
		for _, wantTag := range tags {
			if hasTag == wantTag {
				return true
			}
		}
	}
	return false
}
