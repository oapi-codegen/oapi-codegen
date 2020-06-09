package codegen

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
)

// This describes a Schema, a type definition.
type Schema struct {
	GoType  string // The Go type needed to represent the schema
	RefType string // If the type has a type name, this is set

	Properties               []Property       // For an object, the fields with names
	HasAdditionalProperties  bool             // Whether we support additional properties
	AdditionalPropertiesType *Schema          // And if we do, their type
	AdditionalTypes          []TypeDefinition // We may need to generate auxiliary helper types, stored here

	SkipOptionalPointer bool // Some types don't need a * in front when they're optional
}

func (s Schema) IsRef() bool {
	return s.RefType != ""
}

func (s Schema) TypeDecl() string {
	if s.IsRef() {
		return s.RefType
	}
	return s.GoType
}

func (s *Schema) MergeProperty(p Property) error {
	// Scan all existing properties for a conflict
	for _, e := range s.Properties {
		if e.JsonFieldName == p.JsonFieldName && !PropertiesEqual(e, p) {
			return errors.New(fmt.Sprintf("property '%s' already exists with a different type", e.JsonFieldName))
		}
	}
	s.Properties = append(s.Properties, p)
	return nil
}

func (s Schema) GetAdditionalTypeDefs() []TypeDefinition {
	var result []TypeDefinition
	for _, p := range s.Properties {
		result = append(result, p.Schema.GetAdditionalTypeDefs()...)
	}
	result = append(result, s.AdditionalTypes...)
	return result
}

type Property struct {
	Description   string
	JsonFieldName string
	Schema        Schema
	Required      bool
	Nullable      bool
}

func (p Property) GoFieldName() string {
	return SchemaNameToTypeName(p.JsonFieldName)
}

func (p Property) GoTypeDef() string {
	typeDef := p.Schema.TypeDecl()
	if !p.Schema.SkipOptionalPointer && (!p.Required || p.Nullable) {
		typeDef = "*" + typeDef
	}
	return typeDef
}

type TypeDefinition struct {
	TypeName     string
	JsonName     string
	ResponseName string
	Schema       Schema
}

func PropertiesEqual(a, b Property) bool {
	return a.JsonFieldName == b.JsonFieldName && a.Schema.TypeDecl() == b.Schema.TypeDecl() && a.Required == b.Required
}

func GenerateGoSchema(sref *openapi3.SchemaRef, path []string) (Schema, error) {
	// If Ref is set on the SchemaRef, it means that this type is actually a reference to
	// another type. We're not de-referencing, so simply use the referenced type.
	var refType string

	// Add a fallback value in case the sref is nil.
	// i.e. the parent schema defines a type:array, but the array has
	// no items defined. Therefore we have at least valid Go-Code.
	if sref == nil {
		return Schema{GoType: "interface{}", RefType: refType}, nil
	}

	schema := sref.Value

	if sref.Ref != "" {
		var err error
		// Convert the reference path to Go type
		refType, err = RefPathToGoType(sref.Ref)
		if err != nil {
			return Schema{}, fmt.Errorf("error turning reference (%s) into a Go type: %s",
				sref.Ref, err)
		}
		return Schema{
			GoType: refType,
		}, nil
	}

	// We can't support this in any meaningful way
	if schema.AnyOf != nil {
		return Schema{GoType: "interface{}", RefType: refType}, nil
	}
	// We can't support this in any meaningful way
	if schema.OneOf != nil {
		return Schema{GoType: "interface{}", RefType: refType}, nil
	}

	// AllOf is interesting, and useful. It's the union of a number of other
	// schemas. A common usage is to create a union of an object with an ID,
	// so that in a RESTful paradigm, the Create operation can return
	// (object, id), so that other operations can refer to (id)
	if schema.AllOf != nil {
		mergedSchema, err := MergeSchemas(schema.AllOf, path)
		if err != nil {
			return Schema{}, errors.Wrap(err, "error merging schemas")
		}
		mergedSchema.RefType = refType
		return mergedSchema, nil
	}

	// Schema type and format, eg. string / binary
	t := schema.Type

	outSchema := Schema{
		RefType: refType,
	}
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
			outSchema.GoType = outType
		} else {
			// We've got an object with some properties.
			for _, pName := range SortedSchemaKeys(schema.Properties) {
				p := schema.Properties[pName]
				propertyPath := append(path, pName)
				pSchema, err := GenerateGoSchema(p, propertyPath)
				if err != nil {
					return Schema{}, errors.Wrap(err, fmt.Sprintf("error generating Go schema for property '%s'", pName))
				}

				required := StringInArray(pName, schema.Required)

				if pSchema.HasAdditionalProperties && pSchema.RefType == "" {
					// If we have fields present which have additional properties,
					// but are not a pre-defined type, we need to define a type
					// for them, which will be based on the field names we followed
					// to get to the type.
					typeName := PathToTypeName(propertyPath)

					typeDef := TypeDefinition{
						TypeName: typeName,
						JsonName: strings.Join(propertyPath, "."),
						Schema:   pSchema,
					}
					pSchema.AdditionalTypes = append(pSchema.AdditionalTypes, typeDef)

					pSchema.RefType = typeName
				}
				description := ""
				if p.Value != nil {
					description = p.Value.Description
				}
				prop := Property{
					JsonFieldName: pName,
					Schema:        pSchema,
					Required:      required,
					Description:   description,
					Nullable:      p.Value.Nullable,
				}
				outSchema.Properties = append(outSchema.Properties, prop)
			}

			outSchema.HasAdditionalProperties = SchemaHasAdditionalProperties(schema)
			outSchema.AdditionalPropertiesType = &Schema{
				GoType: "interface{}",
			}
			if schema.AdditionalProperties != nil {
				additionalSchema, err := GenerateGoSchema(schema.AdditionalProperties, path)
				if err != nil {
					return Schema{}, errors.Wrap(err, "error generating type for additional properties")
				}
				outSchema.AdditionalPropertiesType = &additionalSchema
			}

			outSchema.GoType = GenStructFromSchema(outSchema)
		}
		return outSchema, nil
	} else {
		f := schema.Format

		switch t {
		case "array":
			// For arrays, we'll get the type of the Items and throw a
			// [] in front of it.
			arrayType, err := GenerateGoSchema(schema.Items, path)
			if err != nil {
				return Schema{}, errors.Wrap(err, "error generating type for array")
			}
			outSchema.GoType = "[]" + arrayType.TypeDecl()
			outSchema.Properties = arrayType.Properties
		case "integer":
			// We default to int if format doesn't ask for something else.
			if f == "int64" {
				outSchema.GoType = "int64"
			} else if f == "int32" {
				outSchema.GoType = "int32"
			} else if f == "" {
				outSchema.GoType = "int"
			} else {
				return Schema{}, fmt.Errorf("invalid integer format: %s", f)
			}
		case "number":
			// We default to float for "number"
			if f == "double" {
				outSchema.GoType = "float64"
			} else if f == "float" || f == "" {
				outSchema.GoType = "float32"
			} else {
				return Schema{}, fmt.Errorf("invalid number format: %s", f)
			}
		case "boolean":
			if f != "" {
				return Schema{}, fmt.Errorf("invalid format (%s) for boolean", f)
			}
			outSchema.GoType = "bool"
		case "string":
			// Special case string formats here.
			switch f {
			case "byte":
				outSchema.GoType = "[]byte"
			case "date":
				outSchema.GoType = "openapi_types.Date"
			case "date-time":
				outSchema.GoType = "time.Time"
			case "json":
				outSchema.GoType = "json.RawMessage"
				outSchema.SkipOptionalPointer = true
			default:
				// All unrecognized formats are simply a regular string.
				outSchema.GoType = "string"
			}
		default:
			return Schema{}, fmt.Errorf("unhandled Schema type: %s", t)
		}
	}
	return outSchema, nil
}

// This describes a Schema, a type definition.
type SchemaDescriptor struct {
	Fields                   []FieldDescriptor
	HasAdditionalProperties  bool
	AdditionalPropertiesType string
}

type FieldDescriptor struct {
	Required bool   // Is the schema required? If not, we'll pass by pointer
	GoType   string // The Go type needed to represent the json type.
	GoName   string // The Go compatible type name for the type
	JsonName string // The json type name for the type
	IsRef    bool   // Is this schema a reference to predefined object?
}

// Given a list of schema descriptors, produce corresponding field names with
// JSON annotations
func GenFieldsFromProperties(props []Property) []string {
	var fields []string
	for _, p := range props {
		field := ""
		// Add a comment to a field in case we have one, otherwise skip.
		if p.Description != "" {
			// Separate the comment from a previous-defined, unrelated field.
			// Make sure the actual field is separated by a newline.
			field += fmt.Sprintf("\n%s\n", StringToGoComment(p.Description))
		}
		field += fmt.Sprintf("    %s %s", p.GoFieldName(), p.GoTypeDef())
		if p.Required || p.Nullable {
                        field += fmt.Sprintf(" `json:\"%s\" bson:\"%s\"`", p.JsonFieldName, p.JsonFieldName)
		} else {
                        field += fmt.Sprintf(" `json:\"%s,omitempty\" bson:\"%s,omitempty\"`", p.JsonFieldName, p.JsonFieldName)
		}
		fields = append(fields, field)
	}
	return fields
}

func GenStructFromSchema(schema Schema) string {
	// Start out with struct {
	objectParts := []string{"struct {"}
	// Append all the field definitions
	objectParts = append(objectParts, GenFieldsFromProperties(schema.Properties)...)
	// Close the struct
	if schema.HasAdditionalProperties {
		addPropsType := schema.AdditionalPropertiesType.GoType
		if schema.AdditionalPropertiesType.RefType != "" {
			addPropsType = schema.AdditionalPropertiesType.RefType
		}

		objectParts = append(objectParts,
			fmt.Sprintf("AdditionalProperties map[string]%s `json:\"-\"`", addPropsType))
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n")
}

// Merge all the fields in the schemas supplied into one giant schema.
func MergeSchemas(allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	var outSchema Schema
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref

		var refType string
		var err error
		if ref != "" {
			refType, err = RefPathToGoType(ref)
			if err != nil {
				return Schema{}, errors.Wrap(err, "error converting reference path to a go type")
			}
		}

		schema, err := GenerateGoSchema(schemaOrRef, path)
		if err != nil {
			return Schema{}, errors.Wrap(err, "error generating Go schema in allOf")
		}
		schema.RefType = refType

		for _, p := range schema.Properties {
			err = outSchema.MergeProperty(p)
			if err != nil {
				return Schema{}, errors.Wrap(err, "error merging properties")
			}
		}

		if schema.HasAdditionalProperties {
			if outSchema.HasAdditionalProperties {
				// Both this schema, and the aggregate schema have additional
				// properties, they must match.
				if schema.AdditionalPropertiesType.TypeDecl() != outSchema.AdditionalPropertiesType.TypeDecl() {
					return Schema{}, errors.New("additional properties in allOf have incompatible types")
				}
			} else {
				// We're switching from having no additional properties to having
				// them
				outSchema.HasAdditionalProperties = true
				outSchema.AdditionalPropertiesType = schema.AdditionalPropertiesType
			}
		}
	}

	// Now, we generate the struct which merges together all the fields.
	var err error
	outSchema.GoType, err = GenStructFromAllOf(allOf, path)
	if err != nil {
		return Schema{}, errors.Wrap(err, "unable to generate aggregate type for AllOf")
	}
	return outSchema, nil
}

// This function generates an object that is the union of the objects in the
// input array. In the case of Ref objects, we use an embedded struct, otherwise,
// we inline the fields.
func GenStructFromAllOf(allOf []*openapi3.SchemaRef, path []string) (string, error) {
	// Start out with struct {
	objectParts := []string{"struct {"}
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref
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
                                fmt.Sprintf("   %s `bson:\",inline\"`", goType))
		} else {
			// Inline all the fields from the schema into the output struct,
			// just like in the simple case of generating an object.
			goSchema, err := GenerateGoSchema(schemaOrRef, path)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts, "   // Embedded fields due to inline allOf schema")
			objectParts = append(objectParts, GenFieldsFromProperties(goSchema.Properties)...)

		}
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n"), nil
}

// This constructs a Go type for a parameter, looking at either the schema or
// the content, whichever is available
func paramToGoType(param *openapi3.Parameter, path []string) (Schema, error) {
	if param.Content == nil && param.Schema == nil {
		return Schema{}, fmt.Errorf("parameter '%s' has no schema or content", param.Name)
	}

	// We can process the schema through the generic schema processor
	if param.Schema != nil {
		return GenerateGoSchema(param.Schema, path)
	}

	// At this point, we have a content type. We know how to deal with
	// application/json, but if multiple formats are present, we can't do anything,
	// so we'll return the parameter as a string, not bothering to decode it.
	if len(param.Content) > 1 {
		return Schema{
			GoType: "string",
		}, nil
	}

	// Otherwise, look for application/json in there
	mt, found := param.Content["application/json"]
	if !found {
		// If we don't have json, it's a string
		return Schema{
			GoType: "string",
		}, nil
	}

	// For json, we go through the standard schema mechanism
	return GenerateGoSchema(mt.Schema, path)
}
