// Package codegen generates Go code from parsed OpenAPI specs.
package codegen

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/templates"
)

// Generate produces Go code from the parsed OpenAPI document.
func Generate(doc libopenapi.Document, cfg Configuration) (string, error) {
	cfg.ApplyDefaults()

	// Create content type matcher for filtering request/response bodies
	contentTypeMatcher := NewContentTypeMatcher(cfg.ContentTypes)

	// Create content type short namer for friendly type names
	contentTypeNamer := NewContentTypeShortNamer(cfg.ContentTypeShortNames)

	// Pass 1: Gather all schemas that need types
	schemas, err := GatherSchemas(doc, contentTypeMatcher)
	if err != nil {
		return "", fmt.Errorf("gathering schemas: %w", err)
	}

	// Pass 2: Compute names for all schemas
	converter := NewNameConverter(cfg.NameMangling, cfg.NameSubstitutions)
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	// Pass 3: Generate Go code
	importResolver := NewImportResolver(cfg.ImportMapping)
	tagGenerator := NewStructTagGenerator(cfg.StructTags)
	gen := NewTypeGenerator(cfg.TypeMapping, converter, importResolver, tagGenerator)
	gen.IndexSchemas(schemas)

	output := NewOutput(cfg.PackageName)
	// Note: encoding/json and fmt imports are added by generateType when needed

	// Generate models (types for schemas)
	if !cfg.Generation.NoModels {
		for _, desc := range schemas {
			code := generateType(gen, desc)
			if code != "" {
				output.AddType(code)
			}
		}

		// Add imports collected during generation
		output.AddImports(gen.Imports())

		// Add custom type templates (Date, Email, UUID, File, etc.)
		for templateName := range gen.RequiredTemplates() {
			typeCode, typeImports := loadCustomType(templateName)
			if typeCode != "" {
				output.AddType(typeCode)
				for path, alias := range typeImports {
					output.AddImport(path, alias)
				}
			}
		}
	}

	return output.Format()
}

// generateType generates Go code for a single schema descriptor.
func generateType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	kind := GetSchemaKind(desc)

	var code string
	switch kind {
	case KindReference:
		// References don't generate new types; they use the referenced type's name
		return ""

	case KindStruct:
		code = generateStructType(gen, desc)

	case KindMap:
		code = generateMapAlias(gen, desc)

	case KindEnum:
		code = generateEnumType(gen, desc)

	case KindAllOf:
		code = generateAllOfType(gen, desc)

	case KindAnyOf:
		code = generateAnyOfType(gen, desc)

	case KindOneOf:
		code = generateOneOfType(gen, desc)

	case KindAlias:
		code = generateTypeAlias(gen, desc)

	default:
		return ""
	}

	if code == "" {
		return ""
	}

	// Prepend schema path comment
	return schemaPathComment(desc.Path) + code
}

// schemaPathComment returns a comment line showing the schema path.
func schemaPathComment(path SchemaPath) string {
	return fmt.Sprintf("// %s\n", path.String())
}

// generateStructType generates a struct type for an object schema.
func generateStructType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	fields := gen.GenerateStructFields(desc)
	doc := extractDescription(desc.Schema)

	// Check if we need additionalProperties handling
	if gen.HasAdditionalProperties(desc) {
		// Mixed properties need encoding/json and fmt for marshal/unmarshal
		gen.AddJSONImports()

		addPropsType := gen.AdditionalPropertiesType(desc)
		structCode := GenerateStructWithAdditionalProps(desc.StableName, fields, addPropsType, doc, gen.TagGenerator())

		// Generate marshal/unmarshal methods
		marshalCode := GenerateMixedPropertiesMarshal(desc.StableName, fields)
		unmarshalCode := GenerateMixedPropertiesUnmarshal(desc.StableName, fields, addPropsType)

		code := structCode + "\n" + marshalCode + "\n" + unmarshalCode

		// If there's a short name different from stable name, generate an alias
		if desc.ShortName != desc.StableName {
			code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
		}

		// Generate ApplyDefaults method if needed
		if applyDefaults := GenerateApplyDefaults(desc.StableName, fields); applyDefaults != "" {
			code += "\n" + applyDefaults
		}

		return code
	}

	code := GenerateStruct(desc.StableName, fields, doc, gen.TagGenerator())

	// If there's a short name different from stable name, generate an alias
	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}

	// Generate ApplyDefaults method if needed
	if applyDefaults := GenerateApplyDefaults(desc.StableName, fields); applyDefaults != "" {
		code += "\n" + applyDefaults
	}

	return code
}

// generateMapAlias generates a type alias for a pure map schema.
func generateMapAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	mapType := gen.GoTypeExpr(desc)
	doc := extractDescription(desc.Schema)
	code := GenerateTypeAlias(desc.StableName, mapType, doc)

	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}
	return code
}

// generateEnumType generates an enum type with const values.
func generateEnumType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	schema := desc.Schema
	if schema == nil {
		return ""
	}

	// Determine base type
	baseType := "string"
	primaryType := getPrimaryType(schema)
	if primaryType == "integer" {
		baseType = "int"
	}

	// Extract enum values as strings
	var values []string
	for _, v := range schema.Enum {
		values = append(values, fmt.Sprintf("%v", v.Value))
	}

	doc := extractDescription(schema)
	code := GenerateEnum(desc.StableName, baseType, values, doc)

	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}
	return code
}

// generateTypeAlias generates a simple type alias.
func generateTypeAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	goType := gen.GoTypeExpr(desc)
	doc := extractDescription(desc.Schema)
	code := GenerateTypeAlias(desc.StableName, goType, doc)

	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}
	return code
}

// AllOfMergeError represents a conflict when merging allOf schemas.
type AllOfMergeError struct {
	SchemaName   string
	PropertyName string
	Type1        string
	Type2        string
}

func (e AllOfMergeError) Error() string {
	return fmt.Sprintf("allOf merge conflict in %s: property %q has conflicting types %s and %s",
		e.SchemaName, e.PropertyName, e.Type1, e.Type2)
}

// allOfMemberInfo holds information about an allOf member for merging.
type allOfMemberInfo struct {
	fields    []StructField // flattened fields from object schemas
	unionType string        // non-empty if this member is a oneOf/anyOf union
	unionDesc *SchemaDescriptor
	required  []string // required fields from this allOf member
}

// generateAllOfType generates a struct with flattened properties from all allOf members.
// Object schema properties are merged into flat fields.
// oneOf/anyOf members become union fields with json:"-" tag.
func generateAllOfType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	schema := desc.Schema
	if schema == nil {
		return ""
	}

	// Merge all fields, checking for conflicts
	mergedFields := make(map[string]StructField) // keyed by JSONName
	var fieldOrder []string                       // preserve order
	var unionFields []StructField

	// First, collect fields from properties defined directly on the schema
	// (Issue 2102: properties at same level as allOf were being ignored)
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		directFields := gen.GenerateStructFields(desc)
		for _, field := range directFields {
			mergedFields[field.JSONName] = field
			fieldOrder = append(fieldOrder, field.JSONName)
		}
	}

	// Collect info about each allOf member
	var members []allOfMemberInfo
	for i, proxy := range schema.AllOf {
		info := allOfMemberInfo{}

		memberSchema := proxy.Schema()
		if memberSchema == nil {
			continue
		}

		// Check if this member is a oneOf/anyOf (union type)
		if len(memberSchema.OneOf) > 0 || len(memberSchema.AnyOf) > 0 {
			// This is a union - keep as a union field
			if desc.AllOf != nil && i < len(desc.AllOf) {
				info.unionType = desc.AllOf[i].ShortName
				info.unionDesc = desc.AllOf[i]
			}
		} else if proxy.IsReference() {
			// Reference to another schema - get its fields
			ref := proxy.GetReference()
			if target, ok := gen.schemaIndex[ref]; ok {
				info.fields = gen.GenerateStructFields(target)
			}
		} else if memberSchema.Properties != nil && memberSchema.Properties.Len() > 0 {
			// Inline object schema - get its fields
			if desc.AllOf != nil && i < len(desc.AllOf) {
				info.fields = gen.GenerateStructFields(desc.AllOf[i])
			}
		}

		// Also check for required array in allOf members (may mark fields as required)
		info.required = memberSchema.Required

		members = append(members, info)
	}

	// Merge fields from allOf members
	for _, member := range members {
		if member.unionType != "" {
			// Add union as a special field
			unionFields = append(unionFields, StructField{
				Name:     member.unionType,
				Type:     "*" + member.unionType,
				JSONName: "-", // will use json:"-"
			})
			continue
		}

		for _, field := range member.fields {
			if existing, ok := mergedFields[field.JSONName]; ok {
				// Check for type conflict
				if existing.Type != field.Type {
					// Type conflict - generate error comment in output
					// In a real implementation, this should be a proper error
					// For now, we'll include a comment and use the first type
					field.Doc = fmt.Sprintf("CONFLICT: type %s vs %s", existing.Type, field.Type)
				}
				// If same type, keep existing (first wins for required/nullable)
				continue
			}
			mergedFields[field.JSONName] = field
			fieldOrder = append(fieldOrder, field.JSONName)
		}

		// Apply required array from this allOf member to update pointer/omitempty
		for _, reqName := range member.required {
			if field, ok := mergedFields[reqName]; ok {
				if !field.Required {
					field.Required = true
					field.OmitEmpty = false
					// Update pointer status - required non-nullable fields are not pointers
					if !field.Nullable && !strings.HasPrefix(field.Type, "[]") && !strings.HasPrefix(field.Type, "map[") {
						field.Type = strings.TrimPrefix(field.Type, "*")
						field.Pointer = false
					}
					mergedFields[reqName] = field
				}
			}
		}
	}

	// Build final field list in order
	var finalFields []StructField
	for _, jsonName := range fieldOrder {
		finalFields = append(finalFields, mergedFields[jsonName])
	}

	doc := extractDescription(schema)

	// Generate struct
	var code string
	if len(unionFields) > 0 {
		// Has union members - need custom marshal/unmarshal
		gen.AddJSONImport()
		code = generateAllOfStructWithUnions(desc.StableName, finalFields, unionFields, doc, gen.TagGenerator())
	} else {
		// Simple case - just flattened fields
		code = GenerateStruct(desc.StableName, finalFields, doc, gen.TagGenerator())
	}

	// If there's a short name different from stable name, generate an alias
	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}

	// Generate ApplyDefaults method if needed
	if applyDefaults := GenerateApplyDefaults(desc.StableName, finalFields); applyDefaults != "" {
		code += "\n" + applyDefaults
	}

	return code
}

// generateAllOfStructWithUnions generates an allOf struct that contains union fields.
func generateAllOfStructWithUnions(name string, fields []StructField, unionFields []StructField, doc string, tagGen *StructTagGenerator) string {
	b := NewCodeBuilder()

	if doc != "" {
		for _, line := range strings.Split(doc, "\n") {
			b.Line("// %s", line)
		}
	}

	b.Line("type %s struct {", name)
	b.Indent()

	// Regular fields
	for _, f := range fields {
		tag := generateFieldTag(f, tagGen)
		b.Line("%s %s %s", f.Name, f.Type, tag)
	}

	// Union fields with json:"-"
	for _, f := range unionFields {
		b.Line("%s %s `json:\"-\"`", f.Name, f.Type)
	}

	b.Dedent()
	b.Line("}")

	// Generate MarshalJSON
	b.BlankLine()
	b.Line("func (s %s) MarshalJSON() ([]byte, error) {", name)
	b.Indent()
	b.Line("result := make(map[string]any)")
	b.BlankLine()

	// Marshal regular fields
	for _, f := range fields {
		if f.Pointer {
			b.Line("if s.%s != nil {", f.Name)
			b.Indent()
			b.Line("result[%q] = s.%s", f.JSONName, f.Name)
			b.Dedent()
			b.Line("}")
		} else if strings.HasPrefix(f.Type, "[]") || strings.HasPrefix(f.Type, "map[") {
			// Slices and maps - only include if not nil
			b.Line("if s.%s != nil {", f.Name)
			b.Indent()
			b.Line("result[%q] = s.%s", f.JSONName, f.Name)
			b.Dedent()
			b.Line("}")
		} else {
			b.Line("result[%q] = s.%s", f.JSONName, f.Name)
		}
	}

	// Marshal and merge union fields
	for _, f := range unionFields {
		b.BlankLine()
		b.Line("if s.%s != nil {", f.Name)
		b.Indent()
		b.Line("unionData, err := json.Marshal(s.%s)", f.Name)
		b.Line("if err != nil {")
		b.Indent()
		b.Line("return nil, err")
		b.Dedent()
		b.Line("}")
		b.Line("var unionMap map[string]any")
		b.Line("if err := json.Unmarshal(unionData, &unionMap); err == nil {")
		b.Indent()
		b.Line("for k, v := range unionMap {")
		b.Indent()
		b.Line("result[k] = v")
		b.Dedent()
		b.Line("}")
		b.Dedent()
		b.Line("}")
		b.Dedent()
		b.Line("}")
	}

	b.BlankLine()
	b.Line("return json.Marshal(result)")
	b.Dedent()
	b.Line("}")

	// Generate UnmarshalJSON
	b.BlankLine()
	b.Line("func (s *%s) UnmarshalJSON(data []byte) error {", name)
	b.Indent()

	// Unmarshal into raw map for field extraction
	b.Line("var raw map[string]json.RawMessage")
	b.Line("if err := json.Unmarshal(data, &raw); err != nil {")
	b.Indent()
	b.Line("return err")
	b.Dedent()
	b.Line("}")
	b.BlankLine()

	// Unmarshal known fields
	for _, f := range fields {
		b.Line("if v, ok := raw[%q]; ok {", f.JSONName)
		b.Indent()
		if f.Pointer {
			baseType := strings.TrimPrefix(f.Type, "*")
			b.Line("var val %s", baseType)
			b.Line("if err := json.Unmarshal(v, &val); err != nil {")
			b.Indent()
			b.Line("return err")
			b.Dedent()
			b.Line("}")
			b.Line("s.%s = &val", f.Name)
		} else {
			b.Line("if err := json.Unmarshal(v, &s.%s); err != nil {", f.Name)
			b.Indent()
			b.Line("return err")
			b.Dedent()
			b.Line("}")
		}
		b.Dedent()
		b.Line("}")
	}

	// Unmarshal union fields from the full data
	for _, f := range unionFields {
		b.BlankLine()
		baseType := strings.TrimPrefix(f.Type, "*")
		b.Line("var %sVal %s", f.Name, baseType)
		b.Line("if err := json.Unmarshal(data, &%sVal); err != nil {", f.Name)
		b.Indent()
		b.Line("return err")
		b.Dedent()
		b.Line("}")
		b.Line("s.%s = &%sVal", f.Name, f.Name)
	}

	b.BlankLine()
	b.Line("return nil")
	b.Dedent()
	b.Line("}")

	return b.String()
}

// generateAnyOfType generates a union type for anyOf schemas.
func generateAnyOfType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	members := collectUnionMembers(gen, desc, desc.AnyOf, desc.Schema.AnyOf)
	if len(members) == 0 {
		return ""
	}

	// anyOf types only need encoding/json (not fmt like oneOf)
	gen.AddJSONImport()

	doc := extractDescription(desc.Schema)
	structCode := GenerateUnionType(desc.StableName, members, false, doc)

	code := structCode

	// If there's a short name different from stable name, generate an alias
	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}

	marshalCode := GenerateUnionMarshalAnyOf(desc.StableName, members)
	unmarshalCode := GenerateUnionUnmarshalAnyOf(desc.StableName, members)
	applyDefaultsCode := GenerateUnionApplyDefaults(desc.StableName, members)

	code += "\n" + marshalCode + "\n" + unmarshalCode + "\n" + applyDefaultsCode

	return code
}

// generateOneOfType generates a union type for oneOf schemas.
func generateOneOfType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	members := collectUnionMembers(gen, desc, desc.OneOf, desc.Schema.OneOf)
	if len(members) == 0 {
		return ""
	}

	// Union types need encoding/json and fmt for marshal/unmarshal
	gen.AddJSONImports()

	doc := extractDescription(desc.Schema)
	structCode := GenerateUnionType(desc.StableName, members, true, doc)

	code := structCode

	// If there's a short name different from stable name, generate an alias
	if desc.ShortName != desc.StableName {
		code += "\n" + GenerateTypeAlias(desc.ShortName, desc.StableName, "")
	}

	marshalCode := GenerateUnionMarshalOneOf(desc.StableName, members)
	unmarshalCode := GenerateUnionUnmarshalOneOf(desc.StableName, members)
	applyDefaultsCode := GenerateUnionApplyDefaults(desc.StableName, members)

	code += "\n" + marshalCode + "\n" + unmarshalCode + "\n" + applyDefaultsCode

	return code
}

// loadCustomType loads a custom type template and returns its code and imports.
func loadCustomType(templateName string) (string, map[string]string) {
	// Lookup the type definition
	typeName := strings.TrimSuffix(templateName, ".tmpl")

	// Map template name to type info from registry
	var typeDef templates.TypeTemplate
	var found bool

	for name, def := range templates.TypeTemplates {
		if def.Template == "types/"+templateName || strings.ToLower(name) == typeName {
			typeDef = def
			found = true
			break
		}
	}

	if !found {
		return "", nil
	}

	// Read the template file
	content, err := templates.TemplateFS.ReadFile("files/" + typeDef.Template)
	if err != nil {
		return "", nil
	}

	// Remove the template comment header
	code := string(content)
	if idx := strings.Index(code, "}}"); idx != -1 {
		code = strings.TrimLeft(code[idx+2:], "\n")
	}

	// Build imports map
	imports := make(map[string]string)
	for _, imp := range typeDef.Imports {
		imports[imp.Path] = imp.Alias
	}

	return code, imports
}

// schemaHasApplyDefaults returns true if the schema will have an ApplyDefaults method generated.
// This is true for:
// - Object types with properties
// - Union types (oneOf/anyOf)
// - AllOf types (merged structs)
// This is false for:
// - Primitive types (string, integer, boolean, number)
// - Enum types (without object properties)
// - Arrays
// - Maps (additionalProperties only)
func schemaHasApplyDefaults(schema *base.Schema) bool {
	if schema == nil {
		return false
	}

	// Has properties -> object type with ApplyDefaults
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		return true
	}

	// Has oneOf/anyOf -> union type with ApplyDefaults
	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		return true
	}

	// Has allOf -> merged struct with ApplyDefaults
	if len(schema.AllOf) > 0 {
		return true
	}

	return false
}

// collectUnionMembers gathers union member information for anyOf/oneOf.
func collectUnionMembers(gen *TypeGenerator, parentDesc *SchemaDescriptor, memberDescs []*SchemaDescriptor, memberProxies []*base.SchemaProxy) []UnionMember {
	var members []UnionMember

	// Build a map of schema paths to descriptors for lookup
	descByPath := make(map[string]*SchemaDescriptor)
	for _, desc := range memberDescs {
		if desc != nil {
			descByPath[desc.Path.String()] = desc
		}
	}

	for i, proxy := range memberProxies {
		var memberType string
		var fieldName string
		var hasApplyDefaults bool

		if proxy.IsReference() {
			ref := proxy.GetReference()
			if target, ok := gen.schemaIndex[ref]; ok {
				memberType = target.ShortName
				fieldName = target.ShortName
				hasApplyDefaults = schemaHasApplyDefaults(target.Schema)
			} else {
				continue
			}
		} else {
			// Check if this inline schema has a descriptor
			schema := proxy.Schema()
			if schema == nil {
				continue
			}

			// Determine the path for this member to look up its descriptor
			var memberPath SchemaPath
			if parentDesc != nil {
				// Try to find a descriptor by constructing the expected path
				memberPath = parentDesc.Path.Append("anyOf", fmt.Sprintf("%d", i))
				if _, ok := descByPath[memberPath.String()]; !ok {
					memberPath = parentDesc.Path.Append("oneOf", fmt.Sprintf("%d", i))
				}
			}

			if desc, ok := descByPath[memberPath.String()]; ok && desc.ShortName != "" {
				memberType = desc.ShortName
				fieldName = desc.ShortName
				hasApplyDefaults = schemaHasApplyDefaults(desc.Schema)
			} else {
				// This is a primitive type that doesn't have a named type
				goType := gen.goTypeForSchema(schema, nil)
				memberType = goType
				// Create a field name based on the type
				fieldName = gen.converter.ToTypeName(goType) + fmt.Sprintf("%d", i)
				hasApplyDefaults = false // Primitive types don't have ApplyDefaults
			}
		}

		members = append(members, UnionMember{
			FieldName:        fieldName,
			TypeName:         memberType,
			Index:            i,
			HasApplyDefaults: hasApplyDefaults,
		})
	}

	return members
}
