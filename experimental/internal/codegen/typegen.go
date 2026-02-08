package codegen

import (
	"fmt"
	"sort"
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
	ctx            *CodegenContext // centralized import & helper tracking

	// schemaIndex maps JSON pointer refs to their descriptors
	schemaIndex map[string]*SchemaDescriptor
}

// NewTypeGenerator creates a TypeGenerator with the given configuration.
func NewTypeGenerator(typeMapping TypeMapping, converter *NameConverter, importResolver *ImportResolver, tagGenerator *StructTagGenerator, ctx *CodegenContext) *TypeGenerator {
	return &TypeGenerator{
		typeMapping:    typeMapping,
		converter:      converter,
		importResolver: importResolver,
		tagGenerator:   tagGenerator,
		ctx:            ctx,
		schemaIndex:    make(map[string]*SchemaDescriptor),
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
	g.ctx.AddImport(path)
}

// AddImportAlias records an import path with an alias.
func (g *TypeGenerator) AddImportAlias(path, alias string) {
	g.ctx.AddImportAlias(path, alias)
}

// AddJSONImport adds encoding/json import (used by marshal/unmarshal code).
func (g *TypeGenerator) AddJSONImport() {
	g.ctx.AddJSONImport()
}

// AddJSONImports adds encoding/json and fmt imports (used by oneOf marshal/unmarshal code).
func (g *TypeGenerator) AddJSONImports() {
	g.ctx.AddJSONImports()
}

// AddNullableTemplate adds the nullable type template to the output.
func (g *TypeGenerator) AddNullableTemplate() {
	g.ctx.NeedCustomType("nullable")
}

// Imports returns the collected imports as a map[path]alias.
func (g *TypeGenerator) Imports() map[string]string {
	return g.ctx.Imports()
}

// RequiredTemplates returns the set of custom type template names needed.
func (g *TypeGenerator) RequiredTemplates() map[string]bool {
	// Build a map from the context's list for backward compatibility
	result := make(map[string]bool)
	for _, name := range g.ctx.RequiredCustomTypes() {
		result[name] = true
	}
	return result
}

// addTemplate records that a custom type template is needed.
func (g *TypeGenerator) addTemplate(templateName string) {
	g.ctx.NeedCustomType(templateName)
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

	// Check if this is a nullable primitive - wrap in Nullable[T]
	nullable := isNullable(schema)
	isPrimitive := primaryType == "string" || primaryType == "integer" || primaryType == "number" || primaryType == "boolean"

	var baseType string
	switch primaryType {
	case "object":
		return g.objectType(schema, desc)
	case "array":
		return g.arrayType(schema, desc)
	case "string":
		baseType = g.stringType(schema)
	case "integer":
		baseType = g.integerType(schema)
	case "number":
		baseType = g.numberType(schema)
	case "boolean":
		baseType = g.booleanType(schema)
	default:
		// Unknown or empty type - could be a free-form object
		if schema.Properties != nil && schema.Properties.Len() > 0 {
			return g.objectType(schema, desc)
		}
		return "any"
	}

	// Wrap nullable primitives in Nullable[T]
	if nullable && isPrimitive {
		g.AddNullableTemplate()
		return "Nullable[" + baseType + "]"
	}

	return baseType
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
		// Check for external reference first
		if !strings.HasPrefix(ref, "#") && strings.Contains(ref, "#") {
			// External reference - use import mapping
			tempDesc := &SchemaDescriptor{Ref: ref}
			itemType := g.externalRefType(tempDesc)
			return "[]" + itemType
		}
		// Internal reference - look up in schema index
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
	Name            string // Go field name
	Type            string // Go type expression
	JSONName        string // Original JSON property name
	Required        bool   // Is this field required in the schema
	Nullable        bool   // Is this field nullable (type includes "null")
	Pointer         bool   // Should this be a pointer type
	OmitEmpty       bool   // Include omitempty in json tag
	OmitZero        bool   // Include omitzero in json tag (Go 1.24+)
	JSONIgnore      bool   // Use json:"-" tag to exclude from marshaling
	Doc             string // Field documentation
	Default         string // Go literal for default value (empty if no default)
	IsStruct        bool   // True if this field is a struct type (for recursive ApplyDefaults)
	IsExternal      bool   // True if this field references an external type (ApplyDefaults via reflection)
	IsNullableAlias bool   // True if type is a type alias to Nullable[T] (don't wrap or pointer)
	Order           *int   // Optional field ordering (lower values come first)
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
	needsNullableImport := false

	for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
		propName := pair.Key()
		propProxy := pair.Value()

		field := StructField{
			Name:     g.converter.ToPropertyName(propName),
			JSONName: propName,
			Required: required[propName],
		}

		// Parse extensions from the property schema
		var propExtensions *Extensions
		var propSchema *base.Schema

		// Resolve the property schema
		var propType string
		if propProxy.IsReference() {
			ref := propProxy.GetReference()
			// Check if this is an external reference
			if !strings.HasPrefix(ref, "#") && strings.Contains(ref, "#") {
				// External reference - use import mapping
				tempDesc := &SchemaDescriptor{Ref: ref}
				propType = g.externalRefType(tempDesc)
				field.IsExternal = true // external references need reflection-based ApplyDefaults
			} else if target, ok := g.schemaIndex[ref]; ok {
				propType = target.ShortName
				// Only set IsStruct if the referenced schema has ApplyDefaults
				// This filters out array/map type aliases which don't have ApplyDefaults
				field.IsStruct = schemaHasApplyDefaults(target.Schema)
				// Check if the referenced schema is nullable
				// BUT: if it's a nullable primitive, the type alias already wraps Nullable[T],
				// so we shouldn't double-wrap it or add a pointer (Nullable handles "unspecified")
				if isNullablePrimitive(target.Schema) {
					// Already Nullable[T] - use as value type directly
					field.IsNullableAlias = true
				} else if isNullable(target.Schema) {
					field.Nullable = true
				}
				// Extensions from referenced schema apply to the field
				propExtensions = target.Extensions
			} else {
				propType = "any"
			}
		} else {
			propSchema = propProxy.Schema()
			field.Nullable = isNullable(propSchema)
			field.Doc = extractDescription(propSchema)

			// Parse extensions from the property schema
			if propSchema != nil && propSchema.Extensions != nil {
				ext, err := ParseExtensions(propSchema.Extensions, desc.Path.Append("properties", propName).String())
				if err == nil {
					propExtensions = ext
				}
			}

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

		// Apply extensions to the field
		if propExtensions != nil {
			// Name override
			if propExtensions.NameOverride != "" {
				field.Name = propExtensions.NameOverride
			}

			// Type override replaces the generated type entirely
			if propExtensions.TypeOverride != nil {
				propType = propExtensions.TypeOverride.TypeName
				if propExtensions.TypeOverride.ImportPath != "" {
					if propExtensions.TypeOverride.ImportAlias != "" {
						g.AddImportAlias(propExtensions.TypeOverride.ImportPath, propExtensions.TypeOverride.ImportAlias)
					} else {
						g.AddImport(propExtensions.TypeOverride.ImportPath)
					}
				}
				// Type override bypasses nullable wrapping - the user specifies the exact type
				field.IsNullableAlias = true // Don't wrap or add pointer
			}

			// JSON ignore
			if propExtensions.JSONIgnore != nil && *propExtensions.JSONIgnore {
				field.JSONIgnore = true
			}

			// Deprecated reason appended to documentation
			if propExtensions.DeprecatedReason != "" {
				if field.Doc != "" {
					field.Doc = field.Doc + "\nDeprecated: " + propExtensions.DeprecatedReason
				} else {
					field.Doc = "Deprecated: " + propExtensions.DeprecatedReason
				}
			}

			// Order for field sorting
			if propExtensions.Order != nil {
				field.Order = propExtensions.Order
			}
		}

		// Determine type semantics:
		// - Nullable fields: use Nullable[T]
		// - Optional (not nullable) fields: use *T (pointer)
		// - Required (not nullable) fields: use T (value type)
		// - Collections (slices/maps) are never wrapped
		// - Types already wrapped in Nullable[] are not double-wrapped
		// - Type aliases to Nullable[T] are used as-is (IsNullableAlias)
		isCollection := strings.HasPrefix(propType, "[]") || strings.HasPrefix(propType, "map[")
		alreadyNullable := strings.HasPrefix(propType, "Nullable[") || field.IsNullableAlias

		if field.Nullable && !isCollection && !alreadyNullable {
			// Use Nullable[T] for nullable fields (generated inline from template)
			field.Type = "Nullable[" + propType + "]"
			field.Pointer = false
			needsNullableImport = true
		} else if !field.Required && !isCollection && !alreadyNullable {
			// Check for skip optional pointer extension
			skipPointer := false
			if propExtensions != nil && propExtensions.SkipOptionalPointer != nil && *propExtensions.SkipOptionalPointer {
				skipPointer = true
			}

			if skipPointer {
				// Use value type even though optional
				field.Type = propType
				field.Pointer = false
			} else {
				// Use pointer for optional non-nullable fields
				field.Type = "*" + propType
				field.Pointer = true
			}
		} else {
			// Value type for required non-nullable fields, collections, and Nullable aliases
			field.Type = propType
			field.Pointer = false
		}

		// Determine omitempty/omitzero behavior
		field.OmitEmpty = !field.Required
		if propExtensions != nil {
			// Explicit omitempty override
			if propExtensions.OmitEmpty != nil {
				field.OmitEmpty = *propExtensions.OmitEmpty
			}
			// Explicit omitzero
			if propExtensions.OmitZero != nil && *propExtensions.OmitZero {
				field.OmitZero = true
			}
		}

		fields = append(fields, field)
	}

	if needsNullableImport {
		g.AddNullableTemplate()
	}

	// Sort fields by order if any have explicit ordering
	sortFieldsByOrder(fields)

	return fields
}

// collectFieldsRecursive returns the struct fields for a schema, recursively
// following allOf chains. For schemas with direct properties (no allOf), this
// falls through to GenerateStructFields. For allOf-composed schemas, it
// collects fields from all allOf members recursively, so that nested allOf
// references (e.g., A: allOf[$ref:B, ...] where B: allOf[$ref:C, ...]) are
// properly flattened.
func (g *TypeGenerator) collectFieldsRecursive(desc *SchemaDescriptor) []StructField {
	schema := desc.Schema
	if schema == nil {
		return nil
	}

	// If this schema has no allOf, use the standard field generation
	if len(schema.AllOf) == 0 {
		return g.GenerateStructFields(desc)
	}

	// Collect fields from direct properties first
	var fields []StructField
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		fields = append(fields, g.GenerateStructFields(desc)...)
	}

	// Recursively collect fields from each allOf member
	for i, proxy := range schema.AllOf {
		memberSchema := proxy.Schema()
		if memberSchema == nil {
			continue
		}

		var memberFields []StructField
		if proxy.IsReference() {
			ref := proxy.GetReference()
			if target, ok := g.schemaIndex[ref]; ok {
				// Recurse: the target may itself be an allOf composition
				memberFields = g.collectFieldsRecursive(target)
			}
		} else if memberSchema.Properties != nil && memberSchema.Properties.Len() > 0 {
			if desc.AllOf != nil && i < len(desc.AllOf) {
				memberFields = g.GenerateStructFields(desc.AllOf[i])
			}
		}

		// Apply required array from this allOf member to collected fields
		if len(memberSchema.Required) > 0 {
			reqSet := make(map[string]bool)
			for _, r := range memberSchema.Required {
				reqSet[r] = true
			}
			for j := range memberFields {
				if reqSet[memberFields[j].JSONName] && !memberFields[j].Required {
					memberFields[j].Required = true
					memberFields[j].OmitEmpty = false
					if !memberFields[j].Nullable && !strings.HasPrefix(memberFields[j].Type, "[]") && !strings.HasPrefix(memberFields[j].Type, "map[") {
						memberFields[j].Type = strings.TrimPrefix(memberFields[j].Type, "*")
						memberFields[j].Pointer = false
					}
				}
			}
		}

		fields = append(fields, memberFields...)
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

	// OpenAPI 3.0 style: nullable: true
	if schema.Nullable != nil && *schema.Nullable {
		return true
	}

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
		if field.JSONIgnore {
			return "`json:\"-\"`"
		}
		return FormatJSONTag(field.JSONName, field.OmitEmpty)
	}

	info := StructTagInfo{
		FieldName:   field.JSONName,
		GoFieldName: field.Name,
		IsOptional:  !field.Required,
		IsNullable:  field.Nullable,
		IsPointer:   field.Pointer,
		OmitEmpty:   field.OmitEmpty,
		OmitZero:    field.OmitZero,
		JSONIgnore:  field.JSONIgnore,
	}
	return g.tagGenerator.GenerateTags(info)
}

// TagGenerator returns the struct tag generator.
func (g *TypeGenerator) TagGenerator() *StructTagGenerator {
	return g.tagGenerator
}

// sortFieldsByOrder sorts fields by their Order value.
// Fields without an Order value are placed after fields with explicit ordering,
// maintaining their original relative order (stable sort).
func sortFieldsByOrder(fields []StructField) {
	// Check if any fields have explicit ordering
	hasOrder := false
	for _, f := range fields {
		if f.Order != nil {
			hasOrder = true
			break
		}
	}
	if !hasOrder {
		return
	}

	// Stable sort to preserve relative order of fields without explicit ordering
	sort.SliceStable(fields, func(i, j int) bool {
		// Fields with Order come before fields without
		if fields[i].Order == nil && fields[j].Order == nil {
			return false // preserve original order
		}
		if fields[i].Order == nil {
			return false // i (no order) comes after j
		}
		if fields[j].Order == nil {
			return true // i (has order) comes before j
		}
		// Both have order - sort by value
		return *fields[i].Order < *fields[j].Order
	})
}
