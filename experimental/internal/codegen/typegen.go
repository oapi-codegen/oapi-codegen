package codegen

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// TypeGenerator converts OpenAPI schemas to Go type expressions.
// It tracks required imports and handles recursive type references.
type TypeGenerator struct {
	typeMapping    TypeMapping
	converter      *NameConverter
	importResolver *ImportResolver
	tagGenerator   *StructTagGenerator
	imports        map[string]string // path -> alias (empty string = no alias)

	// schemaIndex maps JSON pointer refs to their descriptors
	schemaIndex map[string]*SchemaDescriptor

	// requiredTemplates tracks which custom type templates are needed
	requiredTemplates map[string]bool
}

// NewTypeGenerator creates a TypeGenerator with the given configuration.
func NewTypeGenerator(typeMapping TypeMapping, converter *NameConverter, importResolver *ImportResolver, tagGenerator *StructTagGenerator) *TypeGenerator {
	return &TypeGenerator{
		typeMapping:       typeMapping,
		converter:         converter,
		importResolver:    importResolver,
		tagGenerator:      tagGenerator,
		imports:           make(map[string]string),
		schemaIndex:       make(map[string]*SchemaDescriptor),
		requiredTemplates: make(map[string]bool),
	}
}

// IndexSchemas builds a lookup table from JSON pointer to schema descriptor.
// This is called before generation to enable $ref resolution.
func (g *TypeGenerator) IndexSchemas(schemas []*SchemaDescriptor) {
	for _, s := range schemas {
		ref := s.Path.String()
		g.schemaIndex[ref] = s
	}
}

// AddImport records an import path needed by the generated code.
func (g *TypeGenerator) AddImport(path string) {
	if path != "" {
		g.imports[path] = ""
	}
}

// AddImportAlias records an import path with an alias.
func (g *TypeGenerator) AddImportAlias(path, alias string) {
	if path != "" {
		g.imports[path] = alias
	}
}

// AddJSONImport adds encoding/json import (used by marshal/unmarshal code).
func (g *TypeGenerator) AddJSONImport() {
	g.AddImport("encoding/json")
}

// AddJSONImports adds encoding/json and fmt imports (used by oneOf marshal/unmarshal code).
func (g *TypeGenerator) AddJSONImports() {
	g.AddImport("encoding/json")
	g.AddImport("fmt")
}

// Imports returns the collected imports as a map[path]alias.
func (g *TypeGenerator) Imports() map[string]string {
	return g.imports
}

// RequiredTemplates returns the set of template names needed for custom types.
func (g *TypeGenerator) RequiredTemplates() map[string]bool {
	return g.requiredTemplates
}

// addTemplate records that a custom type template is needed.
func (g *TypeGenerator) addTemplate(templateName string) {
	if templateName != "" {
		g.requiredTemplates[templateName] = true
	}
}

// GoTypeExpr returns the Go type expression for a schema descriptor.
// This handles references by looking up the target schema's name,
// and inline schemas by generating the appropriate Go type.
func (g *TypeGenerator) GoTypeExpr(desc *SchemaDescriptor) string {
	// Handle $ref - return the referenced type's name
	if desc.IsReference() {
		// Check for external reference first
		if desc.IsExternalReference() {
			return g.externalRefType(desc)
		}

		// Internal reference - look up in schema index
		if target, ok := g.schemaIndex[desc.Ref]; ok {
			return target.ShortName
		}
		// Fallback for unresolved references
		return "any"
	}

	return g.goTypeForSchema(desc.Schema, desc)
}

// externalRefType resolves an external reference to a qualified Go type.
// Returns "any" if the external ref cannot be resolved.
func (g *TypeGenerator) externalRefType(desc *SchemaDescriptor) string {
	filePath, internalPath := desc.ParseExternalRef()
	if filePath == "" {
		return "any"
	}

	// Look up import mapping
	if g.importResolver == nil {
		// No import resolver configured - can't resolve external refs
		return "any"
	}

	imp := g.importResolver.Resolve(filePath)
	if imp == nil {
		// External file not in import mapping
		return "any"
	}

	// Extract type name from internal path (e.g., #/components/schemas/Pet -> Pet)
	typeName := extractTypeNameFromRef(internalPath, g.converter)
	if typeName == "" {
		return "any"
	}

	// If alias is empty, it's the current package (marked with "-")
	if imp.Alias == "" {
		return typeName
	}

	// Add the import
	g.AddImportAlias(imp.Path, imp.Alias)

	// Return qualified type
	return imp.Alias + "." + typeName
}

// extractTypeNameFromRef extracts a Go type name from an internal ref path.
// e.g., "#/components/schemas/Pet" -> "Pet"
func extractTypeNameFromRef(ref string, converter *NameConverter) string {
	// Remove leading #/
	ref = strings.TrimPrefix(ref, "#/")
	parts := strings.Split(ref, "/")

	if len(parts) < 3 {
		return ""
	}

	// For #/components/schemas/TypeName, the type name is the last part
	// We assume external refs point to component schemas
	typeName := parts[len(parts)-1]
	return converter.ToTypeName(typeName)
}

// goTypeForSchema generates a Go type expression from an OpenAPI schema.
// The desc parameter provides context (path, parent) for complex types.
func (g *TypeGenerator) goTypeForSchema(schema *base.Schema, desc *SchemaDescriptor) string {
	if schema == nil {
		return "any"
	}

	// Handle composition types
	if len(schema.AllOf) > 0 {
		return g.allOfType(desc)
	}
	if len(schema.AnyOf) > 0 {
		return g.anyOfType(desc)
	}
	if len(schema.OneOf) > 0 {
		return g.oneOfType(desc)
	}

	// Get the primary type from the type array
	// OpenAPI 3.1 allows type to be an array like ["string", "null"]
	primaryType := getPrimaryType(schema)

	switch primaryType {
	case "object":
		return g.objectType(schema, desc)
	case "array":
		return g.arrayType(schema, desc)
	case "string":
		return g.stringType(schema)
	case "integer":
		return g.integerType(schema)
	case "number":
		return g.numberType(schema)
	case "boolean":
		return g.booleanType(schema)
	default:
		// Unknown or empty type - could be a free-form object
		if schema.Properties != nil && schema.Properties.Len() > 0 {
			return g.objectType(schema, desc)
		}
		return "any"
	}
}

// getPrimaryType extracts the primary (non-null) type from a schema.
// OpenAPI 3.1 supports type arrays like ["string", "null"] for nullable.
func getPrimaryType(schema *base.Schema) string {
	if len(schema.Type) == 0 {
		return ""
	}
	for _, t := range schema.Type {
		if t != "null" {
			return t
		}
	}
	return schema.Type[0]
}

// objectType generates the Go type for an object schema.
// Simple objects with only additionalProperties become maps.
// Objects with properties become named struct types.
func (g *TypeGenerator) objectType(schema *base.Schema, desc *SchemaDescriptor) string {
	hasProperties := schema.Properties != nil && schema.Properties.Len() > 0
	hasAdditionalProps := schema.AdditionalProperties != nil

	// Pure map case: no properties, only additionalProperties
	if !hasProperties && hasAdditionalProps {
		return g.mapType(schema, desc)
	}

	// Empty object (no properties, no additionalProperties)
	if !hasProperties && !hasAdditionalProps {
		return "map[string]any"
	}

	// Struct case: has properties (with or without additionalProperties)
	// Return the type name - actual struct definition is generated separately
	if desc != nil && desc.ShortName != "" {
		return desc.ShortName
	}
	return "any"
}

// mapType generates a map[string]T type for additionalProperties schemas.
func (g *TypeGenerator) mapType(schema *base.Schema, desc *SchemaDescriptor) string {
	if schema.AdditionalProperties == nil {
		return "map[string]any"
	}

	// additionalProperties can be a boolean or a schema
	// If it's a schema proxy (A), get the value type
	if schema.AdditionalProperties.A != nil {
		valueSchema := schema.AdditionalProperties.A.Schema()
		if valueSchema != nil {
			valueType := g.goTypeForSchema(valueSchema, nil)
			return "map[string]" + valueType
		}
	}

	// additionalProperties: true or just present
	return "map[string]any"
}

// arrayType generates a []T type for array schemas.
func (g *TypeGenerator) arrayType(schema *base.Schema, desc *SchemaDescriptor) string {
	if schema.Items == nil || schema.Items.A == nil {
		return "[]any"
	}

	// Check if items is a reference
	itemProxy := schema.Items.A
	if itemProxy.IsReference() {
		ref := itemProxy.GetReference()
		if target, ok := g.schemaIndex[ref]; ok {
			return "[]" + target.ShortName
		}
	}

	// Check if we have a descriptor for the items schema
	if desc != nil && desc.Items != nil && desc.Items.ShortName != "" {
		return "[]" + desc.Items.ShortName
	}

	// Inline items schema
	itemSchema := itemProxy.Schema()
	itemType := g.goTypeForSchema(itemSchema, nil)
	return "[]" + itemType
}

// stringType returns the Go type for a string schema.
func (g *TypeGenerator) stringType(schema *base.Schema) string {
	spec := g.typeMapping.String.Default
	if schema.Format != "" {
		if formatSpec, ok := g.typeMapping.String.Formats[schema.Format]; ok {
			spec = formatSpec
		}
	}

	g.AddImport(spec.Import)
	g.addTemplate(spec.Template)
	return spec.Type
}

// integerType returns the Go type for an integer schema.
func (g *TypeGenerator) integerType(schema *base.Schema) string {
	spec := g.typeMapping.Integer.Default
	if schema.Format != "" {
		if formatSpec, ok := g.typeMapping.Integer.Formats[schema.Format]; ok {
			spec = formatSpec
		}
	}

	g.AddImport(spec.Import)
	return spec.Type
}

// numberType returns the Go type for a number schema.
func (g *TypeGenerator) numberType(schema *base.Schema) string {
	spec := g.typeMapping.Number.Default
	if schema.Format != "" {
		if formatSpec, ok := g.typeMapping.Number.Formats[schema.Format]; ok {
			spec = formatSpec
		}
	}

	g.AddImport(spec.Import)
	return spec.Type
}

// booleanType returns the Go type for a boolean schema.
func (g *TypeGenerator) booleanType(schema *base.Schema) string {
	spec := g.typeMapping.Boolean.Default
	g.AddImport(spec.Import)
	return spec.Type
}

// allOfType returns the type name for an allOf composition.
// allOf is typically used for struct embedding/inheritance.
func (g *TypeGenerator) allOfType(desc *SchemaDescriptor) string {
	if desc != nil && desc.ShortName != "" {
		return desc.ShortName
	}
	return "any"
}

// anyOfType returns the type name for an anyOf composition.
func (g *TypeGenerator) anyOfType(desc *SchemaDescriptor) string {
	if desc != nil && desc.ShortName != "" {
		return desc.ShortName
	}
	return "any"
}

// oneOfType returns the type name for a oneOf composition.
func (g *TypeGenerator) oneOfType(desc *SchemaDescriptor) string {
	if desc != nil && desc.ShortName != "" {
		return desc.ShortName
	}
	return "any"
}

// StructField represents a field in a generated Go struct.
type StructField struct {
	Name       string // Go field name
	Type       string // Go type expression
	JSONName   string // Original JSON property name
	Required   bool   // Is this field required in the schema
	Nullable   bool   // Is this field nullable (type includes "null")
	Pointer    bool   // Should this be a pointer type
	OmitEmpty  bool   // Include omitempty in json tag
	Doc        string // Field documentation
	Default    string // Go literal for default value (empty if no default)
	IsStruct   bool   // True if this field is a struct type (for recursive ApplyDefaults)
}

// GenerateStructFields creates the list of struct fields for an object schema.
func (g *TypeGenerator) GenerateStructFields(desc *SchemaDescriptor) []StructField {
	schema := desc.Schema
	if schema == nil || schema.Properties == nil {
		return nil
	}

	// Build required set
	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}

	var fields []StructField

	for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
		propName := pair.Key()
		propProxy := pair.Value()

		field := StructField{
			Name:     g.converter.ToPropertyName(propName),
			JSONName: propName,
			Required: required[propName],
		}

		// Resolve the property schema
		var propType string
		var propSchema *base.Schema
		if propProxy.IsReference() {
			ref := propProxy.GetReference()
			// Check if this is an external reference
			if !strings.HasPrefix(ref, "#") && strings.Contains(ref, "#") {
				// External reference - use import mapping
				tempDesc := &SchemaDescriptor{Ref: ref}
				propType = g.externalRefType(tempDesc)
				field.IsStruct = true // external references are typically to struct types
			} else if target, ok := g.schemaIndex[ref]; ok {
				propType = target.ShortName
				field.IsStruct = true // references are typically to struct types
			} else {
				propType = "any"
			}
		} else {
			propSchema = propProxy.Schema()
			field.Nullable = isNullable(propSchema)
			field.Doc = extractDescription(propSchema)

			// Generate the Go type for this property
			// Always use goTypeForSchema to get the correct type expression
			// This handles arrays, maps, and primitive types correctly
			propType = g.goTypeForSchema(propSchema, desc.Properties[propName])

			// Check if this is a struct type (object with properties, or a named type)
			if propSchema != nil {
				if propSchema.Properties != nil && propSchema.Properties.Len() > 0 {
					field.IsStruct = true
				}
				// Extract default value
				if propSchema.Default != nil {
					field.Default = formatDefaultValue(propSchema.Default.Value, propType)
				}
			}
		}

		// Determine pointer semantics:
		// - Required and not nullable: value type
		// - Optional or nullable: pointer type
		// - But: slices and maps are never pointers
		isCollection := strings.HasPrefix(propType, "[]") || strings.HasPrefix(propType, "map[")
		field.Pointer = !isCollection && (!field.Required || field.Nullable)
		field.OmitEmpty = !field.Required

		if field.Pointer {
			field.Type = "*" + propType
		} else {
			field.Type = propType
		}

		fields = append(fields, field)
	}

	return fields
}

// isNullable checks if a schema allows null values.
// In OpenAPI 3.1, this is expressed as type: ["string", "null"]
// In OpenAPI 3.0, this is expressed as nullable: true
func isNullable(schema *base.Schema) bool {
	if schema == nil {
		return false
	}

	// OpenAPI 3.1 style: type array includes "null"
	for _, t := range schema.Type {
		if t == "null" {
			return true
		}
	}

	// OpenAPI 3.0 style: nullable extension (stored in extensions)
	// Note: libopenapi may expose this differently; check schema.Nullable if available
	return false
}

// extractDescription gets the description from a schema.
func extractDescription(schema *base.Schema) string {
	if schema == nil {
		return ""
	}
	return schema.Description
}

// formatDefaultValue converts an OpenAPI default value to a Go literal.
// goType is used to determine the correct format for the literal.
func formatDefaultValue(value any, goType string) string {
	if value == nil {
		return ""
	}

	// Strip pointer prefix for type matching
	baseType := strings.TrimPrefix(goType, "*")

	switch v := value.(type) {
	case string:
		// Check if the target type is not a string
		// YAML/JSON might parse "10" or "true" as strings
		switch baseType {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			// Return the string as-is if it looks like a number
			return v
		case "bool":
			// Return the string as-is if it looks like a bool
			if v == "true" || v == "false" {
				return v
			}
		case "float32", "float64":
			return v
		}
		// It's actually a string type - quote it
		return fmt.Sprintf("%q", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case float64:
		// JSON numbers are always float64
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			// It's a whole number
			if strings.HasPrefix(baseType, "int") || strings.HasPrefix(baseType, "uint") {
				return fmt.Sprintf("%d", int64(v))
			}
		}
		return fmt.Sprintf("%v", v)
	case int, int64:
		return fmt.Sprintf("%d", v)
	case []any:
		// Arrays - generate a slice literal
		// For now, return empty slice if complex
		if len(v) == 0 {
			return fmt.Sprintf("%s{}", goType)
		}
		// Complex array defaults would need recursive handling
		return ""
	case map[string]any:
		// Objects - for now, skip complex defaults
		if len(v) == 0 {
			return fmt.Sprintf("%s{}", goType)
		}
		return ""
	default:
		// Try a simple string conversion
		return fmt.Sprintf("%v", v)
	}
}

// HasAdditionalProperties returns true if the schema has explicit additionalProperties.
func (g *TypeGenerator) HasAdditionalProperties(desc *SchemaDescriptor) bool {
	if desc == nil || desc.Schema == nil {
		return false
	}
	return desc.Schema.AdditionalProperties != nil
}

// AdditionalPropertiesType returns the Go type for the additionalProperties.
func (g *TypeGenerator) AdditionalPropertiesType(desc *SchemaDescriptor) string {
	if desc == nil || desc.Schema == nil || desc.Schema.AdditionalProperties == nil {
		return "any"
	}

	if desc.Schema.AdditionalProperties.A != nil {
		valueSchema := desc.Schema.AdditionalProperties.A.Schema()
		if valueSchema != nil {
			return g.goTypeForSchema(valueSchema, nil)
		}
	}

	return "any"
}

// SchemaKind represents the kind of schema for code generation.
type SchemaKind int

const (
	KindStruct SchemaKind = iota
	KindMap
	KindAlias
	KindEnum
	KindAllOf
	KindAnyOf
	KindOneOf
	KindReference
)

// GetSchemaKind determines what kind of Go type to generate for a schema.
func GetSchemaKind(desc *SchemaDescriptor) SchemaKind {
	if desc.IsReference() {
		return KindReference
	}

	schema := desc.Schema
	if schema == nil {
		return KindAlias
	}

	// Enum check first
	if len(schema.Enum) > 0 {
		return KindEnum
	}

	// Composition types
	if len(schema.AllOf) > 0 {
		return KindAllOf
	}
	if len(schema.AnyOf) > 0 {
		return KindAnyOf
	}
	if len(schema.OneOf) > 0 {
		return KindOneOf
	}

	// Object with properties -> struct
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		return KindStruct
	}

	// Object with only additionalProperties -> map
	primaryType := getPrimaryType(schema)
	if primaryType == "object" {
		if schema.AdditionalProperties != nil {
			return KindMap
		}
		return KindStruct // empty struct
	}

	// Everything else is an alias to a primitive type
	return KindAlias
}

// FormatJSONTag generates a JSON struct tag for a field.
// Deprecated: Use StructTagGenerator instead.
func FormatJSONTag(jsonName string, omitEmpty bool) string {
	if omitEmpty {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", jsonName)
	}
	return fmt.Sprintf("`json:\"%s\"`", jsonName)
}

// GenerateFieldTag generates struct tags for a field using the configured templates.
func (g *TypeGenerator) GenerateFieldTag(field StructField) string {
	if g.tagGenerator == nil {
		// Fallback to legacy behavior
		return FormatJSONTag(field.JSONName, field.OmitEmpty)
	}

	info := StructTagInfo{
		FieldName:   field.JSONName,
		GoFieldName: field.Name,
		IsOptional:  !field.Required,
		IsNullable:  field.Nullable,
		IsPointer:   field.Pointer,
	}
	return g.tagGenerator.GenerateTags(info)
}

// TagGenerator returns the struct tag generator.
func (g *TypeGenerator) TagGenerator() *StructTagGenerator {
	return g.tagGenerator
}
