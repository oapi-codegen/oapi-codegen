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
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/tools/imports"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/constants"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/openapiv3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/preprocess"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/schema"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/singleton"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// Embed the templates directory
//
//go:embed templates
var templates embed.FS

// Generate uses the Go templating engine to generate all of our server wrappers from
// the descriptions we've built up above from the schema objects.
// opts defines
func Generate(spec *openapi3.T, opts openapiv3.Configuration) (string, error) {
	// This is global state
	singleton.GlobalState.Options = opts
	singleton.GlobalState.Spec = spec
	singleton.GlobalState.ImportMapping = singleton.ConstructImportMapping(opts.ImportMapping)

	preprocess.FilterOperationsByTag(spec, opts)
	preprocess.FilterOperationsByOperationID(spec, opts)
	if !opts.OutputOptions.SkipPrune {
		preprocess.PruneUnusedComponents(spec)
	}

	// if we are provided an override for the response type suffix update it
	if opts.OutputOptions.ResponseTypeSuffix != "" {
		constants.ResponseTypeSuffix = opts.OutputOptions.ResponseTypeSuffix
	}

	if singleton.GlobalState.Options.OutputOptions.ClientTypeName == "" {
		singleton.GlobalState.Options.OutputOptions.ClientTypeName = defaultClientTypeName
	}

	nameNormalizerFunction := schema.NameNormalizerFunction(opts.OutputOptions.NameNormalizer)
	nameNormalizer := schema.NameNormalizers[nameNormalizerFunction]
	if nameNormalizer == nil {
		return "", fmt.Errorf(`the name-normalizer option %v could not be found among options %q`,
			opts.OutputOptions.NameNormalizer, schema.NameNormalizers.Options())
	}

	// This creates the golang templates text package
	TemplateFunctions["opts"] = func() openapiv3.Configuration { return singleton.GlobalState.Options }
	t := template.New("oapi-codegen").Funcs(TemplateFunctions)
	// This parses all of our own template files into the template object
	// above
	err := LoadTemplates(templates, t)
	if err != nil {
		return "", fmt.Errorf("error parsing oapi-codegen templates: %w", err)
	}

	// load user-provided templates. Will Override built-in versions.
	for name, template := range opts.OutputOptions.UserTemplates {
		utpl := t.New(name)

		txt, err := GetUserTemplateText(template)
		if err != nil {
			return "", fmt.Errorf("error loading user-provided template %q: %w", name, err)
		}

		_, err = utpl.Parse(txt)
		if err != nil {
			return "", fmt.Errorf("error parsing user-provided template %q: %w", name, err)
		}
	}

	ops, err := OperationDefinitions(spec, opts.OutputOptions.InitialismOverrides)
	if err != nil {
		return "", fmt.Errorf("error creating operation definitions: %w", err)
	}

	xGoTypeImports, err := OperationImports(ops)
	if err != nil {
		return "", fmt.Errorf("error getting operation imports: %w", err)
	}

	var typeDefinitions, constantDefinitions string
	if opts.Generate.Models {
		typeDefinitions, err = GenerateTypeDefinitions(t, spec, ops, opts.OutputOptions.ExcludeSchemas)
		if err != nil {
			return "", fmt.Errorf("error generating type definitions: %w", err)
		}

		constantDefinitions, err = GenerateConstants(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating constants: %w", err)
		}

		imprts, err := GetTypeDefinitionsImports(spec, opts.OutputOptions.ExcludeSchemas)
		if err != nil {
			return "", fmt.Errorf("error getting type definition imports: %w", err)
		}
		schema.MergeImports(xGoTypeImports, imprts)
	}

	var irisServerOut string
	if opts.Generate.IrisServer {
		irisServerOut, err = GenerateIrisServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var echoServerOut string
	if opts.Generate.EchoServer {
		echoServerOut, err = GenerateEchoServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var chiServerOut string
	if opts.Generate.ChiServer {
		chiServerOut, err = GenerateChiServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var fiberServerOut string
	if opts.Generate.FiberServer {
		fiberServerOut, err = GenerateFiberServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var ginServerOut string
	if opts.Generate.GinServer {
		ginServerOut, err = GenerateGinServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var gorillaServerOut string
	if opts.Generate.GorillaServer {
		gorillaServerOut, err = GenerateGorillaServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var stdHTTPServerOut string
	if opts.Generate.StdHTTPServer {
		stdHTTPServerOut, err = GenerateStdHTTPServer(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var strictServerOut string
	if opts.Generate.Strict {
		var responses []schema.ResponseDefinition
		if spec.Components != nil {
			responses, err = GenerateResponseDefinitions("", spec.Components.Responses)
			if err != nil {
				return "", fmt.Errorf("error generation response definitions for schema: %w", err)
			}
		}
		strictServerResponses, err := GenerateStrictResponses(t, responses)
		if err != nil {
			return "", fmt.Errorf("error generation response definitions for schema: %w", err)
		}
		strictServerOut, err = GenerateStrictServer(t, ops, opts)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
		strictServerOut = strictServerResponses + strictServerOut
	}

	var clientOut string
	if opts.Generate.Client {
		clientOut, err = GenerateClient(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating client: %w", err)
		}
	}

	var clientWithResponsesOut string
	if opts.Generate.Client {
		clientWithResponsesOut, err = GenerateClientWithResponses(t, ops)
		if err != nil {
			return "", fmt.Errorf("error generating client with responses: %w", err)
		}
	}

	var inlinedSpec string
	if opts.Generate.EmbeddedSpec {
		inlinedSpec, err = GenerateInlinedSpec(t, singleton.GlobalState.ImportMapping, spec)
		if err != nil {
			return "", fmt.Errorf("error generating Go handlers for Paths: %w", err)
		}
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	externalImports := append(singleton.GlobalState.ImportMapping.GoImports(), singleton.ImportMap(xGoTypeImports).GoImports()...)
	importsOut, err := GenerateImports(
		t,
		externalImports,
		opts.PackageName,
		opts.NoVCSVersionOverride,
	)
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

	if opts.Generate.Client {
		_, err = w.WriteString(clientOut)
		if err != nil {
			return "", fmt.Errorf("error writing client: %w", err)
		}
		_, err = w.WriteString(clientWithResponsesOut)
		if err != nil {
			return "", fmt.Errorf("error writing client: %w", err)
		}
	}

	if opts.Generate.IrisServer {
		_, err = w.WriteString(irisServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}

	}

	if opts.Generate.EchoServer {
		_, err = w.WriteString(echoServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.ChiServer {
		_, err = w.WriteString(chiServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.FiberServer {
		_, err = w.WriteString(fiberServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.GinServer {
		_, err = w.WriteString(ginServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.GorillaServer {
		_, err = w.WriteString(gorillaServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.StdHTTPServer {
		_, err = w.WriteString(stdHTTPServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.Strict {
		_, err = w.WriteString(strictServerOut)
		if err != nil {
			return "", fmt.Errorf("error writing server path handlers: %w", err)
		}
	}

	if opts.Generate.EmbeddedSpec {
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
	if opts.OutputOptions.SkipFmt {
		return goCode, nil
	}

	outBytes, err := imports.Process(opts.PackageName+".go", []byte(goCode), nil)
	if err != nil {
		return "", fmt.Errorf("error formatting Go code %s: %w", goCode, err)
	}
	return string(outBytes), nil
}

func GenerateTypeDefinitions(t *template.Template, swagger *openapi3.T, ops []schema.OperationDefinition, excludeSchemas []string) (string, error) {
	var allTypes []schema.TypeDefinition
	if swagger.Components != nil {
		schemaTypes, err := GenerateTypesForSchemas(t, swagger.Components.Schemas, excludeSchemas)
		if err != nil {
			return "", fmt.Errorf("error generating Go types for component schemas: %w", err)
		}

		paramTypes, err := GenerateTypesForParameters(t, swagger.Components.Parameters)
		if err != nil {
			return "", fmt.Errorf("error generating Go types for component parameters: %w", err)
		}
		allTypes = append(schemaTypes, paramTypes...)

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
	}

	// Go through all operations, and add their types to allTypes, so that we can
	// scan all of them for enums. Operation definitions are handled differently
	// from the rest, so let's keep track of enumTypes separately, which will contain
	// all types needed to be scanned for enums, which includes those within operations.
	enumTypes := allTypes
	for _, op := range ops {
		enumTypes = append(enumTypes, op.TypeDefinitions...)
	}

	operationsOut, err := GenerateTypesForOperations(t, ops)
	if err != nil {
		return "", fmt.Errorf("error generating Go types for component request bodies: %w", err)
	}

	enumsOut, err := GenerateEnums(t, enumTypes)
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

	unionBoilerplate, err := GenerateUnionBoilerplate(t, allTypes)
	if err != nil {
		return "", fmt.Errorf("error generating union boilerplate: %w", err)
	}

	unionAndAdditionalBoilerplate, err := GenerateUnionAndAdditionalProopertiesBoilerplate(t, allTypes)
	if err != nil {
		return "", fmt.Errorf("error generating boilerplate for union types with additionalProperties: %w", err)
	}

	typeDefinitions := strings.Join([]string{enumsOut, typesOut, operationsOut, allOfBoilerplate, unionBoilerplate, unionAndAdditionalBoilerplate}, "")
	return typeDefinitions, nil
}

// GenerateConstants generates operation ids, context keys, paths, etc. to be exported as constants
func GenerateConstants(t *template.Template, ops []schema.OperationDefinition) (string, error) {
	constants := schema.Constants{
		SecuritySchemeProviderNames: []string{},
	}

	providerNameMap := map[string]struct{}{}
	for _, op := range ops {
		for _, def := range op.SecurityDefinitions {
			providerName := schema.SanitizeGoIdentity(def.ProviderName)
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

// GenerateTypesForSchemas generates type definitions for any custom types defined in the
// components/schemas section of the Swagger spec.
func GenerateTypesForSchemas(t *template.Template, schemas map[string]*openapi3.SchemaRef, excludeSchemas []string) ([]schema.TypeDefinition, error) {
	excludeSchemasMap := make(map[string]bool)
	for _, schema := range excludeSchemas {
		excludeSchemasMap[schema] = true
	}
	types := make([]schema.TypeDefinition, 0)
	// We're going to define Go types for every object under components/schemas
	for _, schemaName := range schema.SortedSchemaKeys(schemas) {
		if _, ok := excludeSchemasMap[schemaName]; ok {
			continue
		}
		schemaRef := schemas[schemaName]

		goSchema, err := schema.GenerateGoSchema(schemaRef, []string{schemaName})
		if err != nil {
			return nil, fmt.Errorf("error converting Schema %s to Go type: %w", schemaName, err)
		}

		goTypeName, err := schema.RenameSchema(schemaName, schemaRef)
		if err != nil {
			return nil, fmt.Errorf("error making name for components/schemas/%s: %w", schemaName, err)
		}

		types = append(types, schema.TypeDefinition{
			JsonName: schemaName,
			TypeName: goTypeName,
			Schema:   goSchema,
		})

		types = append(types, goSchema.AdditionalTypes...)
	}
	return types, nil
}

// GenerateTypesForParameters generates type definitions for any custom types defined in the
// components/parameters section of the Swagger spec.
func GenerateTypesForParameters(t *template.Template, params map[string]*openapi3.ParameterRef) ([]schema.TypeDefinition, error) {
	var types []schema.TypeDefinition
	for _, paramName := range schema.SortedMapKeys(params) {
		paramOrRef := params[paramName]

		goType, err := schema.ParamToGoType(paramOrRef.Value, nil)
		if err != nil {
			return nil, fmt.Errorf("error generating Go type for schema in parameter %s: %w", paramName, err)
		}

		goTypeName, err := schema.RenameParameter(paramName, paramOrRef)
		if err != nil {
			return nil, fmt.Errorf("error making name for components/parameters/%s: %w", paramName, err)
		}

		typeDef := schema.TypeDefinition{
			JsonName: paramName,
			Schema:   goType,
			TypeName: goTypeName,
		}

		if paramOrRef.Ref != "" {
			// Generate a reference type for referenced parameters
			refType, err := schema.RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return nil, fmt.Errorf("error generating Go type for (%s) in parameter %s: %w", paramOrRef.Ref, paramName, err)
			}
			typeDef.TypeName = schema.SchemaNameToTypeName(refType)
		}

		types = append(types, typeDef)
	}
	return types, nil
}

// GenerateTypesForResponses generates type definitions for any custom types defined in the
// components/responses section of the Swagger spec.
func GenerateTypesForResponses(t *template.Template, responses openapi3.ResponseBodies) ([]schema.TypeDefinition, error) {
	var types []schema.TypeDefinition

	for _, responseName := range schema.SortedMapKeys(responses) {
		responseOrRef := responses[responseName]

		// We have to generate the response object. We're only going to
		// handle media types that conform to JSON. Other responses should
		// simply be specified as strings or byte arrays.
		response := responseOrRef.Value

		jsonCount := 0
		for mediaType := range response.Content {
			if util.IsMediaTypeJson(mediaType) {
				jsonCount++
			}
		}

		SortedMapKeys := schema.SortedMapKeys(response.Content)
		for _, mediaType := range SortedMapKeys {
			response := response.Content[mediaType]
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}

			goType, err := schema.GenerateGoSchema(response.Schema, []string{responseName})
			if err != nil {
				return nil, fmt.Errorf("error generating Go type for schema in response %s: %w", responseName, err)
			}

			goTypeName, err := schema.RenameResponse(responseName, responseOrRef)
			if err != nil {
				return nil, fmt.Errorf("error making name for components/responses/%s: %w", responseName, err)
			}

			typeDef := schema.TypeDefinition{
				JsonName: responseName,
				Schema:   goType,
				TypeName: goTypeName,
			}

			if responseOrRef.Ref != "" {
				// Generate a reference type for referenced parameters
				refType, err := schema.RefPathToGoType(responseOrRef.Ref)
				if err != nil {
					return nil, fmt.Errorf("error generating Go type for (%s) in parameter %s: %w", responseOrRef.Ref, responseName, err)
				}
				typeDef.TypeName = schema.SchemaNameToTypeName(refType)
			}

			if jsonCount > 1 {
				typeDef.TypeName = typeDef.TypeName + schema.MediaTypeToCamelCase(mediaType)
			}

			types = append(types, typeDef)
		}
	}
	return types, nil
}

// GenerateTypesForRequestBodies generates type definitions for any custom types defined in the
// components/requestBodies section of the Swagger spec.
func GenerateTypesForRequestBodies(t *template.Template, bodies map[string]*openapi3.RequestBodyRef) ([]schema.TypeDefinition, error) {
	var types []schema.TypeDefinition

	for _, requestBodyName := range schema.SortedMapKeys(bodies) {
		requestBodyRef := bodies[requestBodyName]

		// As for responses, we will only generate Go code for JSON bodies,
		// the other body formats are up to the user.
		response := requestBodyRef.Value
		for mediaType, body := range response.Content {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}

			goType, err := schema.GenerateGoSchema(body.Schema, []string{requestBodyName})
			if err != nil {
				return nil, fmt.Errorf("error generating Go type for schema in body %s: %w", requestBodyName, err)
			}

			goTypeName, err := schema.RenameRequestBody(requestBodyName, requestBodyRef)
			if err != nil {
				return nil, fmt.Errorf("error making name for components/schemas/%s: %w", requestBodyName, err)
			}

			typeDef := schema.TypeDefinition{
				JsonName: requestBodyName,
				Schema:   goType,
				TypeName: goTypeName,
			}

			if requestBodyRef.Ref != "" {
				// Generate a reference type for referenced bodies
				refType, err := schema.RefPathToGoType(requestBodyRef.Ref)
				if err != nil {
					return nil, fmt.Errorf("error generating Go type for (%s) in body %s: %w", requestBodyRef.Ref, requestBodyName, err)
				}
				typeDef.TypeName = schema.SchemaNameToTypeName(refType)
			}
			types = append(types, typeDef)
		}
	}
	return types, nil
}

// GenerateTypes passes a bunch of types to the template engine, and buffers
// its output into a string.
func GenerateTypes(t *template.Template, types []schema.TypeDefinition) (string, error) {
	m := map[string]schema.TypeDefinition{}
	var ts []schema.TypeDefinition

	for _, typ := range types {
		if prevType, found := m[typ.TypeName]; found {
			// If type names collide, we need to see if they refer to the same
			// exact type definition, in which case, we can de-dupe. If they don't
			// match, we error out.
			if schema.TypeDefinitionsEquivalent(prevType, typ) {
				continue
			}
			// We want to create an error when we try to define the same type twice.
			return "", fmt.Errorf("duplicate typename '%s' detected, can't auto-rename, "+
				"please use x-go-name to specify your own name for one of them", typ.TypeName)
		}

		m[typ.TypeName] = typ

		ts = append(ts, typ)
	}

	context := struct {
		Types []schema.TypeDefinition
	}{
		Types: ts,
	}

	return GenerateTemplates([]string{"typedef.tmpl"}, t, context)
}

func GenerateEnums(t *template.Template, types []schema.TypeDefinition) (string, error) {
	enums := []schema.EnumDefinition{}

	// Keep track of which enums we've generated
	m := map[string]bool{}

	// These are all types defined globally
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
			enums = append(enums, schema.EnumDefinition{
				Schema:         tp.Schema,
				TypeName:       tp.TypeName,
				ValueWrapper:   wrapper,
				PrefixTypeName: singleton.GlobalState.Options.Compatibility.AlwaysPrefixEnumValues,
			})
		}
	}

	// Now, go through all the enums, and figure out if we have conflicts with
	// any others.
	for i := range enums {
		// Look through all other enums not compared so far. Make sure we don't
		// compare against self.
		e1 := enums[i]
		for j := i + 1; j < len(enums); j++ {
			e2 := enums[j]

			for e1key := range e1.GetValues() {
				_, found := e2.GetValues()[e1key]
				if found {
					e1.PrefixTypeName = true
					e2.PrefixTypeName = true
					enums[i] = e1
					enums[j] = e2
					break
				}
			}
		}

		// now see if this enum conflicts with any global type names.
		for _, tp := range types {
			// Skip over enums, since we've handled those above.
			if len(tp.Schema.EnumValues) > 0 {
				continue
			}
			_, found := e1.Schema.EnumValues[tp.TypeName]
			if found {
				e1.PrefixTypeName = true
				enums[i] = e1
			}
		}

		// Another edge case is that an enum value can conflict with its own
		// type name.
		_, found := e1.GetValues()[e1.TypeName]
		if found {
			e1.PrefixTypeName = true
			enums[i] = e1
		}
	}

	// Now see if enums conflict with any non-enum typenames

	return GenerateTemplates([]string{"constants.tmpl"}, t, schema.Constants{EnumDefinitions: enums})
}

// GenerateImports generates our import statements and package definition.
func GenerateImports(t *template.Template, externalImports []string, packageName string, versionOverride *string) (string, error) {
	// Read build version for incorporating into generated files
	// Unit tests have ok=false, so we'll just use "unknown" for the
	// version if we can't read this.

	modulePath := "unknown module path"
	moduleVersion := "unknown version"
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Path != "" {
			modulePath = bi.Main.Path
		}
		if bi.Main.Version != "" {
			moduleVersion = bi.Main.Version
		}
		if versionOverride != nil {
			moduleVersion = *versionOverride
		}
	}

	context := struct {
		ExternalImports   []string
		PackageName       string
		ModuleName        string
		Version           string
		AdditionalImports []openapiv3.AdditionalImport
	}{
		ExternalImports:   externalImports,
		PackageName:       packageName,
		ModuleName:        modulePath,
		Version:           moduleVersion,
		AdditionalImports: singleton.GlobalState.Options.AdditionalImports,
	}

	return GenerateTemplates([]string{"imports.tmpl"}, t, context)
}

// GenerateAdditionalPropertyBoilerplate generates all the glue code which provides
// the API for interacting with additional properties and JSON-ification
func GenerateAdditionalPropertyBoilerplate(t *template.Template, typeDefs []schema.TypeDefinition) (string, error) {
	var filteredTypes []schema.TypeDefinition

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
		Types []schema.TypeDefinition
	}{
		Types: filteredTypes,
	}

	return GenerateTemplates([]string{"additional-properties.tmpl"}, t, context)
}

func GenerateUnionBoilerplate(t *template.Template, typeDefs []schema.TypeDefinition) (string, error) {
	var filteredTypes []schema.TypeDefinition
	for _, t := range typeDefs {
		if len(t.Schema.UnionElements) != 0 {
			filteredTypes = append(filteredTypes, t)
		}
	}

	if len(filteredTypes) == 0 {
		return "", nil
	}

	context := struct {
		Types []schema.TypeDefinition
	}{
		Types: filteredTypes,
	}

	return GenerateTemplates([]string{"union.tmpl"}, t, context)
}

func GenerateUnionAndAdditionalProopertiesBoilerplate(t *template.Template, typeDefs []schema.TypeDefinition) (string, error) {
	var filteredTypes []schema.TypeDefinition
	for _, t := range typeDefs {
		if len(t.Schema.UnionElements) != 0 && t.Schema.HasAdditionalProperties {
			filteredTypes = append(filteredTypes, t)
		}
	}

	if len(filteredTypes) == 0 {
		return "", nil
	}
	context := struct {
		Types []schema.TypeDefinition
	}{
		Types: filteredTypes,
	}

	return GenerateTemplates([]string{"union-and-additional-properties.tmpl"}, t, context)
}

// SanitizeCode runs sanitizers across the generated Go code to ensure the
// generated code will be able to compile.
func SanitizeCode(goCode string) string {
	// remove any byte-order-marks which break Go-Code
	// See: https://groups.google.com/forum/#!topic/golang-nuts/OToNIPdfkks
	return strings.ReplaceAll(goCode, "\uFEFF", "")
}

// GetUserTemplateText attempts to retrieve the template text from a passed in URL or file
// path when inputData is more than one line.
// This function will attempt to load a file first, and if it fails, will try to get the
// data from the remote endpoint.
// The timeout for remote download file is 30 seconds.
func GetUserTemplateText(inputData string) (template string, err error) {
	// if the input data is more than one line, assume its a template and return that data.
	if strings.Contains(inputData, "\n") {
		return inputData, nil
	}

	// load data from file
	data, err := os.ReadFile(inputData)
	// return data if found and loaded
	if err == nil {
		return string(data), nil
	}

	// check for non "not found" errors
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to open file %s: %w", inputData, err)
	}

	// attempt to get data from url with timeout
	const downloadTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, inputData, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request GET %s: %w", inputData, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute GET request data from %s: %w", inputData, err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("got non %d status code on GET %s", resp.StatusCode, inputData)
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from GET %s: %w", inputData, err)
	}

	return string(data), nil
}

// LoadTemplates loads all of our template files into a text/template. The
// path of template is relative to the templates directory.
func LoadTemplates(src embed.FS, t *template.Template) error {
	return fs.WalkDir(src, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}

		buf, err := src.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file '%s': %w", path, err)
		}

		templateName := strings.TrimPrefix(path, "templates/")
		tmpl := t.New(templateName)
		_, err = tmpl.Parse(string(buf))
		if err != nil {
			return fmt.Errorf("parsing template '%s': %w", path, err)
		}
		return nil
	})
}

func OperationSchemaImports(s *schema.Schema) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}

	for _, p := range s.Properties {
		imprts, err := GoSchemaImports(&openapi3.SchemaRef{Value: p.Schema.OAPISchema})
		if err != nil {
			return nil, err
		}
		schema.MergeImports(res, imprts)
	}

	imprts, err := GoSchemaImports(&openapi3.SchemaRef{Value: s.OAPISchema})
	if err != nil {
		return nil, err
	}
	schema.MergeImports(res, imprts)
	return res, nil
}

func OperationImports(ops []schema.OperationDefinition) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	for _, op := range ops {
		for _, pd := range [][]schema.ParameterDefinition{op.PathParams, op.QueryParams} {
			for _, p := range pd {
				imprts, err := OperationSchemaImports(&p.Schema)
				if err != nil {
					return nil, err
				}
				schema.MergeImports(res, imprts)
			}
		}

		for _, b := range op.Bodies {
			imprts, err := OperationSchemaImports(&b.Schema)
			if err != nil {
				return nil, err
			}
			schema.MergeImports(res, imprts)
		}

		for _, b := range op.Responses {
			for _, c := range b.Contents {
				imprts, err := OperationSchemaImports(&c.Schema)
				if err != nil {
					return nil, err
				}
				schema.MergeImports(res, imprts)
			}
		}

	}
	return res, nil
}

func GetTypeDefinitionsImports(swagger *openapi3.T, excludeSchemas []string) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	if swagger.Components == nil {
		return res, nil
	}

	schemaImports, err := GetSchemaImports(swagger.Components.Schemas, excludeSchemas)
	if err != nil {
		return nil, err
	}

	reqBodiesImports, err := GetRequestBodiesImports(swagger.Components.RequestBodies)
	if err != nil {
		return nil, err
	}

	responsesImports, err := GetResponsesImports(swagger.Components.Responses)
	if err != nil {
		return nil, err
	}

	parametersImports, err := GetParametersImports(swagger.Components.Parameters)
	if err != nil {
		return nil, err
	}

	for _, imprts := range []map[string]singleton.GoImport{schemaImports, reqBodiesImports, responsesImports, parametersImports} {
		schema.MergeImports(res, imprts)
	}
	return res, nil
}

func GoSchemaImports(schemas ...*openapi3.SchemaRef) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	for _, sref := range schemas {
		if sref == nil || sref.Value == nil || schema.IsGoTypeReference(sref.Ref) {
			return nil, nil
		}
		if gi, err := schema.ParseGoImportExtension(sref); err != nil {
			return nil, err
		} else {
			if gi != nil {
				res[gi.String()] = *gi
			}
		}
		schemaVal := sref.Value

		t := schemaVal.Type
		if t.Slice() == nil || t.Is("object") {
			for _, v := range schemaVal.Properties {
				imprts, err := GoSchemaImports(v)
				if err != nil {
					return nil, err
				}
				schema.MergeImports(res, imprts)
			}
		} else if t.Is("array") {
			imprts, err := GoSchemaImports(schemaVal.Items)
			if err != nil {
				return nil, err
			}
			schema.MergeImports(res, imprts)
		}
	}
	return res, nil
}

func GetSchemaImports(schemas map[string]*openapi3.SchemaRef, excludeSchemas []string) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	excludeSchemasMap := make(map[string]bool)
	for _, schema := range excludeSchemas {
		excludeSchemasMap[schema] = true
	}
	for schemaName, schem := range schemas {
		if _, ok := excludeSchemasMap[schemaName]; ok {
			continue
		}

		imprts, err := GoSchemaImports(schem)
		if err != nil {
			return nil, err
		}
		schema.MergeImports(res, imprts)
	}
	return res, nil
}

func GetRequestBodiesImports(bodies map[string]*openapi3.RequestBodyRef) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	for _, r := range bodies {
		response := r.Value
		for mediaType, body := range response.Content {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}

			imprts, err := GoSchemaImports(body.Schema)
			if err != nil {
				return nil, err
			}
			schema.MergeImports(res, imprts)
		}
	}
	return res, nil
}

func GetResponsesImports(responses map[string]*openapi3.ResponseRef) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	for _, r := range responses {
		response := r.Value
		for mediaType, body := range response.Content {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}

			imprts, err := GoSchemaImports(body.Schema)
			if err != nil {
				return nil, err
			}
			schema.MergeImports(res, imprts)
		}
	}
	return res, nil
}

func GetParametersImports(params map[string]*openapi3.ParameterRef) (map[string]singleton.GoImport, error) {
	res := map[string]singleton.GoImport{}
	for _, param := range params {
		if param.Value == nil {
			continue
		}
		imprts, err := GoSchemaImports(param.Value.Schema)
		if err != nil {
			return nil, err
		}
		schema.MergeImports(res, imprts)
	}
	return res, nil
}

func SetGlobalStateSpec(spec *openapi3.T) {
	singleton.GlobalState.Spec = spec
}
