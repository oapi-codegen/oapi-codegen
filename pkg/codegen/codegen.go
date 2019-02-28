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
	"errors"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/deepmap/oapi-codegen/v2/pkg/codegen/templates"
)

// This describes a Schema, a type definition.
type SchemaDescriptor struct {
	Required bool   // Is the schema required? If not, we'll pass by pointer
	GoType   string // The Go type needed to represent the json type.
	GoName   string // The Go compatible type name for the type
	JsonName string // The json type name for the type
	IsRef    bool   // Is this schema a reference to predefined object?
}

// Walk the Properties field of the specified schema, generating SchemaDescriptors
// for each of them.
func DescribeSchemaProperties(schema *openapi3.Schema) ([]SchemaDescriptor, error) {
	var desc []SchemaDescriptor
	propNames := SortedSchemaKeys(schema.Properties)
	for _, propName := range propNames {
		propOrRef := schema.Properties[propName]
		propRequired := StringInArray(propName, schema.Required)
		propType, err := schemaToGoType(propOrRef, propRequired)
		if err != nil {
			return nil, fmt.Errorf("error generating type for property '%s': %s", propName, err)
		}
		goFieldName := ToCamelCase(propName)
		desc = append(desc, SchemaDescriptor{
			Required: propRequired,
			GoType:   propType,
			GoName:   goFieldName,
			JsonName: propName,
			IsRef:    propOrRef.Ref != "",
		})
	}

	// Here, we may handle additionalProperties in the future.
	if schema.AdditionalProperties != nil {
		return nil, errors.New("additional_properties are not yet supported")
	}
	if schema.PatternProperties != "" {
		return nil, errors.New("pattern_properties are not yet supported")
	}

	return desc, nil
}

// Given a list of schema descriptors, produce corresponding field names with
// JSON annotations
func GenFieldsFromSchemaDescriptors(desc []SchemaDescriptor) []string {
	var fields []string
	for _, s := range desc {
		field := fmt.Sprintf("    %s %s", s.GoName, s.GoType)
		if s.Required {
			field += fmt.Sprintf(" `json:\"%s\"`", s.JsonName)
		} else {
			field += fmt.Sprintf(" `json,omitempty:\"%s\"`", s.JsonName)
		}
		fields = append(fields, field)
	}
	return fields
}

// Given the list of schema descriptors above, generate a Go struct to represent
// a type, with one field per SchemaDescriptor
func GenStructFromSchemaDescriptors(desc []SchemaDescriptor) string {
	// Start out with struct {
	objectParts := []string{"struct {"}
	// Append all the field definitions
	objectParts = append(objectParts, GenFieldsFromSchemaDescriptors(desc)...)
	// Close the struct
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n")
}

// This function generates an object that is the union of the objects in the
// input array. In the case of Ref objects, we use an embedded struct, otherwise,
// we inline the fields.
func GenStructFromAllOf(allOf []*openapi3.SchemaRef) (string, error) {
	// Start out with struct {
	objectParts := []string{"struct {"}
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref
		val := schemaOrRef.Value

		if ref != "" {
			// We have a referenced type, we will generate an inlined struct
			// member.
			// struct {
			//   InlinedMember
			//   ...
			// }
			goType, err := RefPathToGoType(ref)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts,
				fmt.Sprintf("   // Embedded struct due to allOf(%s)", ref))
			objectParts = append(objectParts,
				fmt.Sprintf("   %s", goType))
		} else {
			// Inline all the fields from the schema into the output struct,
			// just like in the simple case of generating an object.
			descriptors, err := DescribeSchemaProperties(val)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts, "   // Embedded fields due to inline allOf schema")
			objectParts = append(objectParts, GenFieldsFromSchemaDescriptors(descriptors)...)

		}
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n"), nil
}

// This function recursively walks the given schema and generates a Go type to represent
// that schema. References are not followed, and it is assumed that each Ref will be
// a concrete Go type.
// "required" tells us if this field is required. Optional fields have a
// * prepended in the correct place.
// skipRef tells us not to use the top level Ref, instead use the resolved
// type. This is only set to true for handling AllOf
func schemaToGoType(sref *openapi3.SchemaRef, required bool) (string, error) {
	schema := sref.Value
	// We can't support this in any meaningful way
	if schema.AnyOf != nil {
		return "interface{}", nil
	}
	// We can't support this in any meaningful way
	if schema.OneOf != nil {
		return "interface{}", nil
	}
	// AllOf is interesting, and useful. It's the union of a number of other
	// schemas. A common usage is to create a union of an object with an ID,
	// so that in a RESTful paradigm, the Create operation can return
	// (object, id), so that other operations can refer to (id)
	if schema.AllOf != nil {
		outType, err := GenStructFromAllOf(schema.AllOf)
		if err != nil {
			return "", err
		}
		return outType, nil
	}

	// If Ref is set on the SchemaRef, it means that this type is actually a reference to
	// another type. We're not de-referencing, so simply use the referenced type.
	if sref.Ref != "" {
		// Convert the reference path to Go type
		goType, err := RefPathToGoType(sref.Ref)
		if err != nil {
			return "", fmt.Errorf("error turning reference (%s) into a Go type: %s",
				sref.Ref, err)
		}
		if !required {
			goType = "*" + goType
		}
		return goType, nil
	}

	// Here, we handle several types of non-object schemas, and exit early if we
	// can. Objects have a schema of empty string. See
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md#dataTypes
	t := schema.Type
	f := schema.Format
	if t != "" {
		var result string
		switch t {
		case "array":
			// For arrays, we'll get the type of the Items and throw a
			// [] in front of it.
			arrayType, err := schemaToGoType(schema.Items, true)
			if err != nil {
				return "", fmt.Errorf("error generating type for array: %s", err)
			}
			result = "[]" + arrayType
			// Arrays are nullable, so we return our result here, whether or
			// not this field is required
			return result, nil
		case "integer":
			// We default to int32 if format doesn't ask for something else.
			if f == "int64" {
				result = "int64"
			} else if f == "int32" || f == "" {
				result = "int32"
			} else {
				return "", fmt.Errorf("invalid integer format: %s", f)
			}
		case "number":
			// We default to float for "number"
			if f == "double" {
				result = "float64"
			} else if f == "float" || f == "" {
				result = "float32"
			} else {
				return "", fmt.Errorf("invalid number format: %s", f)
			}
		case "boolean":
			if f != "" {
				return "", fmt.Errorf("invalid format (%s) for boolean", f)
			}
			result = "bool"
		case "string":
			switch f {
			case "", "password":
				result = "string"
			case "date-time", "date":
				result = "time.Time"
			default:
				return "", fmt.Errorf("invalid string format: %s", f)
			}
		default:
			return "", fmt.Errorf("unhandled Schema type: %s", t)
		}

		// If a field isn't required, we will pass it by pointer, so that it
		// is nullable.
		if !required {
			result = "*" + result
		}
		return result, nil
	}

	desc, err := DescribeSchemaProperties(schema)
	if err != nil {
		return "", err
	}
	outType := GenStructFromSchemaDescriptors(desc)

	if !required {
		outType = "*" + outType
	}

	return outType, nil
}

// This structure is passed into our type generation code to give the templating
// system the context needed to produce our type definitions.
type TypeDefinition struct {
	TypeName     string // The Go type name of an object
	JsonTypeName string // The corresponding JSON field name
	TypeDef      string // The Go type definition for the type
}

// Uses the Go templating engine to generate all of our server wrappers from
// the descriptions we've built up above from the schema objects.
func GenerateServer(swagger *openapi3.Swagger, packageName string) (string, error) {
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

	// Imports needed for the generated code to compile
	imports := []string{
		"github.com/labstack/echo/v4",
		"github.com/deepmap/oapi-codegen/v2/pkg/codegen",
		"net/http",
		"fmt",
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// Now that we've generated the types, we know whether we reference time.Time,
	// so add that to import list, and generate all the imports.
	for _, str := range []string{schemasOut, paramsOut, responsesOut, bodiesOut} {
		if strings.Contains(str, "time.Time") {
			imports = append(imports, "time")
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

	handlersOut, err := GeneratePathHandlers(t, swagger)
	if err != nil {
		return "", fmt.Errorf("error generating Go handlers for Paths: %s", err)
	}
	_, err = w.WriteString(handlersOut)
	if err != nil {
		return "", fmt.Errorf("error writing path handlers: %s", err)
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
			goType, err := schemaToGoType(paramOrRef.Value.Schema, true)
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
