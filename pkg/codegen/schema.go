package codegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
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
			field += fmt.Sprintf(" `json:\"%s,omitempty\"`", s.JsonName)
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

// This structure is passed into our type generation code to give the templating
// system the context needed to produce our type definitions.
type TypeDefinition struct {
	TypeName     string // The Go type name of an object
	JsonTypeName string // The corresponding JSON field name
	TypeDef      string // The Go type definition for the type
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

// This constructs a Go type for a parameter, looking at either the schema or
// the content, whichever is available
func paramToGoType(param *openapi3.Parameter) (string, error) {
	if param.Content == nil && param.Schema == nil {
		return "", fmt.Errorf("parameter '%s' has no schema or content", param.Name)
	}

	// We can process the schema through the generic schema processor
	if param.Schema != nil {
		return schemaToGoType(param.Schema, true)
	}

	// At this point, we have a content type. We know how to deal with
	// application/json, but if multiple formats are present, we can't do anything,
	// so we'll return the parameter as a string, not bothering to decode it.
	if len(param.Content) > 1 {
		return "string", nil
	}

	// Otherwise, look for application/json in there
	mt, found := param.Content["application/json"]
	if !found {
		// If we don't have json, it's a string
		return "string", nil
	}

	// For json, we go through the standard schema mechanism
	return schemaToGoType(mt.Schema, true)
}
