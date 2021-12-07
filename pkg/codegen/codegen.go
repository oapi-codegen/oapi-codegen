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
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"golang.org/x/tools/imports"

	"github.com/deepmap/oapi-codegen/pkg/codegen/templates"
)

// Options defines the optional code to generate.
type Options struct {
	GenerateChiServer   bool              // GenerateChiServer specifies whether to generate chi server boilerplate
	GenerateEchoServer  bool              // GenerateEchoServer specifies whether to generate echo server boilerplate
	GenerateGinServer   bool              // GenerateGinServer specifies whether to generate echo server boilerplate
	GenerateClient      bool              // GenerateClient specifies whether to generate client boilerplate
	GenerateTypes       bool              // GenerateTypes specifies whether to generate type definitions
	GenerateNestedTypes bool              // GenerateNestedTypes specifies whether to generate anonymous nesting type definitions in Operation/responses
	EmbedSpec           bool              // Whether to embed the swagger spec in the generated code
	SkipFmt             bool              // Whether to skip go imports on the generated code
	SkipPrune           bool              // Whether to skip pruning unused components on the generated code
	AliasTypes          bool              // Whether to alias types if possible
	IncludeTags         []string          // Only include operations that have one of these tags. Ignored when empty.
	ExcludeTags         []string          // Exclude operations that have one of these tags. Ignored when empty.
	UserTemplates       map[string]string // Override built-in templates from user-provided files
	ImportMapping       map[string]string // ImportMapping specifies the golang package path for each external reference
	ExcludeSchemas      []string          // Exclude from generation schemas with given names. Ignored when empty.
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
		typeDefinitions, err = GenerateTypeDefinitions(t, swagger, ops, opts.ExcludeSchemas, opts.GenerateNestedTypes)
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

	var ginServerOut string
	if opts.GenerateGinServer {
		ginServerOut, err = GenerateGinServer(t, ops)
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

	if opts.GenerateGinServer {
		_, err = w.WriteString(ginServerOut)
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

func GenerateTypeDefinitions(t *template.Template, swagger *openapi3.T, ops []OperationDefinition, excludeSchemas []string, includeOpRespNestedTypes bool) (string, error) {
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

	if includeOpRespNestedTypes {
		opRespTypes, err := GenerateTypesForOpResponses(t, ops)
		if err != nil {
			return "", errors.Wrap(err, "error generating Go types for operation responses")
		}
		allTypes = append(allTypes, opRespTypes...)
	}

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

		tds, err := GenerateTypesFromSchemaRef(schemaRef, schemaName)
		if err != nil {
			return nil, fmt.Errorf("error converting Schema %s to Go type: %w", schemaName, err)
		}

		types = append(types, tds...)
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

// Generates type definitions for any Operation/responses of the Swagger spec.
func GenerateTypesForOpResponses(t *template.Template, ops []OperationDefinition) ([]TypeDefinition, error) {
	tds := []TypeDefinition{}
	for _, op := range ops {
		for _, respName := range SortedResponsesKeys(op.Spec.Responses) {
			response := op.Spec.Responses[respName]
			if response.Ref != "" {
				log.Println(op.OperationId + respName + " not supprt response ref")
				continue
			} else if response.Value == nil {
				return nil, errors.New(op.OperationId + respName + " has neither a ref nor a value")
			}
			for _, contentTypeName := range SortedContentKeys(response.Value.Content) {
				content := response.Value.Content[contentTypeName]
				name := op.Spec.OperationID + respName
				var typeName string
				switch {
				case StringInArray(contentTypeName, contentTypesJSON):
					typeName = fmt.Sprintf("%sJSON", SchemaNameToTypeName(name))
				// YAML:
				case StringInArray(contentTypeName, contentTypesYAML):
					typeName = fmt.Sprintf("%sYAML", SchemaNameToTypeName(name))
				// XML:
				case StringInArray(contentTypeName, contentTypesXML):
					typeName = fmt.Sprintf("%sXML", SchemaNameToTypeName(name))
				default:
					ss := strings.Split(contentTypeName, "/")
					typeName = fmt.Sprintf("%s%s", SchemaNameToTypeName(name), strings.ToUpper(ss[len(ss)-1]))
				}
				stds, err := GenerateTypesFromSchemaRef(content.Schema, typeName)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for %s", name))
				}
				tds = append(tds, stds...)
			}
		}
	}
	return tds, nil
}

// Generate type definitions from openapi3.SchemaRef recursively.
func GenerateTypesFromSchemaRef(schemaref *openapi3.SchemaRef, name string) ([]TypeDefinition, error) {
	tds := []TypeDefinition{
		{
			TypeName: SchemaNameToTypeName(name),
			JsonName: name,
		},
	}
	if schemaref == nil {
		tds[0].Schema = Schema{GoType: "interface{}"}
		return tds, nil
	}
	// GoType
	if schemaref.Ref != "" {
		refType, err := RefPathToGoType(schemaref.Ref)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error generating Go type for %s", name))
		}
		tds[0].Schema = Schema{GoType: refType}
		return tds, nil
	}
	schema := schemaref.Value

	// We can't support these in any meaningful way
	if schema.AnyOf != nil || schema.OneOf != nil {
		tds[0].Schema = Schema{GoType: "interface{}"}
		return tds, nil
	}
	// AllOf is interesting, and useful. It's the union of a number of other
	// schemas. A common usage is to create a union of an object with an ID,
	// so that in a RESTful paradigm, the Create operation can return
	// (object, id), so that other operations can refer to (id)
	if schema.AllOf != nil {
		mergedSchemaTypes, err := MergeSchemasAndGenerateTypes(schema.AllOf, name)
		if err != nil {
			return nil, errors.Wrap(err, "error merging schemas")
		}
		return mergedSchemaTypes, nil
	}

	// Check for custom Go type extension
	if extension, ok := schema.Extensions[extPropGoType]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid value for %q", extPropGoType)
		}
		tds[0].Schema = Schema{GoType: typeName}
		return tds, nil
	}

	// Schema type and format, eg. string / binary
	t := schema.Type
	var err error
	// Handle objects and empty schemas first as a special case
	if t == "" || t == "object" {
		var outType string

		if len(schema.Properties) == 0 && !SchemaHasAdditionalProperties(schema) {
			// If the object has no properties or additional properties, we
			// have some special cases for its type.
			if t == "object" {
				// We have an object with no properties. This is a generic object
				// expressed as a map.
				outType = "map[string]interface{}"
			} else { // t == ""
				// If we don't even have the object designator, we're a completely
				// generic type.
				outType = "interface{}"
			}
			tds[0].Schema = Schema{GoType: outType}
		} else {
			// We've got an object with some properties.
			for _, pName := range SortedSchemaKeys(schema.Properties) {
				p := schema.Properties[pName]
				propertyName := SchemaNameToTypeName(name) + SchemaNameToTypeName(pName)
				propertyAndChildrenTypes, err := GenerateTypesFromSchemaRef(p, propertyName)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("error generating Go schema for property '%s'", pName))
				}
				pt := propertyAndChildrenTypes[0]
				ptSchema := pt.Schema
				if len(pt.Schema.Properties) != 0 {
					// new struct
					ptSchema = Schema{GoType: propertyName}
					tds = append(tds, propertyAndChildrenTypes...)
				} else if pt.Schema.ArrayType != nil {
					tds = append(tds, propertyAndChildrenTypes[1:]...)
				}

				required := StringInArray(pName, schema.Required)

				if pt.Schema.HasAdditionalProperties && pt.Schema.RefType == "" {
					// If we have fields present which have additional properties,
					// but are not a pre-defined type, we need to define a type
					// for them, which will be based on the field names we followed
					// to get to the type.
					typeDef := TypeDefinition{
						TypeName: propertyName,
						JsonName: propertyName,
						Schema:   pt.Schema,
					}
					pt.Schema.AdditionalTypes = append(pt.Schema.AdditionalTypes, typeDef)

					pt.Schema.RefType = propertyName
				}
				description := ""
				if p.Value != nil {
					description = p.Value.Description
				}
				prop := Property{
					JsonFieldName:  pName,
					Schema:         ptSchema,
					Required:       required,
					Description:    description,
					Nullable:       p.Value.Nullable,
					ExtensionProps: &p.Value.ExtensionProps,
				}
				tds[0].Schema.Properties = append(tds[0].Schema.Properties, prop)
			}

			tds[0].Schema.HasAdditionalProperties = SchemaHasAdditionalProperties(schema)
			tds[0].Schema.AdditionalPropertiesType = &Schema{
				GoType: "interface{}",
			}
			if schema.AdditionalProperties != nil {
				additionalAndChildrenTypes, err := GenerateTypesFromSchemaRef(schema.AdditionalProperties, name)
				if err != nil {
					return nil, errors.Wrap(err, "error generating type for additional properties")
				}
				tds[0].Schema.AdditionalPropertiesType = &additionalAndChildrenTypes[0].Schema
				if len(additionalAndChildrenTypes) > 1 {
					tds = append(tds, additionalAndChildrenTypes[1:]...)
				}
			}
			tds[0].Schema.GoType = GenStructFromSchema(tds[0].Schema)
		}
		return tds, nil
	} else if len(schema.Enum) > 0 {
		if tds, err = resolveTypeForSchemaRef(schemaref.Value, name, tds); err != nil {
			return nil, err
		}
		enumValues := make([]string, len(schema.Enum))
		for i, enumValue := range schema.Enum {
			enumValues[i] = fmt.Sprintf("%v", enumValue)
		}

		sanitizedValues := SanitizeEnumNames(enumValues)
		tds[0].Schema.EnumValues = make(map[string]string, len(sanitizedValues))
		var constNamePath []string
		for k, v := range sanitizedValues {
			if v == "" {
				constNamePath = []string{name, "Empty"}
			} else {
				constNamePath = []string{name, k}
			}
			tds[0].Schema.EnumValues[SchemaNameToTypeName(PathToTypeName(constNamePath))] = v
		}
	} else {
		if tds, err = resolveTypeForSchemaRef(schemaref.Value, name, tds); err != nil {
			return nil, err
		}
	}
	return tds, nil
}

func resolveTypeForSchemaRef(schema *openapi3.Schema, name string, tds []TypeDefinition) ([]TypeDefinition, error) {
	f := schema.Format
	t := schema.Type

	switch t {
	case "array":
		// For arrays, we'll get the type of the Items and throw a
		// [] in front of it.
		itemName := name + "Item"
		arrAndChildrenTypes, err := GenerateTypesFromSchemaRef(schema.Items, itemName)
		if err != nil {
			return nil, errors.Wrap(err, "error generating type for array")
		}
		if len(arrAndChildrenTypes[0].Schema.Properties) != 0 {
			// new struct
			tds[0].Schema = Schema{GoType: "[]" + itemName, ArrayType: &arrAndChildrenTypes[0].Schema}
			tds = append(tds, arrAndChildrenTypes...)
		} else if arrAndChildrenTypes[0].Schema.ArrayType != nil {
			return nil, errors.New("not supprt array of array: " + name)
			// tds[0].Schema = Schema{GoType: "[]" + name, ArrayType: &arrAndChildrenTypes[0].Schema}
			// tds = append(tds, arrAndChildrenTypes...)
		} else {
			// basic types or ref
			tds[0].Schema = Schema{GoType: "[]" + arrAndChildrenTypes[0].Schema.GoType, ArrayType: &arrAndChildrenTypes[0].Schema}
		}
	case "integer":
		// We default to int if format doesn't ask for something else.
		if f == "int64" {
			tds[0].Schema.GoType = "int64"
		} else if f == "uint64" {
			tds[0].Schema.GoType = "uint64"
		} else if f == "int32" {
			tds[0].Schema.GoType = "int32"
		} else if f == "uint32" {
			tds[0].Schema.GoType = "uint32"
		} else if f == "" {
			tds[0].Schema.GoType = "int"
		} else {
			return nil, fmt.Errorf("invalid integer format: %s", f)
		}
	case "number":
		// We default to float for "number"
		if f == "double" {
			tds[0].Schema.GoType = "float64"
		} else if f == "float" || f == "" {
			tds[0].Schema.GoType = "float32"
		} else {
			return nil, fmt.Errorf("invalid number format: %s", f)
		}
	case "boolean":
		if f != "" {
			return nil, fmt.Errorf("invalid format (%s) for boolean", f)
		}
		tds[0].Schema.GoType = "bool"
	case "string":
		// Special case string formats here.
		switch f {
		case "byte":
			tds[0].Schema.GoType = "[]byte"
		case "email":
			tds[0].Schema.GoType = "openapi_types.Email"
		case "date":
			tds[0].Schema.GoType = "openapi_types.Date"
		case "date-time":
			tds[0].Schema.GoType = "time.Time"
		case "json":
			tds[0].Schema.GoType = "json.RawMessage"
			tds[0].Schema.SkipOptionalPointer = true
		default:
			// All unrecognized formats are simply a regular string.
			tds[0].Schema.GoType = "string"
		}
	default:
		return nil, fmt.Errorf("unhandled Schema type: %s", t)
	}
	return tds, nil
}

// Merge all the fields in the schemas supplied into one giant schema.
func MergeSchemasAndGenerateTypes(allOf []*openapi3.SchemaRef, name string) ([]TypeDefinition, error) {
	tds := []TypeDefinition{{TypeName: SchemaNameToTypeName(name), JsonName: name, Schema: Schema{}}}
	subTDs := []TypeDefinition{}
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref

		var refType string
		var err error
		if ref != "" {
			refType, err = RefPathToGoType(ref)
			if err != nil {
				return nil, errors.Wrap(err, "error converting reference path to a go type")
			}
		}

		schemaAndChildrenTypes, err := GenerateTypesFromSchemaRef(schemaOrRef, name)
		if err != nil {
			return nil, errors.Wrap(err, "error generating Go schema in allOf")
		}
		schemaAndChildrenTypes[0].Schema.RefType = refType
		if len(schemaAndChildrenTypes) > 1 {
			tds = append(tds, schemaAndChildrenTypes[1:]...)
		}

		for _, p := range schemaAndChildrenTypes[0].Schema.Properties {
			err = tds[0].Schema.MergeProperty(p)
			if err != nil {
				return nil, errors.Wrap(err, "error merging properties")
			}
		}

		if schemaAndChildrenTypes[0].Schema.HasAdditionalProperties {
			if tds[0].Schema.HasAdditionalProperties {
				// Both this schema, and the aggregate schema have additional
				// properties, they must match.
				if schemaAndChildrenTypes[0].Schema.AdditionalPropertiesType.TypeDecl() != tds[0].Schema.AdditionalPropertiesType.TypeDecl() {
					return nil, errors.New("additional properties in allOf have incompatible types")
				}
			} else {
				// We're switching from having no additional properties to having
				// them
				tds[0].Schema.HasAdditionalProperties = true
				tds[0].Schema.AdditionalPropertiesType = schemaAndChildrenTypes[0].Schema.AdditionalPropertiesType
			}
		}
		subTDs = append(subTDs, schemaAndChildrenTypes[0])
	}

	// Now, we generate the struct which merges together all the fields.
	tds[0].Schema.GoType = GenStructFromAllOfTypes(subTDs)
	return tds, nil
}

// This function generates an object that is the union of the objects in the
// input array. In the case of Ref objects, we use an embedded struct, otherwise,
// we inline the fields.
func GenStructFromAllOfTypes(allOf []TypeDefinition) string {
	// Start out with struct {
	objectParts := []string{"struct {"}
	for _, td := range allOf {
		ref := td.Schema.RefType
		if ref != "" {
			// We have a referenced type, we will generate an inlined struct
			// member.
			// struct {
			//   InlinedMember
			//   ...
			// }
			objectParts = append(objectParts,
				fmt.Sprintf("   // Embedded struct due to allOf(%s)", ref))
			objectParts = append(objectParts,
				fmt.Sprintf("   %s `yaml:\",inline\"`", ref))
		} else {
			// Inline all the fields from the schema into the output struct,
			// just like in the simple case of generating an object.
			objectParts = append(objectParts, "   // Embedded fields due to inline allOf schema")
			objectParts = append(objectParts, GenFieldsFromProperties(td.Schema.Properties)...)

			if td.Schema.HasAdditionalProperties {
				addPropsType := td.Schema.AdditionalPropertiesType.GoType
				if td.Schema.AdditionalPropertiesType.RefType != "" {
					addPropsType = td.Schema.AdditionalPropertiesType.RefType
				}

				additionalPropertiesPart := fmt.Sprintf("AdditionalProperties map[string]%s `json:\"-\"`", addPropsType)
				if !StringInArray(additionalPropertiesPart, objectParts) {
					objectParts = append(objectParts, additionalPropertiesPart)
				}
			}
		}
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n")
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
