package codegen

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/singleton"
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

	if singleton.GlobalState.Options.Compatibility.AllowUnexportedStructFieldNames {
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
	if singleton.GlobalState.Options.OutputOptions.NullableType && p.Nullable {
		return "nullable.Nullable[" + typeDef + "]"
	}
	if !p.Schema.SkipOptionalPointer &&
		(!p.Required || p.Nullable ||
			(p.ReadOnly && (!p.Required || !singleton.GlobalState.Options.Compatibility.DisableRequiredReadOnlyAsPointer)) ||
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
	return !singleton.GlobalState.Options.Compatibility.OldAliasing && t.Schema.DefineViaAlias
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
