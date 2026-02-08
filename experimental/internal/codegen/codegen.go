// Package codegen generates Go code from parsed OpenAPI specs.
package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"

	"github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/templates"
)

// Generate produces Go code from the parsed OpenAPI document.
// specData is the raw spec bytes used to embed the spec in the generated code.
func Generate(doc libopenapi.Document, specData []byte, cfg Configuration) (string, error) {
	cfg.ApplyDefaults()

	// Create a single CodegenContext that all generators share.
	ctx := NewCodegenContext()

	// Create content type matcher for filtering request/response bodies
	contentTypeMatcher := NewContentTypeMatcher(cfg.ContentTypes)

	// Create content type short namer for friendly type names
	contentTypeNamer := NewContentTypeShortNamer(cfg.ContentTypeShortNames)

	// Pass 1: Gather all schemas that need types.
	// Operation filters (include/exclude tags, operation IDs) are applied during
	// gathering so that schemas from excluded operations are never collected.
	schemas, err := GatherSchemas(doc, contentTypeMatcher, cfg.OutputOptions)
	if err != nil {
		return "", fmt.Errorf("gathering schemas: %w", err)
	}

	// Filter explicitly excluded schemas
	schemas = FilterSchemasByName(schemas, cfg.OutputOptions.ExcludeSchemas)

	// Optionally prune component schemas that aren't referenced by any other schema
	if cfg.OutputOptions.PruneUnreferencedSchemas {
		schemas = PruneUnreferencedSchemas(schemas)
	}

	// Pass 2: Compute names for all schemas
	converter := NewNameConverter(cfg.NameMangling, cfg.NameSubstitutions)
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	// Build schema index for type resolution
	schemaIndex := make(map[string]*SchemaDescriptor)
	for _, s := range schemas {
		schemaIndex[s.Path.String()] = s
	}

	// Pass 3: Generate Go code
	importResolver := NewImportResolver(cfg.ImportMapping)
	tagGenerator := NewStructTagGenerator(cfg.StructTags)
	gen := NewTypeGenerator(cfg.TypeMapping, converter, importResolver, tagGenerator, ctx)
	gen.IndexSchemas(schemas)

	// Enum pre-pass: collect EnumInfo for all enum schemas, run collision detection
	gen.resolveEnumNames(schemas, cfg.OutputOptions.AlwaysPrefixEnumValues)

	output := NewOutput(cfg.PackageName)

	// ── Phase 1: Generate all code sections ──

	// Generate models (types for schemas) unless using external models package
	if cfg.Generation.ModelsPackage == nil {
		for _, desc := range schemas {
			code := generateType(gen, desc)
			if code != "" {
				output.AddType(code)
			}
		}

		// Embed the raw OpenAPI spec if specData was provided
		if len(specData) > 0 {
			embeddedCode, err := generateEmbeddedSpec(specData)
			if err != nil {
				return "", fmt.Errorf("generating embedded spec: %w", err)
			}
			output.AddType(embeddedCode)
			ctx.AddImport("bytes")
			ctx.AddImport("compress/gzip")
			ctx.AddImport("encoding/base64")
			ctx.AddImport("fmt")
			ctx.AddImport("strings")
			ctx.AddImport("sync")
		}
	}

	// Generate client code if requested
	if cfg.Generation.Client {
		ops, err := GatherOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering operations: %w", err)
		}

		ops = FilterOperations(ops, cfg.OutputOptions)

		clientGen, err := NewClientGenerator(schemaIndex, cfg.Generation.SimpleClient, cfg.Generation.ModelsPackage)
		if err != nil {
			return "", fmt.Errorf("creating client generator: %w", err)
		}

		clientCode, err := clientGen.GenerateClient(ops)
		if err != nil {
			return "", fmt.Errorf("generating client code: %w", err)
		}
		output.AddType(clientCode)

		// Add client template imports
		for _, ct := range templates.ClientTemplates {
			ctx.AddTemplateImports(ct.Imports)
		}

		// Add models package import if using external models
		if cfg.Generation.ModelsPackage != nil && cfg.Generation.ModelsPackage.Path != "" {
			ctx.AddImportAlias(cfg.Generation.ModelsPackage.Path, cfg.Generation.ModelsPackage.Alias)
		}

		// Register form helper if any operation has form-encoded bodies
		ctx.NeedFormHelper(ops)
	}

	// Track whether shared error types have been generated to avoid duplication.
	generatedErrors := false

	// Generate server code for path operations if a server framework is set.
	if cfg.Generation.Server != "" {
		ops, err := GatherOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering operations: %w", err)
		}

		ops = FilterOperations(ops, cfg.OutputOptions)

		if len(ops) > 0 {
			serverGen, err := NewServerGenerator(cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("creating server generator: %w", err)
			}

			serverCode, err := serverGen.GenerateServer(ops)
			if err != nil {
				return "", fmt.Errorf("generating server code: %w", err)
			}
			output.AddType(serverCode)
			generatedErrors = true

			// Add server template imports
			serverTemplates, err := getServerTemplates(cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("getting server templates: %w", err)
			}
			for _, st := range serverTemplates {
				ctx.AddTemplateImports(st.Imports)
			}
		}
	}

	// Generate webhook initiator code if requested
	if cfg.Generation.WebhookInitiator {
		webhookOps, err := GatherWebhookOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering webhook operations: %w", err)
		}

		if len(webhookOps) > 0 {
			initiatorGen, err := NewInitiatorGenerator("Webhook", schemaIndex, true, cfg.Generation.ModelsPackage)
			if err != nil {
				return "", fmt.Errorf("creating webhook initiator generator: %w", err)
			}

			initiatorCode, err := initiatorGen.GenerateInitiator(webhookOps)
			if err != nil {
				return "", fmt.Errorf("generating webhook initiator code: %w", err)
			}
			output.AddType(initiatorCode)

			for _, pt := range templates.InitiatorTemplates {
				ctx.AddTemplateImports(pt.Imports)
			}

			if cfg.Generation.ModelsPackage != nil && cfg.Generation.ModelsPackage.Path != "" {
				ctx.AddImportAlias(cfg.Generation.ModelsPackage.Path, cfg.Generation.ModelsPackage.Alias)
			}

			ctx.NeedFormHelper(webhookOps)
		}
	}

	// Generate callback initiator code if requested
	if cfg.Generation.CallbackInitiator {
		callbackOps, err := GatherCallbackOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering callback operations: %w", err)
		}

		if len(callbackOps) > 0 {
			initiatorGen, err := NewInitiatorGenerator("Callback", schemaIndex, true, cfg.Generation.ModelsPackage)
			if err != nil {
				return "", fmt.Errorf("creating callback initiator generator: %w", err)
			}

			initiatorCode, err := initiatorGen.GenerateInitiator(callbackOps)
			if err != nil {
				return "", fmt.Errorf("generating callback initiator code: %w", err)
			}
			output.AddType(initiatorCode)

			for _, pt := range templates.InitiatorTemplates {
				ctx.AddTemplateImports(pt.Imports)
			}

			if cfg.Generation.ModelsPackage != nil && cfg.Generation.ModelsPackage.Path != "" {
				ctx.AddImportAlias(cfg.Generation.ModelsPackage.Path, cfg.Generation.ModelsPackage.Alias)
			}

			ctx.NeedFormHelper(callbackOps)
		}
	}

	// Generate webhook receiver code if requested
	if cfg.Generation.WebhookReceiver {
		if cfg.Generation.Server == "" {
			return "", fmt.Errorf("webhook-receiver requires server to be set")
		}

		webhookOps, err := GatherWebhookOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering webhook operations: %w", err)
		}

		if len(webhookOps) > 0 {
			receiverGen, err := NewReceiverGenerator("Webhook", cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("creating webhook receiver generator: %w", err)
			}

			receiverCode, err := receiverGen.GenerateReceiver(webhookOps)
			if err != nil {
				return "", fmt.Errorf("generating webhook receiver code: %w", err)
			}
			output.AddType(receiverCode)

			paramTypes, err := receiverGen.GenerateParamTypes(webhookOps)
			if err != nil {
				return "", fmt.Errorf("generating webhook receiver param types: %w", err)
			}
			output.AddType(paramTypes)

			if !generatedErrors {
				errors, err := receiverGen.GenerateErrors()
				if err != nil {
					return "", fmt.Errorf("generating webhook receiver errors: %w", err)
				}
				output.AddType(errors)
				generatedErrors = true
			}

			receiverTemplates, err := getReceiverTemplates(cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("getting receiver templates: %w", err)
			}
			for _, ct := range receiverTemplates {
				ctx.AddTemplateImports(ct.Imports)
			}
			for _, st := range templates.SharedServerTemplates {
				ctx.AddTemplateImports(st.Imports)
			}
		}
	}

	// Generate callback receiver code if requested
	if cfg.Generation.CallbackReceiver {
		if cfg.Generation.Server == "" {
			return "", fmt.Errorf("callback-receiver requires server to be set")
		}

		callbackOps, err := GatherCallbackOperations(doc, ctx, contentTypeMatcher)
		if err != nil {
			return "", fmt.Errorf("gathering callback operations: %w", err)
		}

		if len(callbackOps) > 0 {
			receiverGen, err := NewReceiverGenerator("Callback", cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("creating callback receiver generator: %w", err)
			}

			receiverCode, err := receiverGen.GenerateReceiver(callbackOps)
			if err != nil {
				return "", fmt.Errorf("generating callback receiver code: %w", err)
			}
			output.AddType(receiverCode)

			paramTypes, err := receiverGen.GenerateParamTypes(callbackOps)
			if err != nil {
				return "", fmt.Errorf("generating callback receiver param types: %w", err)
			}
			output.AddType(paramTypes)

			if !generatedErrors {
				errors, err := receiverGen.GenerateErrors()
				if err != nil {
					return "", fmt.Errorf("generating callback receiver errors: %w", err)
				}
				output.AddType(errors)
				generatedErrors = true //nolint:ineffassign // kept for symmetry with webhook loop
			}

			receiverTemplates, err := getReceiverTemplates(cfg.Generation.Server)
			if err != nil {
				return "", fmt.Errorf("getting receiver templates: %w", err)
			}
			for _, ct := range receiverTemplates {
				ctx.AddTemplateImports(ct.Imports)
			}
			for _, st := range templates.SharedServerTemplates {
				ctx.AddTemplateImports(st.Imports)
			}
		}
	}

	// ── Phase 2: Render helpers ──

	// Emit custom type templates (Date, Email, UUID, File, Nullable, etc.)
	for _, templateName := range ctx.RequiredCustomTypes() {
		typeCode := ctx.loadAndRegisterCustomType(templateName)
		if typeCode != "" {
			output.AddType(typeCode)
		}
	}

	// Emit param functions
	paramFuncs, err := generateParamFunctionsFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("generating param functions: %w", err)
	}
	if paramFuncs != "" {
		output.AddType(paramFuncs)
	}

	// Emit helper templates (e.g., marshal_form)
	for _, helperName := range ctx.RequiredHelpers() {
		helperCode, err := generateHelper(helperName, ctx)
		if err != nil {
			return "", fmt.Errorf("generating helper %s: %w", helperName, err)
		}
		if helperCode != "" {
			output.AddType(helperCode)
		}
	}

	// ── Phase 3: Assemble output ──
	// Transfer all imports from ctx to output
	output.AddImports(ctx.Imports())

	return output.Format()
}

// generateHelper generates a helper template by name and registers its imports on the context.
func generateHelper(name string, ctx *CodegenContext) (string, error) {
	switch name {
	case "marshal_form":
		ctx.AddTemplateImports(templates.MarshalFormHelperTemplate.Imports)
		return generateMarshalFormHelper()
	default:
		return "", fmt.Errorf("unknown helper: %s", name)
	}
}

// generateMarshalFormHelper generates the marshalForm helper function.
func generateMarshalFormHelper() (string, error) {
	tmplInfo := templates.MarshalFormHelperTemplate
	content, err := templates.TemplateFS.ReadFile("files/" + tmplInfo.Template)
	if err != nil {
		return "", fmt.Errorf("reading form helper template: %w", err)
	}

	tmpl, err := template.New(tmplInfo.Name).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing form helper template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, nil); err != nil {
		return "", fmt.Errorf("executing form helper template: %w", err)
	}

	return result.String(), nil
}

// generateParamFunctionsFromContext generates the parameter styling/binding functions based on CodegenContext usage.
func generateParamFunctionsFromContext(ctx *CodegenContext) (string, error) {
	if !ctx.HasAnyParams() {
		return "", nil
	}

	var result strings.Builder

	requiredTemplates := ctx.GetRequiredParamTemplates()

	for _, tmplInfo := range requiredTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + tmplInfo.Template)
		if err != nil {
			return "", fmt.Errorf("reading param template %s: %w", tmplInfo.Template, err)
		}

		tmpl, err := template.New(tmplInfo.Name).Parse(string(content))
		if err != nil {
			return "", fmt.Errorf("parsing param template %s: %w", tmplInfo.Template, err)
		}

		if err := tmpl.Execute(&result, nil); err != nil {
			return "", fmt.Errorf("executing param template %s: %w", tmplInfo.Template, err)
		}
		result.WriteString("\n")
	}

	// Register param imports on the context
	for _, imp := range ctx.GetRequiredParamImports() {
		ctx.AddImportAlias(imp.Path, imp.Alias)
	}

	return result.String(), nil
}

// generateType generates Go code for a single schema descriptor.
func generateType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	kind := GetSchemaKind(desc)

	// If schema has TypeOverride extension, generate a type alias to the external type
	// instead of generating the full type definition
	if desc.Extensions != nil && desc.Extensions.TypeOverride != nil {
		return generateTypeOverrideAlias(gen, desc)
	}

	var code string
	switch kind {
	case KindReference:
		// Internal references don't generate new types; they use the referenced type's name.
		// However, component schemas that are $ref to external types need a type alias
		// so that local types can reference them by their local name.
		if desc.IsExternalReference() && desc.IsTopLevelComponentSchema() && desc.ShortName != "" {
			code = generateExternalRefAlias(gen, desc)
			break
		}
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
		// Mixed properties need encoding/json for marshal/unmarshal (but not fmt)
		gen.AddJSONImport()

		addPropsType := gen.AdditionalPropertiesType(desc)
		structCode := GenerateStructWithAdditionalProps(desc.ShortName, fields, addPropsType, doc, gen.TagGenerator())

		// Generate marshal/unmarshal methods
		marshalCode := GenerateMixedPropertiesMarshal(desc.ShortName, fields)
		unmarshalCode := GenerateMixedPropertiesUnmarshal(desc.ShortName, fields, addPropsType)

		code := structCode + "\n" + marshalCode + "\n" + unmarshalCode

		// Generate ApplyDefaults method if needed
		if applyDefaults, needsReflect := GenerateApplyDefaults(desc.ShortName, fields); applyDefaults != "" {
			code += "\n" + applyDefaults
			if needsReflect {
				gen.AddImport("reflect")
			}
		}

		return code
	}

	code := GenerateStruct(desc.ShortName, fields, doc, gen.TagGenerator())

	// Generate ApplyDefaults method if needed
	if applyDefaults, needsReflect := GenerateApplyDefaults(desc.ShortName, fields); applyDefaults != "" {
		code += "\n" + applyDefaults
		if needsReflect {
			gen.AddImport("reflect")
		}
	}

	return code
}

// generateMapAlias generates a type alias for a pure map schema.
func generateMapAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	mapType := gen.GoTypeExpr(desc)
	doc := extractDescription(desc.Schema)
	return GenerateTypeAlias(desc.ShortName, mapType, doc)
}

// generateEnumType generates an enum type with const values.
// It uses pre-computed EnumInfo from the enum pre-pass for collision-aware naming.
func generateEnumType(gen *TypeGenerator, desc *SchemaDescriptor) string {
	info, ok := gen.enumInfoMap[desc.Path.String()]
	if !ok {
		// Fallback: shouldn't happen, but build info on the fly
		info = gen.buildEnumInfo(desc)
		if info == nil {
			return ""
		}
		computeEnumConstantNames([]*EnumInfo{info}, gen.converter)
	}

	return GenerateEnumFromInfo(info)
}

// generateExternalRefAlias generates a type alias for a component schema that is
// a $ref to an external type. This ensures the local name exists so other local
// types can reference it.
func generateExternalRefAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	goType := gen.externalRefType(desc)
	if goType == "any" {
		return ""
	}
	return GenerateTypeAlias(desc.ShortName, goType, "")
}

// generateTypeAlias generates a simple type alias.
func generateTypeAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	goType := gen.GoTypeExpr(desc)
	doc := extractDescription(desc.Schema)
	return GenerateTypeAlias(desc.ShortName, goType, doc)
}

// generateTypeOverrideAlias generates a type alias to an external type specified via x-oapi-codegen-type-override.
func generateTypeOverrideAlias(gen *TypeGenerator, desc *SchemaDescriptor) string {
	override := desc.Extensions.TypeOverride

	// Register the import
	if override.ImportPath != "" {
		if override.ImportAlias != "" {
			gen.AddImportAlias(override.ImportPath, override.ImportAlias)
		} else {
			gen.AddImport(override.ImportPath)
		}
	}

	doc := extractDescription(desc.Schema)
	return GenerateTypeAlias(desc.ShortName, override.TypeName, doc)
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
				info.fields = gen.collectFieldsRecursive(target)
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
		code = generateAllOfStructWithUnions(desc.ShortName, finalFields, unionFields, doc, gen.TagGenerator())
	} else {
		// Simple case - just flattened fields
		code = GenerateStruct(desc.ShortName, finalFields, doc, gen.TagGenerator())
	}

	// Generate ApplyDefaults method if needed
	if applyDefaults, needsReflect := GenerateApplyDefaults(desc.ShortName, finalFields); applyDefaults != "" {
		code += "\n" + applyDefaults
		if needsReflect {
			gen.AddImport("reflect")
		}
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
	code := GenerateUnionType(desc.ShortName, members, false, doc)

	marshalCode := GenerateUnionMarshalAnyOf(desc.ShortName, members)
	unmarshalCode := GenerateUnionUnmarshalAnyOf(desc.ShortName, members)
	applyDefaultsCode := GenerateUnionApplyDefaults(desc.ShortName, members)

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
	code := GenerateUnionType(desc.ShortName, members, true, doc)

	marshalCode := GenerateUnionMarshalOneOf(desc.ShortName, members)
	unmarshalCode := GenerateUnionUnmarshalOneOf(desc.ShortName, members)
	applyDefaultsCode := GenerateUnionApplyDefaults(desc.ShortName, members)

	code += "\n" + marshalCode + "\n" + unmarshalCode + "\n" + applyDefaultsCode

	return code
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
