package codegen

import (
	"fmt"
	"log/slog"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// GatherResult contains the results of gathering from an OpenAPI document.
type GatherResult struct {
	Schemas []*SchemaDescriptor
	Ctx     *CodegenContext
}

// GatherSchemas traverses an OpenAPI document and collects all schemas into a list.
// When outputOpts contains operation filters (include/exclude tags or operation IDs),
// schemas from excluded operations are not gathered.
func GatherSchemas(doc libopenapi.Document, contentTypeMatcher *ContentTypeMatcher, outputOpts OutputOptions) ([]*SchemaDescriptor, error) {
	result, err := GatherAll(doc, contentTypeMatcher, outputOpts)
	if err != nil {
		return nil, err
	}
	return result.Schemas, nil
}

// GatherAll traverses an OpenAPI document and collects all schemas and parameter usage.
// When outputOpts contains operation filters, schemas from excluded operations are skipped.
func GatherAll(doc libopenapi.Document, contentTypeMatcher *ContentTypeMatcher, outputOpts OutputOptions) (*GatherResult, error) {
	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("building v3 model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("failed to build v3 model")
	}

	g := &gatherer{
		schemas:            make([]*SchemaDescriptor, 0),
		contentTypeMatcher: contentTypeMatcher,
		ctx:                NewCodegenContext(),
		outputOpts:         outputOpts,
	}

	g.gatherFromDocument(&model.Model)
	return &GatherResult{
		Schemas: g.schemas,
		Ctx:     g.ctx,
	}, nil
}

type gatherer struct {
	schemas            []*SchemaDescriptor
	contentTypeMatcher *ContentTypeMatcher
	ctx                *CodegenContext
	outputOpts         OutputOptions
	// Context for the current operation being gathered (for nicer naming)
	currentOperationID string
	currentContentType string
}

func (g *gatherer) gatherFromDocument(doc *v3.Document) {
	// Gather from components/schemas
	if doc.Components != nil && doc.Components.Schemas != nil {
		for pair := doc.Components.Schemas.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			schemaProxy := pair.Value()
			path := SchemaPath{"components", "schemas", name}
			g.gatherFromSchemaProxy(schemaProxy, path, nil)
		}
	}

	// Gather from paths
	if doc.Paths != nil && doc.Paths.PathItems != nil {
		for pair := doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			pathStr := pair.Key()
			pathItem := pair.Value()
			g.gatherFromPathItem(pathItem, SchemaPath{"paths", pathStr})
		}
	}

	// Gather from webhooks (3.1+)
	if doc.Webhooks != nil {
		for pair := doc.Webhooks.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			pathItem := pair.Value()
			g.gatherFromPathItem(pathItem, SchemaPath{"webhooks", name})
		}
	}
}

func (g *gatherer) gatherFromPathItem(pathItem *v3.PathItem, basePath SchemaPath) {
	if pathItem == nil {
		return
	}

	// Path-level parameters are gathered if any operation on this path is included.
	// We defer gathering them until we know at least one operation passes filters.
	pathParamsGathered := false

	// Operations
	ops := pathItem.GetOperations()
	if ops != nil {
		for pair := ops.First(); pair != nil; pair = pair.Next() {
			method := pair.Key()
			op := pair.Value()

			// Apply operation filters: skip operations excluded by tags or operation IDs
			if g.shouldSkipOperation(op) {
				continue
			}

			// Gather path-level parameters once (shared across all operations on this path)
			if !pathParamsGathered {
				for i, param := range pathItem.Parameters {
					g.gatherFromParameter(param, basePath.Append("parameters", fmt.Sprintf("%d", i)))
				}
				pathParamsGathered = true
			}

			g.gatherFromOperation(op, basePath.Append(method))
		}
	}
}

// shouldSkipOperation returns true if the operation should be excluded based on
// the configured tag and operation ID filters.
func (g *gatherer) shouldSkipOperation(op *v3.Operation) bool {
	if op == nil {
		return true
	}

	// Apply exclude tags first
	if len(g.outputOpts.ExcludeTags) > 0 {
		for _, tag := range op.Tags {
			for _, excluded := range g.outputOpts.ExcludeTags {
				if tag == excluded {
					return true
				}
			}
		}
	}

	// Apply include tags (operation must have at least one included tag)
	if len(g.outputOpts.IncludeTags) > 0 {
		hasIncludedTag := false
		for _, tag := range op.Tags {
			for _, included := range g.outputOpts.IncludeTags {
				if tag == included {
					hasIncludedTag = true
					break
				}
			}
			if hasIncludedTag {
				break
			}
		}
		if !hasIncludedTag {
			return true
		}
	}

	// Apply exclude operation IDs
	if len(g.outputOpts.ExcludeOperationIDs) > 0 && op.OperationId != "" {
		for _, excluded := range g.outputOpts.ExcludeOperationIDs {
			if op.OperationId == excluded {
				return true
			}
		}
	}

	// Apply include operation IDs (operation must match one)
	if len(g.outputOpts.IncludeOperationIDs) > 0 {
		if op.OperationId == "" {
			return true
		}
		found := false
		for _, included := range g.outputOpts.IncludeOperationIDs {
			if op.OperationId == included {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	return false
}

func (g *gatherer) gatherFromOperation(op *v3.Operation, basePath SchemaPath) {
	if op == nil {
		return
	}

	// Set operation context for nicer naming
	prevOperationID := g.currentOperationID
	if op.OperationId != "" {
		g.currentOperationID = op.OperationId
	}

	// Parameters
	for i, param := range op.Parameters {
		g.gatherFromParameter(param, basePath.Append("parameters", fmt.Sprintf("%d", i)))
	}

	// Request body
	if op.RequestBody != nil {
		g.gatherFromRequestBody(op.RequestBody, basePath.Append("requestBody"))
	}

	// Responses
	if op.Responses != nil && op.Responses.Codes != nil {
		for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
			code := pair.Key()
			response := pair.Value()
			g.gatherFromResponse(response, basePath.Append("responses", code))
		}
	}

	// Callbacks
	if op.Callbacks != nil {
		for pair := op.Callbacks.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			callback := pair.Value()
			g.gatherFromCallback(callback, basePath.Append("callbacks", name))
		}
	}

	// Restore previous operation context
	g.currentOperationID = prevOperationID
}

func (g *gatherer) gatherFromParameter(param *v3.Parameter, basePath SchemaPath) {
	if param == nil {
		return
	}

	// Track parameter styling usage for code generation
	if g.ctx != nil && param.Schema != nil {
		// Determine style (with defaults based on location)
		style := param.Style
		if style == "" {
			style = DefaultParamStyle(param.In)
		}

		// Determine explode (with defaults based on location)
		explode := DefaultParamExplode(param.In)
		if param.Explode != nil {
			explode = *param.Explode
		}

		// Record both style (client) and bind (server) usage
		g.ctx.NeedParam(style, explode)
	}

	if param.Schema != nil {
		g.gatherFromSchemaProxy(param.Schema, basePath.Append("schema"), nil)
	}

	// Parameter can also have content with schemas
	if param.Content != nil {
		for pair := param.Content.First(); pair != nil; pair = pair.Next() {
			contentType := pair.Key()
			mediaType := pair.Value()
			g.gatherFromMediaType(mediaType, basePath.Append("content", contentType))
		}
	}
}

func (g *gatherer) gatherFromRequestBody(rb *v3.RequestBody, basePath SchemaPath) {
	if rb == nil || rb.Content == nil {
		return
	}

	for pair := rb.Content.First(); pair != nil; pair = pair.Next() {
		contentType := pair.Key()
		// Skip content types that don't match the configured patterns
		if g.contentTypeMatcher != nil && !g.contentTypeMatcher.Matches(contentType) {
			continue
		}
		// Set content type context
		prevContentType := g.currentContentType
		g.currentContentType = contentType

		mediaType := pair.Value()
		g.gatherFromMediaType(mediaType, basePath.Append("content", contentType))

		g.currentContentType = prevContentType
	}
}

func (g *gatherer) gatherFromResponse(response *v3.Response, basePath SchemaPath) {
	if response == nil {
		return
	}

	if response.Content != nil {
		for pair := response.Content.First(); pair != nil; pair = pair.Next() {
			contentType := pair.Key()
			// Skip content types that don't match the configured patterns
			if g.contentTypeMatcher != nil && !g.contentTypeMatcher.Matches(contentType) {
				continue
			}
			// Set content type context
			prevContentType := g.currentContentType
			g.currentContentType = contentType

			mediaType := pair.Value()
			g.gatherFromMediaType(mediaType, basePath.Append("content", contentType))

			g.currentContentType = prevContentType
		}
	}

	// Response headers can have schemas
	if response.Headers != nil {
		for pair := response.Headers.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			header := pair.Value()
			if header != nil && header.Schema != nil {
				g.gatherFromSchemaProxy(header.Schema, basePath.Append("headers", name, "schema"), nil)
			}
		}
	}
}

func (g *gatherer) gatherFromMediaType(mt *v3.MediaType, basePath SchemaPath) {
	if mt == nil || mt.Schema == nil {
		return
	}
	g.gatherFromSchemaProxy(mt.Schema, basePath.Append("schema"), nil)
}

func (g *gatherer) gatherFromCallback(callback *v3.Callback, basePath SchemaPath) {
	if callback == nil || callback.Expression == nil {
		return
	}

	for pair := callback.Expression.First(); pair != nil; pair = pair.Next() {
		expr := pair.Key()
		pathItem := pair.Value()
		g.gatherFromPathItem(pathItem, basePath.Append(expr))
	}
}

func (g *gatherer) gatherFromSchemaProxy(proxy *base.SchemaProxy, path SchemaPath, parent *SchemaDescriptor) *SchemaDescriptor {
	if proxy == nil {
		return nil
	}

	// Check if this is a reference
	isRef := proxy.IsReference()
	ref := ""
	if isRef {
		ref = proxy.GetReference()
	}

	// Get the resolved schema
	schema := proxy.Schema()

	// Check if schema has extensions that require type generation.
	// Skip extension checks for references — libopenapi copies extensions from the
	// resolved target, but those extensions belong to the component schema, not to
	// each reference site. Without this guard, a property like
	//   pagination: { $ref: '#/components/schemas/Pagination' }
	// would inherit Pagination's x-go-type and be gathered as a separate type,
	// producing duplicate declarations. (See oapi-codegen-exp#14.)
	hasTypeOverride := !isRef && schema != nil && schema.Extensions != nil && hasExtension(schema.Extensions, ExtTypeOverride, legacyExtGoType)
	hasTypeNameOverride := !isRef && schema != nil && schema.Extensions != nil && hasExtension(schema.Extensions, ExtTypeNameOverride, legacyExtGoTypeName)

	// Only gather schemas that need a generated type
	// References are always gathered (they point to real schemas)
	// Simple types (primitives without enum) are skipped for inline schemas
	// Inline nullable primitives (under properties/) don't need types - they use Nullable[T] directly
	// Schemas with type-override or type-name-override extensions always need types
	// Component schemas (components/schemas/*) always get a type alias, even for primitives,
	// so that external packages can reference them by name.
	isComponentSchema := len(path) == 3 && path[0] == "components" && path[1] == "schemas"
	isInlineProperty := path.ContainsProperties()
	skipInlineNullablePrimitive := isInlineProperty && isNullablePrimitive(schema)
	needsType := isRef || needsGeneratedType(schema) || hasTypeOverride || hasTypeNameOverride || isComponentSchema
	if needsType && !skipInlineNullablePrimitive {
		desc := &SchemaDescriptor{
			Path:        path,
			Parent:      parent,
			Ref:         ref,
			Schema:      schema,
			OperationID: g.currentOperationID,
			ContentType: g.currentContentType,
		}

		// Parse extensions from the schema — but not for references.
		// When libopenapi resolves a $ref, the resolved schema carries extensions
		// from the target (e.g., x-go-type on a component schema). Those extensions
		// belong to the component schema descriptor, not to every reference site.
		if !isRef && schema != nil && schema.Extensions != nil {
			ext, err := ParseExtensions(schema.Extensions, path.String())
			if err != nil {
				slog.Warn("failed to parse extensions",
					"path", path.String(),
					"error", err)
			} else {
				desc.Extensions = ext
			}
		}

		g.schemas = append(g.schemas, desc)

		// Don't recurse into references - they point to schemas we'll gather elsewhere
		if isRef {
			return desc
		}

		// Recurse into schema structure
		if schema != nil {
			g.gatherFromSchema(schema, path, desc)
		}
		return desc
	} else if schema != nil {
		// Even if we don't gather this schema, we still need to recurse
		// to find any nested complex schemas (e.g., array items that are objects)
		g.gatherFromSchema(schema, path, nil)
	}
	return nil
}

// gatherSchemaDescriptorOnly creates a descriptor for field extraction without adding it
// to the schemas list (i.e., no type will be generated for it).
// This is used for inline allOf members whose fields are flattened into the parent.
func (g *gatherer) gatherSchemaDescriptorOnly(proxy *base.SchemaProxy, path SchemaPath, parent *SchemaDescriptor) *SchemaDescriptor {
	if proxy == nil {
		return nil
	}

	schema := proxy.Schema()
	if schema == nil {
		return nil
	}

	desc := &SchemaDescriptor{
		Path:   path,
		Parent: parent,
		Schema: schema,
	}

	// Parse extensions from the schema
	if schema.Extensions != nil {
		ext, err := ParseExtensions(schema.Extensions, path.String())
		if err != nil {
			slog.Warn("failed to parse extensions",
				"path", path.String(),
				"error", err)
		} else {
			desc.Extensions = ext
		}
	}

	// Still recurse to gather any nested complex schemas that DO need types
	// (e.g., nested objects within properties)
	g.gatherFromSchema(schema, path, desc)

	return desc
}

// needsGeneratedType returns true if a schema requires a generated Go type.
// Simple primitive types (string, integer, number, boolean) without enums
// don't need generated types - they map directly to Go builtins.
// However, nullable primitives DO need generated types (Nullable[T]).
func needsGeneratedType(schema *base.Schema) bool {
	if schema == nil {
		return false
	}

	// Nullable primitives need a generated type (Nullable[T])
	if isNullablePrimitive(schema) {
		return true
	}

	// Enums always need a generated type
	if len(schema.Enum) > 0 {
		return true
	}

	// Objects need a generated type
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		return true
	}

	// Check explicit type
	types := schema.Type
	for _, t := range types {
		if t == "object" {
			return true
		}
	}

	// Composition types need generated types
	if len(schema.AllOf) > 0 || len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return true
	}

	// Arrays with complex items need generated types for the array type itself
	// But we handle items separately in gatherFromSchema
	if schema.Items != nil && schema.Items.A != nil {
		itemSchema := schema.Items.A.Schema()
		if needsGeneratedType(itemSchema) {
			return true
		}
	}

	// AdditionalProperties with complex schema needs a type
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.A != nil {
		addSchema := schema.AdditionalProperties.A.Schema()
		if needsGeneratedType(addSchema) {
			return true
		}
	}

	// Simple primitives (string, integer, number, boolean) without enum
	// don't need generated types
	return false
}

// isNullablePrimitive returns true if the schema is a nullable primitive type.
// Nullable primitives need Nullable[T] wrapper types.
func isNullablePrimitive(schema *base.Schema) bool {
	if schema == nil {
		return false
	}

	// Check for nullable
	isNullable := false
	// OpenAPI 3.1 style: type array includes "null"
	for _, t := range schema.Type {
		if t == "null" {
			isNullable = true
			break
		}
	}
	// OpenAPI 3.0 style: nullable: true
	if schema.Nullable != nil && *schema.Nullable {
		isNullable = true
	}

	if !isNullable {
		return false
	}

	// Check if it's a primitive type (not object, array, or composition)
	if schema.Properties != nil && schema.Properties.Len() > 0 {
		return false // object with properties
	}
	if len(schema.AllOf) > 0 || len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 {
		return false // composition type
	}
	if schema.Items != nil {
		return false // array
	}

	// Get the primary type
	for _, t := range schema.Type {
		switch t {
		case "string", "integer", "number", "boolean":
			return true
		case "object":
			return false
		case "array":
			return false
		}
	}

	return false
}

func (g *gatherer) gatherFromSchema(schema *base.Schema, basePath SchemaPath, parent *SchemaDescriptor) {
	if schema == nil {
		return
	}

	// Properties
	if schema.Properties != nil {
		if parent != nil {
			parent.Properties = make(map[string]*SchemaDescriptor)
		}
		for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
			propName := pair.Key()
			propProxy := pair.Value()
			propPath := basePath.Append("properties", propName)
			propDesc := g.gatherFromSchemaProxy(propProxy, propPath, parent)
			if parent != nil && propDesc != nil {
				parent.Properties[propName] = propDesc
			}
		}
	}

	// Items (array element schema)
	if schema.Items != nil && schema.Items.A != nil {
		itemsPath := basePath.Append("items")
		itemsDesc := g.gatherFromSchemaProxy(schema.Items.A, itemsPath, parent)
		if parent != nil && itemsDesc != nil {
			parent.Items = itemsDesc
		}
	}

	// AllOf - inline object members don't need separate types since fields are flattened into parent
	// However, inline oneOf/anyOf members DO need union types generated
	for i, proxy := range schema.AllOf {
		allOfPath := basePath.Append("allOf", fmt.Sprintf("%d", i))
		var allOfDesc *SchemaDescriptor
		if proxy.IsReference() {
			// References still need to be gathered normally
			allOfDesc = g.gatherFromSchemaProxy(proxy, allOfPath, parent)
		} else {
			memberSchema := proxy.Schema()
			// If the allOf member is itself a oneOf/anyOf, we need to generate a union type
			if memberSchema != nil && (len(memberSchema.OneOf) > 0 || len(memberSchema.AnyOf) > 0) {
				allOfDesc = g.gatherFromSchemaProxy(proxy, allOfPath, parent)
			} else {
				// Simple inline objects: create descriptor for field extraction but don't generate a type
				allOfDesc = g.gatherSchemaDescriptorOnly(proxy, allOfPath, parent)
			}
		}
		if parent != nil && allOfDesc != nil {
			parent.AllOf = append(parent.AllOf, allOfDesc)
		}
	}

	// AnyOf
	for i, proxy := range schema.AnyOf {
		anyOfPath := basePath.Append("anyOf", fmt.Sprintf("%d", i))
		anyOfDesc := g.gatherFromSchemaProxy(proxy, anyOfPath, parent)
		if parent != nil && anyOfDesc != nil {
			parent.AnyOf = append(parent.AnyOf, anyOfDesc)
		}
	}

	// OneOf
	for i, proxy := range schema.OneOf {
		oneOfPath := basePath.Append("oneOf", fmt.Sprintf("%d", i))
		oneOfDesc := g.gatherFromSchemaProxy(proxy, oneOfPath, parent)
		if parent != nil && oneOfDesc != nil {
			parent.OneOf = append(parent.OneOf, oneOfDesc)
		}
	}

	// AdditionalProperties (if it's a schema, not a boolean)
	if schema.AdditionalProperties != nil && schema.AdditionalProperties.A != nil {
		addPropsPath := basePath.Append("additionalProperties")
		addPropsDesc := g.gatherFromSchemaProxy(schema.AdditionalProperties.A, addPropsPath, parent)
		if parent != nil && addPropsDesc != nil {
			parent.AdditionalProps = addPropsDesc
		}
	}

	// Not
	if schema.Not != nil {
		g.gatherFromSchemaProxy(schema.Not, basePath.Append("not"), parent)
	}

	// PrefixItems (3.1 tuple validation)
	for i, proxy := range schema.PrefixItems {
		g.gatherFromSchemaProxy(proxy, basePath.Append("prefixItems", fmt.Sprintf("%d", i)), parent)
	}

	// Contains (3.1)
	if schema.Contains != nil {
		g.gatherFromSchemaProxy(schema.Contains, basePath.Append("contains"), parent)
	}

	// If/Then/Else (3.1)
	if schema.If != nil {
		g.gatherFromSchemaProxy(schema.If, basePath.Append("if"), parent)
	}
	if schema.Then != nil {
		g.gatherFromSchemaProxy(schema.Then, basePath.Append("then"), parent)
	}
	if schema.Else != nil {
		g.gatherFromSchemaProxy(schema.Else, basePath.Append("else"), parent)
	}

	// DependentSchemas (3.1)
	if schema.DependentSchemas != nil {
		for pair := schema.DependentSchemas.First(); pair != nil; pair = pair.Next() {
			name := pair.Key()
			proxy := pair.Value()
			g.gatherFromSchemaProxy(proxy, basePath.Append("dependentSchemas", name), parent)
		}
	}

	// PatternProperties (3.1)
	if schema.PatternProperties != nil {
		for pair := schema.PatternProperties.First(); pair != nil; pair = pair.Next() {
			pattern := pair.Key()
			proxy := pair.Value()
			g.gatherFromSchemaProxy(proxy, basePath.Append("patternProperties", pattern), parent)
		}
	}

	// PropertyNames (3.1)
	if schema.PropertyNames != nil {
		g.gatherFromSchemaProxy(schema.PropertyNames, basePath.Append("propertyNames"), parent)
	}

	// UnevaluatedItems (3.1)
	if schema.UnevaluatedItems != nil {
		g.gatherFromSchemaProxy(schema.UnevaluatedItems, basePath.Append("unevaluatedItems"), parent)
	}

	// UnevaluatedProperties (3.1) - can be schema or bool
	if schema.UnevaluatedProperties != nil && schema.UnevaluatedProperties.A != nil {
		g.gatherFromSchemaProxy(schema.UnevaluatedProperties.A, basePath.Append("unevaluatedProperties"), parent)
	}
}
