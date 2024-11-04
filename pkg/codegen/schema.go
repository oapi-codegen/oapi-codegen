package codegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Schema describes an OpenAPI schema, with lots of helper fields to use in the
// templating engine.
type Schema struct {
	GoType  string // The Go type needed to represent the schema
	RefType string // If the type has a type name, this is set

	ArrayType *Schema // The schema of array element

	EnumValues map[string]string // Enum values

	Properties               []Property       // For an object, the fields with names
	HasAdditionalProperties  bool             // Whether we support additional properties
	AdditionalPropertiesType *Schema          // And if we do, their type
	AdditionalTypes          []TypeDefinition // We may need to generate auxiliary helper types, stored here

	SkipOptionalPointer bool // Some types don't need a * in front when they're optional

	Description string // The description of the element

	UnionElements []UnionElement // Possible elements of oneOf/anyOf union
	Discriminator *Discriminator // Describes which value is stored in a union

	// If this is set, the schema will declare a type via alias, eg,
	// `type Foo = bool`. If this is not set, we will define this type via
	// type definition `type Foo bool`
	//
	// Can be overriden by the OutputOptions#DisableTypeAliasesForType field
	DefineViaAlias bool

	// The original OpenAPIv3 Schema.
	OAPISchema *openapi3.Schema
}

func (s Schema) IsRef() bool {
	return s.RefType != ""
}

func (s Schema) IsExternalRef() bool {
	if !s.IsRef() {
		return false
	}
	return strings.Contains(s.RefType, ".")
}

func (s Schema) TypeDecl() string {
	if s.IsRef() {
		return s.RefType
	}
	return s.GoType
}

// AddProperty adds a new property to the current Schema, and returns an error
// if it collides. Two identical fields will not collide, but two properties by
// the same name, but different definition, will collide. It's safe to merge the
// fields of two schemas with overlapping properties if those properties are
// identical.
func (s *Schema) AddProperty(p Property) error {
	// Scan all existing properties for a conflict
	for _, e := range s.Properties {
		if e.JsonFieldName == p.JsonFieldName && !PropertiesEqual(e, p) {
			return fmt.Errorf("property '%s' already exists with a different type", e.JsonFieldName)
		}
	}
	s.Properties = append(s.Properties, p)
	return nil
}

func (s Schema) GetAdditionalTypeDefs() []TypeDefinition {
	return s.AdditionalTypes
}

type Property struct {
	Description   string
	JsonFieldName string
	Schema        Schema
	Required      bool
	Nullable      bool
	ReadOnly      bool
	WriteOnly     bool
	NeedsFormTag  bool
	Extensions    map[string]interface{}
	Deprecated    bool
}

func (p Property) GoFieldName() string {
	goFieldName := p.JsonFieldName
	if extension, ok := p.Extensions[extGoName]; ok {
		if extGoFieldName, err := extParseGoFieldName(extension); err == nil {
			goFieldName = extGoFieldName
		}
	}

	if globalState.options.Compatibility.AllowUnexportedStructFieldNames {
		if extension, ok := p.Extensions[extOapiCodegenOnlyHonourGoName]; ok {
			if extOapiCodegenOnlyHonourGoName, err := extParseOapiCodegenOnlyHonourGoName(extension); err == nil {
				if extOapiCodegenOnlyHonourGoName {
					return goFieldName
				}
			}
		}
	}

	return SchemaNameToTypeName(goFieldName)
}

func (p Property) GoTypeDef() string {
	typeDef := p.Schema.TypeDecl()
	if globalState.options.OutputOptions.NullableType && p.Nullable {
		return "nullable.Nullable[" + typeDef + "]"
	}
	if !p.Schema.SkipOptionalPointer &&
		(!p.Required || p.Nullable ||
			(p.ReadOnly && (!p.Required || !globalState.options.Compatibility.DisableRequiredReadOnlyAsPointer)) ||
			p.WriteOnly) {

		typeDef = "*" + typeDef
	}
	return typeDef
}

// EnumDefinition holds type information for enum
type EnumDefinition struct {
	// Schema is the scheme of a type which has a list of enum values, eg, the
	// "container" of the enum.
	Schema Schema
	// TypeName is the name of the enum's type, usually aliased from something.
	TypeName string
	// ValueWrapper wraps the value. It's used to conditionally apply quotes
	// around strings.
	ValueWrapper string
	// PrefixTypeName determines if the enum value is prefixed with its TypeName.
	// This is set to true when this enum conflicts with another in terms of
	// TypeNames or when explicitly requested via the
	// `compatibility.always-prefix-enum-values` option.
	PrefixTypeName bool
}

// GetValues generates enum names in a way to minimize global conflicts
func (e *EnumDefinition) GetValues() map[string]string {
	// in case there are no conflicts, it's safe to use the values as-is
	if !e.PrefixTypeName {
		return e.Schema.EnumValues
	}
	// If we do have conflicts, we will prefix the enum's typename to the values.
	newValues := make(map[string]string, len(e.Schema.EnumValues))
	for k, v := range e.Schema.EnumValues {
		newName := e.TypeName + UppercaseFirstCharacter(k)
		newValues[newName] = v
	}
	return newValues
}

type Constants struct {
	// SecuritySchemeProviderNames holds all provider names for security schemes.
	SecuritySchemeProviderNames []string
	// EnumDefinitions holds type and value information for all enums
	EnumDefinitions []EnumDefinition
}

// TypeDefinition describes a Go type definition in generated code.
//
// Let's use this example schema:
// components:
//
//	schemas:
//	  Person:
//	    type: object
//	    properties:
//	    name:
//	      type: string
type TypeDefinition struct {
	// The name of the type, eg, type <...> Person
	TypeName string

	// The name of the corresponding JSON description, as it will sometimes
	// differ due to invalid characters.
	JsonName string

	// This is the Schema wrapper is used to populate the type description
	Schema Schema
}

// ResponseTypeDefinition is an extension of TypeDefinition, specifically for
// response unmarshaling in ClientWithResponses.
type ResponseTypeDefinition struct {
	TypeDefinition
	// The content type name where this is used, eg, application/json
	ContentTypeName string

	// The type name of a response model.
	ResponseName string

	AdditionalTypeDefinitions []TypeDefinition
}

func (t *TypeDefinition) IsAlias() bool {
	return !globalState.options.Compatibility.OldAliasing && t.Schema.DefineViaAlias
}

type Discriminator struct {
	// maps discriminator value to go type
	Mapping map[string]string

	// JSON property name that holds the discriminator
	Property string
}

func (d *Discriminator) JSONTag() string {
	return fmt.Sprintf("`json:\"%s\"`", d.Property)
}

func (d *Discriminator) PropertyName() string {
	return SchemaNameToTypeName(d.Property)
}

// UnionElement describe union element, based on prefix externalRef\d+ and real ref name from external schema.
type UnionElement string

// String returns externalRef\d+ and real ref name from external schema, like externalRef0.SomeType.
func (u UnionElement) String() string {
	return string(u)
}

// Method generate union method name for template functions `As/From/Merge`.
func (u UnionElement) Method() string {
	var method string
	for _, part := range strings.Split(string(u), `.`) {
		method += UppercaseFirstCharacter(part)
	}
	return method
}

func PropertiesEqual(a, b Property) bool {
	return a.JsonFieldName == b.JsonFieldName && a.Schema.TypeDecl() == b.Schema.TypeDecl() && a.Required == b.Required
}

func GenerateGoSchema(sref *openapi3.SchemaRef, path []string) (Schema, error) {
	// Add a fallback value in case the sref is nil.
	// i.e. the parent schema defines a type:array, but the array has
	// no items defined. Therefore, we have at least valid Go-Code.
	if sref == nil {
		return Schema{GoType: "interface{}"}, nil
	}

	schema := sref.Value

	// If Ref is set on the SchemaRef, it means that this type is actually a reference to
	// another type. We're not de-referencing, so simply use the referenced type.
	if IsGoTypeReference(sref.Ref) {
		// Convert the reference path to Go type
		refType, err := RefPathToGoType(sref.Ref)
		if err != nil {
			return Schema{}, fmt.Errorf("error turning reference (%s) into a Go type: %s",
				sref.Ref, err)
		}
		return Schema{
			GoType:         refType,
			Description:    schema.Description,
			DefineViaAlias: true,
			OAPISchema:     schema,
		}, nil
	}

	outSchema := Schema{
		Description: schema.Description,
		OAPISchema:  schema,
	}

	// AllOf is interesting, and useful. It's the union of a number of other
	// schemas. A common usage is to create a union of an object with an ID,
	// so that in a RESTful paradigm, the Create operation can return
	// (object, id), so that other operations can refer to (id)
	if schema.AllOf != nil {
		mergedSchema, err := MergeSchemas(schema.AllOf, path)
		if err != nil {
			return Schema{}, fmt.Errorf("error merging schemas: %w", err)
		}
		mergedSchema.OAPISchema = schema
		return mergedSchema, nil
	}

	// Check x-go-type, which will completely override the definition of this
	// schema with the provided type.
	if extension, ok := schema.Extensions[extPropGoType]; ok {
		typeName, err := extTypeName(extension)
		if err != nil {
			return outSchema, fmt.Errorf("invalid value for %q: %w", extPropGoType, err)
		}
		outSchema.GoType = typeName
		outSchema.DefineViaAlias = true

		return outSchema, nil
	}

	// Check x-go-type-skip-optional-pointer, which will override if the type
	// should be a pointer or not when the field is optional.
	if extension, ok := schema.Extensions[extPropGoTypeSkipOptionalPointer]; ok {
		skipOptionalPointer, err := extParsePropGoTypeSkipOptionalPointer(extension)
		if err != nil {
			return outSchema, fmt.Errorf("invalid value for %q: %w", extPropGoTypeSkipOptionalPointer, err)
		}
		outSchema.SkipOptionalPointer = skipOptionalPointer
	}

	// Schema type and format, eg. string / binary
	t := schema.Type
	// Handle objects and empty schemas first as a special case
	if t.Slice() == nil || t.Is("object") {
		var outType string

		if len(schema.Properties) == 0 && !SchemaHasAdditionalProperties(schema) && schema.AnyOf == nil && schema.OneOf == nil {
			// If the object has no properties or additional properties, we
			// have some special cases for its type.
			if t.Is("object") {
				// We have an object with no properties. This is a generic object
				// expressed as a map.
				outType = "map[string]interface{}"
			} else { // t == ""
				// If we don't even have the object designator, we're a completely
				// generic type.
				outType = "interface{}"
			}
			outSchema.GoType = outType
			outSchema.DefineViaAlias = true
		} else {
			// When we define an object, we want it to be a type definition,
			// not a type alias, eg, "type Foo struct {...}"
			outSchema.DefineViaAlias = false

			// If the schema has additional properties, we need to special case
			// a lot of behaviors.
			outSchema.HasAdditionalProperties = SchemaHasAdditionalProperties(schema)

			// Until we have a concrete additional properties type, we default to
			// any schema.
			outSchema.AdditionalPropertiesType = &Schema{
				GoType: "interface{}",
			}

			// If additional properties are defined, we will override the default
			// above with the specific definition.
			if schema.AdditionalProperties.Schema != nil {
				additionalSchema, err := GenerateGoSchema(schema.AdditionalProperties.Schema, path)
				if err != nil {
					return Schema{}, fmt.Errorf("error generating type for additional properties: %w", err)
				}
				if additionalSchema.HasAdditionalProperties || len(additionalSchema.UnionElements) != 0 {
					// If we have fields present which have additional properties or union values,
					// but are not a pre-defined type, we need to define a type
					// for them, which will be based on the field names we followed
					// to get to the type.
					typeName := PathToTypeName(append(path, "AdditionalProperties"))

					typeDef := TypeDefinition{
						TypeName: typeName,
						JsonName: strings.Join(append(path, "AdditionalProperties"), "."),
						Schema:   additionalSchema,
					}
					additionalSchema.RefType = typeName
					additionalSchema.AdditionalTypes = append(additionalSchema.AdditionalTypes, typeDef)
				}
				outSchema.AdditionalPropertiesType = &additionalSchema
				outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, additionalSchema.AdditionalTypes...)
			}

			// If the schema has no properties, and only additional properties, we will
			// early-out here and generate a map[string]<schema> instead of an object
			// that contains this map. We skip over anyOf/oneOf here because they can
			// introduce properties. allOf was handled above.
			if !globalState.options.Compatibility.DisableFlattenAdditionalProperties &&
				len(schema.Properties) == 0 && schema.AnyOf == nil && schema.OneOf == nil {
				// We have a dictionary here. Returns the goType to be just a map from
				// string to the property type. HasAdditionalProperties=false means
				// that we won't generate custom json.Marshaler and json.Unmarshaler functions,
				// since we don't need them for a simple map.
				outSchema.HasAdditionalProperties = false
				outSchema.GoType = fmt.Sprintf("map[string]%s", additionalPropertiesType(outSchema))
				return outSchema, nil
			}

			// We've got an object with some properties.
			for _, pName := range SortedSchemaKeys(schema.Properties) {
				p := schema.Properties[pName]
				propertyPath := append(path, pName)
				pSchema, err := GenerateGoSchema(p, propertyPath)
				if err != nil {
					return Schema{}, fmt.Errorf("error generating Go schema for property '%s': %w", pName, err)
				}

				required := StringInArray(pName, schema.Required)

				if (pSchema.HasAdditionalProperties || len(pSchema.UnionElements) != 0) && pSchema.RefType == "" {
					// If we have fields present which have additional properties or union values,
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
					ReadOnly:      p.Value.ReadOnly,
					WriteOnly:     p.Value.WriteOnly,
					Extensions:    p.Value.Extensions,
					Deprecated:    p.Value.Deprecated,
				}
				outSchema.Properties = append(outSchema.Properties, prop)
				if len(pSchema.AdditionalTypes) > 0 {
					outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, pSchema.AdditionalTypes...)
				}
			}

			if schema.AnyOf != nil {
				if err := generateUnion(&outSchema, schema.AnyOf, schema.Discriminator, path); err != nil {
					return Schema{}, fmt.Errorf("error generating type for anyOf: %w", err)
				}
			}
			if schema.OneOf != nil {
				if err := generateUnion(&outSchema, schema.OneOf, schema.Discriminator, path); err != nil {
					return Schema{}, fmt.Errorf("error generating type for oneOf: %w", err)
				}
			}

			outSchema.GoType = GenStructFromSchema(outSchema)
		}

		// Check for x-go-type-name. It behaves much like x-go-type, however, it will
		// create a type definition for the named type, and use the named type in place
		// of this schema.
		if extension, ok := schema.Extensions[extGoTypeName]; ok {
			typeName, err := extTypeName(extension)
			if err != nil {
				return outSchema, fmt.Errorf("invalid value for %q: %w", extGoTypeName, err)
			}

			newTypeDef := TypeDefinition{
				TypeName: typeName,
				Schema:   outSchema,
			}
			outSchema = Schema{
				Description:     newTypeDef.Schema.Description,
				GoType:          typeName,
				DefineViaAlias:  true,
				AdditionalTypes: append(outSchema.AdditionalTypes, newTypeDef),
			}
		}

		return outSchema, nil
	} else if len(schema.Enum) > 0 {
		err := oapiSchemaToGoType(schema, path, &outSchema)
		// Enums need to be typed, so that the values aren't interchangeable,
		// so no matter what schema conversion thinks, we need to define a
		// new type.
		outSchema.DefineViaAlias = false

		if err != nil {
			return Schema{}, fmt.Errorf("error resolving primitive type: %w", err)
		}
		enumValues := make([]string, len(schema.Enum))
		for i, enumValue := range schema.Enum {
			enumValues[i] = fmt.Sprintf("%v", enumValue)
		}

		enumNames := enumValues
		for _, key := range []string{extEnumVarNames, extEnumNames} {
			if extension, ok := schema.Extensions[key]; ok {
				if extEnumNames, err := extParseEnumVarNames(extension); err == nil {
					enumNames = extEnumNames
					break
				}
			}
		}

		sanitizedValues := SanitizeEnumNames(enumNames, enumValues)
		outSchema.EnumValues = make(map[string]string, len(sanitizedValues))

		for k, v := range sanitizedValues {
			var enumName string
			if v == "" {
				enumName = "Empty"
			} else {
				enumName = k
			}
			if globalState.options.Compatibility.OldEnumConflicts {
				outSchema.EnumValues[SchemaNameToTypeName(PathToTypeName(append(path, enumName)))] = v
			} else {
				outSchema.EnumValues[SchemaNameToTypeName(k)] = v
			}
		}
		if len(path) > 1 { // handle additional type only on non-toplevel types
			// Allow overriding autogenerated enum type names, since these may
			// cause conflicts - see https://github.com/oapi-codegen/oapi-codegen/issues/832
			var typeName string
			if extension, ok := schema.Extensions[extGoTypeName]; ok {
				typeName, err = extString(extension)
				if err != nil {
					return outSchema, fmt.Errorf("invalid value for %q: %w", extGoTypeName, err)
				}
			} else {
				typeName = SchemaNameToTypeName(PathToTypeName(path))
			}

			typeDef := TypeDefinition{
				TypeName: typeName,
				JsonName: strings.Join(path, "."),
				Schema:   outSchema,
			}
			outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, typeDef)
			outSchema.RefType = typeName
		}
	} else {
		err := oapiSchemaToGoType(schema, path, &outSchema)
		if err != nil {
			return Schema{}, fmt.Errorf("error resolving primitive type: %w", err)
		}
	}
	return outSchema, nil
}

// oapiSchemaToGoType converts an OpenApi schema into a Go type definition for
// all non-object types.
func oapiSchemaToGoType(schema *openapi3.Schema, path []string, outSchema *Schema) error {
	f := schema.Format
	t := schema.Type

	if t.Is("array") {
		// For arrays, we'll get the type of the Items and throw a
		// [] in front of it.
		arrayType, err := GenerateGoSchema(schema.Items, path)
		if err != nil {
			return fmt.Errorf("error generating type for array: %w", err)
		}
		if (arrayType.HasAdditionalProperties || len(arrayType.UnionElements) != 0) && arrayType.RefType == "" {
			// If we have items which have additional properties or union values,
			// but are not a pre-defined type, we need to define a type
			// for them, which will be based on the field names we followed
			// to get to the type.
			typeName := PathToTypeName(append(path, "Item"))

			typeDef := TypeDefinition{
				TypeName: typeName,
				JsonName: strings.Join(append(path, "Item"), "."),
				Schema:   arrayType,
			}
			arrayType.AdditionalTypes = append(arrayType.AdditionalTypes, typeDef)

			arrayType.RefType = typeName
		}
		outSchema.ArrayType = &arrayType
		outSchema.GoType = "[]" + arrayType.TypeDecl()
		outSchema.AdditionalTypes = arrayType.AdditionalTypes
		outSchema.Properties = arrayType.Properties
		outSchema.DefineViaAlias = true
		if sliceContains(globalState.options.OutputOptions.DisableTypeAliasesForType, "array") {
			outSchema.DefineViaAlias = false
		}

	} else if t.Is("integer") {
		// We default to int if format doesn't ask for something else.
		if f == "int64" {
			outSchema.GoType = "int64"
		} else if f == "int32" {
			outSchema.GoType = "int32"
		} else if f == "int16" {
			outSchema.GoType = "int16"
		} else if f == "int8" {
			outSchema.GoType = "int8"
		} else if f == "int" {
			outSchema.GoType = "int"
		} else if f == "uint64" {
			outSchema.GoType = "uint64"
		} else if f == "uint32" {
			outSchema.GoType = "uint32"
		} else if f == "uint16" {
			outSchema.GoType = "uint16"
		} else if f == "uint8" {
			outSchema.GoType = "uint8"
		} else if f == "uint" {
			outSchema.GoType = "uint"
		} else {
			outSchema.GoType = "int"
		}
		outSchema.DefineViaAlias = true
	} else if t.Is("number") {
		// We default to float for "number"
		if f == "double" {
			outSchema.GoType = "float64"
		} else if f == "float" || f == "" {
			outSchema.GoType = "float32"
		} else {
			return fmt.Errorf("invalid number format: %s", f)
		}
		outSchema.DefineViaAlias = true
	} else if t.Is("boolean") {
		if f != "" {
			return fmt.Errorf("invalid format (%s) for boolean", f)
		}
		outSchema.GoType = "bool"
		outSchema.DefineViaAlias = true
	} else if t.Is("string") {
		// Special case string formats here.
		switch f {
		case "byte":
			outSchema.GoType = "[]byte"
		case "email":
			outSchema.GoType = "openapi_types.Email"
		case "date":
			outSchema.GoType = "openapi_types.Date"
		case "date-time":
			outSchema.GoType = "time.Time"
		case "json":
			outSchema.GoType = "json.RawMessage"
			outSchema.SkipOptionalPointer = true
		case "uuid":
			outSchema.GoType = "openapi_types.UUID"
		case "binary":
			outSchema.GoType = "openapi_types.File"
		default:
			// All unrecognized formats are simply a regular string.
			outSchema.GoType = "string"
		}
		outSchema.DefineViaAlias = true
	} else {
		return fmt.Errorf("unhandled Schema type: %v", t)
	}
	return nil
}

// SchemaDescriptor describes a Schema, a type definition.
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

// GenFieldsFromProperties produce corresponding field names with JSON annotations,
// given a list of schema descriptors
func GenFieldsFromProperties(props []Property) []string {
	var fields []string
	for i, p := range props {
		field := ""

		goFieldName := p.GoFieldName()

		// Add a comment to a field in case we have one, otherwise skip.
		if p.Description != "" {
			// Separate the comment from a previous-defined, unrelated field.
			// Make sure the actual field is separated by a newline.
			if i != 0 {
				field += "\n"
			}
			field += fmt.Sprintf("%s\n", StringWithTypeNameToGoComment(p.Description, p.GoFieldName()))
		}

		if p.Deprecated {
			// This comment has to be on its own line for godoc & IDEs to pick up
			var deprecationReason string
			if extension, ok := p.Extensions[extDeprecationReason]; ok {
				if extOmitEmpty, err := extParseDeprecationReason(extension); err == nil {
					deprecationReason = extOmitEmpty
				}
			}

			field += fmt.Sprintf("%s\n", DeprecationComment(deprecationReason))
		}

		// Check x-go-type-skip-optional-pointer, which will override if the type
		// should be a pointer or not when the field is optional.
		if extension, ok := p.Extensions[extPropGoTypeSkipOptionalPointer]; ok {
			if skipOptionalPointer, err := extParsePropGoTypeSkipOptionalPointer(extension); err == nil {
				p.Schema.SkipOptionalPointer = skipOptionalPointer
			}
		}

		field += fmt.Sprintf("    %s %s", goFieldName, p.GoTypeDef())

		shouldOmitEmpty := (!p.Required || p.ReadOnly || p.WriteOnly) &&
			(!p.Required || !p.ReadOnly || !globalState.options.Compatibility.DisableRequiredReadOnlyAsPointer)

		omitEmpty := !p.Nullable && shouldOmitEmpty

		if p.Nullable && globalState.options.OutputOptions.NullableType {
			omitEmpty = shouldOmitEmpty
		}

		// Support x-omitempty
		if extOmitEmptyValue, ok := p.Extensions[extPropOmitEmpty]; ok {
			if extOmitEmpty, err := extParseOmitEmpty(extOmitEmptyValue); err == nil {
				omitEmpty = extOmitEmpty
			}
		}

		fieldTags := make(map[string]string)

		if !omitEmpty {
			fieldTags["json"] = p.JsonFieldName
			if p.NeedsFormTag {
				fieldTags["form"] = p.JsonFieldName
			}
		} else {
			fieldTags["json"] = p.JsonFieldName + ",omitempty"
			if p.NeedsFormTag {
				fieldTags["form"] = p.JsonFieldName + ",omitempty"
			}
		}

		// Support x-go-json-ignore
		if extension, ok := p.Extensions[extPropGoJsonIgnore]; ok {
			if goJsonIgnore, err := extParseGoJsonIgnore(extension); err == nil && goJsonIgnore {
				fieldTags["json"] = "-"
			}
		}

		// Support x-oapi-codegen-extra-tags
		if extension, ok := p.Extensions[extPropExtraTags]; ok {
			if tags, err := extExtraTags(extension); err == nil {
				keys := SortedMapKeys(tags)
				for _, k := range keys {
					fieldTags[k] = tags[k]
				}
			}
		}
		// Convert the fieldTags map into Go field annotations.
		keys := SortedMapKeys(fieldTags)
		tags := make([]string, len(keys))
		for i, k := range keys {
			tags[i] = fmt.Sprintf(`%s:"%s"`, k, fieldTags[k])
		}
		field += "`" + strings.Join(tags, " ") + "`"
		fields = append(fields, field)
	}
	return fields
}

func additionalPropertiesType(schema Schema) string {
	addPropsType := schema.AdditionalPropertiesType.GoType
	if schema.AdditionalPropertiesType.RefType != "" {
		addPropsType = schema.AdditionalPropertiesType.RefType
	}
	if schema.AdditionalPropertiesType.OAPISchema != nil && schema.AdditionalPropertiesType.OAPISchema.Nullable {
		addPropsType = "*" + addPropsType
	}
	return addPropsType
}

func GenStructFromSchema(schema Schema) string {
	// Start out with struct {
	objectParts := []string{"struct {"}
	// Append all the field definitions
	objectParts = append(objectParts, GenFieldsFromProperties(schema.Properties)...)
	// Close the struct
	if schema.HasAdditionalProperties {
		objectParts = append(objectParts,
			fmt.Sprintf("AdditionalProperties map[string]%s `json:\"-\"`",
				additionalPropertiesType(schema)))
	}
	if len(schema.UnionElements) != 0 {
		objectParts = append(objectParts, "union json.RawMessage")
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n")
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
			GoType:      "string",
			Description: StringToGoComment(param.Description),
		}, nil
	}

	// Otherwise, look for application/json in there
	mt, found := param.Content["application/json"]
	if !found {
		// If we don't have json, it's a string
		return Schema{
			GoType:      "string",
			Description: StringToGoComment(param.Description),
		}, nil
	}

	// For json, we go through the standard schema mechanism
	return GenerateGoSchema(mt.Schema, path)
}

func generateUnion(outSchema *Schema, elements openapi3.SchemaRefs, discriminator *openapi3.Discriminator, path []string) error {
	if discriminator != nil {
		outSchema.Discriminator = &Discriminator{
			Property: discriminator.PropertyName,
			Mapping:  make(map[string]string),
		}
	}

	refToGoTypeMap := make(map[string]string)
	for i, element := range elements {
		elementPath := append(path, fmt.Sprint(i))
		elementSchema, err := GenerateGoSchema(element, elementPath)
		if err != nil {
			return err
		}

		if element.Ref == "" {
			elementName := SchemaNameToTypeName(PathToTypeName(elementPath))
			if elementSchema.TypeDecl() == elementName {
				elementSchema.GoType = elementName
			} else {
				td := TypeDefinition{Schema: elementSchema, TypeName: elementName}
				outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, td)
				elementSchema.GoType = td.TypeName
			}
			outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, elementSchema.AdditionalTypes...)
		} else {
			refToGoTypeMap[element.Ref] = elementSchema.GoType
		}

		if discriminator != nil {
			if len(discriminator.Mapping) != 0 && element.Ref == "" {
				return errors.New("ambiguous discriminator.mapping: please replace inlined object with $ref")
			}

			// Explicit mapping.
			var mapped bool
			for k, v := range discriminator.Mapping {
				if v == element.Ref {
					outSchema.Discriminator.Mapping[k] = elementSchema.GoType
					mapped = true
					break
				}
			}
			// Implicit mapping.
			if !mapped {
				outSchema.Discriminator.Mapping[RefPathToObjName(element.Ref)] = elementSchema.GoType
			}
		}
		outSchema.UnionElements = append(outSchema.UnionElements, UnionElement(elementSchema.GoType))
	}

	if (outSchema.Discriminator != nil) && len(outSchema.Discriminator.Mapping) != len(elements) {
		return errors.New("discriminator: not all schemas were mapped")
	}

	return nil
}
