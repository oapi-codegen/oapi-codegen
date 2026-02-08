package codegen

import (
	"strings"
)

// SchemaContext identifies what kind of schema this is based on its location.
type SchemaContext int

const (
	ContextUnknown SchemaContext = iota
	ContextComponentSchema
	ContextParameter
	ContextRequestBody
	ContextResponse
	ContextHeader
	ContextCallback
	ContextWebhook
	ContextProperty
	ContextItems
	ContextAllOf
	ContextAnyOf
	ContextOneOf
	ContextAdditionalProperties
)

// ComputeSchemaNames assigns StableName and ShortName to each schema descriptor.
// StableName is deterministic from the path; ShortName is a friendly alias.
// If a schema has a TypeNameOverride extension, that takes precedence over computed names.
func ComputeSchemaNames(schemas []*SchemaDescriptor, converter *NameConverter, contentTypeNamer *ContentTypeShortNamer) {
	// First: compute stable names from full paths
	for _, s := range schemas {
		// Check for TypeNameOverride extension
		if s.Extensions != nil && s.Extensions.TypeNameOverride != "" {
			s.StableName = s.Extensions.TypeNameOverride
		} else {
			s.StableName = computeStableName(s.Path, converter)
		}
	}

	// Second: generate candidate short names
	candidates := make(map[*SchemaDescriptor]string)
	for _, s := range schemas {
		// TypeNameOverride also applies to short names
		if s.Extensions != nil && s.Extensions.TypeNameOverride != "" {
			candidates[s] = s.Extensions.TypeNameOverride
		} else {
			candidates[s] = generateCandidateName(s, converter, contentTypeNamer)
		}
	}

	// Third: detect collisions and resolve them for short names
	resolveCollisions(schemas, candidates, converter)

	// Assign final short names
	for _, s := range schemas {
		s.ShortName = candidates[s]
	}
}

// computeStableName generates a deterministic type name from the full path.
// The format is: {meaningful_names}{reversed_context_suffix}
// Example: #/components/schemas/Cat -> CatSchemaComponent
func computeStableName(path SchemaPath, converter *NameConverter) string {
	if len(path) == 0 {
		return "Schema"
	}

	// Separate path into name parts and context parts
	var nameParts []string
	var contextParts []string

	for i := 0; i < len(path); i++ {
		part := path[i]
		// Strip leading slash from API paths (e.g., "/pets" -> "pets")
		part = strings.TrimPrefix(part, "/")
		if part == "" {
			continue
		}
		if isContextKeyword(part) {
			contextParts = append(contextParts, part)
		} else {
			nameParts = append(nameParts, part)
		}
	}

	// Build the name: names first, then reversed context as suffix
	var result strings.Builder

	// Add name parts
	// First part uses ToTypeName (adds numeric prefix if needed)
	// Subsequent parts use ToTypeNamePart (no numeric prefix since they're not at the start)
	for i, name := range nameParts {
		if i == 0 {
			result.WriteString(converter.ToTypeName(name))
		} else {
			result.WriteString(converter.ToTypeNamePart(name))
		}
	}

	// Add reversed context as suffix (singularized)
	for i := len(contextParts) - 1; i >= 0; i-- {
		suffix := contextToSuffix(contextParts[i])
		result.WriteString(suffix)
	}

	name := result.String()
	if name == "" {
		return "Schema"
	}
	return name
}

// isContextKeyword returns true if the path segment is a structural keyword
// rather than a user-defined name.
func isContextKeyword(s string) bool {
	switch s {
	case "components", "schemas", "parameters", "responses", "requestBodies",
		"headers", "callbacks", "paths", "webhooks",
		"properties", "items", "additionalProperties",
		"allOf", "anyOf", "oneOf", "not",
		"prefixItems", "contains", "if", "then", "else",
		"dependentSchemas", "patternProperties", "propertyNames",
		"unevaluatedItems", "unevaluatedProperties",
		"content", "schema", "requestBody":
		return true
	default:
		return false
	}
}

// contextToSuffix converts a context keyword to its singular suffix form.
func contextToSuffix(context string) string {
	switch context {
	case "components":
		return "Component"
	case "schemas":
		return "Schema"
	case "parameters":
		return "Parameter"
	case "responses":
		return "Response"
	case "requestBodies", "requestBody":
		return "Request"
	case "headers":
		return "Header"
	case "callbacks":
		return "Callback"
	case "paths":
		return "Path"
	case "webhooks":
		return "Webhook"
	case "properties":
		return "Property"
	case "items":
		return "Item"
	case "additionalProperties":
		return "Value"
	case "allOf":
		return "AllOf"
	case "anyOf":
		return "AnyOf"
	case "oneOf":
		return "OneOf"
	case "not":
		return "Not"
	case "prefixItems":
		return "PrefixItem"
	case "contains":
		return "Contains"
	case "if":
		return "If"
	case "then":
		return "Then"
	case "else":
		return "Else"
	case "content":
		return "Content"
	case "schema":
		return "" // Skip redundant "Schema" suffix from content/schema
	default:
		return ""
	}
}

// generateCandidateName creates a candidate short name based on the schema's path.
func generateCandidateName(s *SchemaDescriptor, converter *NameConverter, contentTypeNamer *ContentTypeShortNamer) string {
	path := s.Path
	if len(path) == 0 {
		return "Schema"
	}

	ctx, parts := parsePathContext(path)

	switch ctx {
	case ContextComponentSchema:
		// #/components/schemas/Cat -> "Cat"
		// #/components/schemas/Cat/properties/name -> "CatName"
		return buildComponentSchemaName(parts, converter)

	case ContextParameter:
		// Always suffix with Parameter
		return buildParameterName(parts, converter)

	case ContextRequestBody:
		// Use operationId if available for nicer names, but only for the direct request body schema
		// (not nested items, properties, etc.)
		if s.OperationID != "" && isDirectBodySchema(path) {
			return buildOperationRequestName(s.OperationID, s.ContentType, converter, contentTypeNamer)
		}
		return buildRequestBodyName(parts, converter)

	case ContextResponse:
		// Use operationId if available for nicer names, but only for the direct response schema
		// (not nested items, properties, etc.)
		if s.OperationID != "" && isDirectBodySchema(path) {
			// Extract status code from path
			statusCode := extractStatusCode(parts)
			return buildOperationResponseName(s.OperationID, statusCode, s.ContentType, converter, contentTypeNamer)
		}
		return buildResponseName(parts, converter)

	case ContextHeader:
		return buildHeaderName(parts, converter)

	case ContextCallback:
		return buildCallbackName(parts, converter)

	case ContextWebhook:
		return buildWebhookName(parts, converter)

	default:
		// Fallback: join all meaningful parts
		return buildFallbackName(path, converter)
	}
}

// parsePathContext determines the schema context and extracts relevant path parts.
func parsePathContext(path SchemaPath) (SchemaContext, []string) {
	if len(path) == 0 {
		return ContextUnknown, nil
	}

	switch path[0] {
	case "components":
		if len(path) >= 3 && path[1] == "schemas" {
			return ContextComponentSchema, path[2:]
		}
		if len(path) >= 3 && path[1] == "parameters" {
			return ContextParameter, path[2:]
		}
		if len(path) >= 3 && path[1] == "requestBodies" {
			return ContextRequestBody, path[2:]
		}
		if len(path) >= 3 && path[1] == "responses" {
			return ContextResponse, path[2:]
		}
		if len(path) >= 3 && path[1] == "headers" {
			return ContextHeader, path[2:]
		}
		if len(path) >= 3 && path[1] == "callbacks" {
			return ContextCallback, path[2:]
		}

	case "paths":
		// paths/{path}/{method}/...
		if len(path) >= 3 {
			remaining := path[3:] // skip paths, {path}, {method}
			return detectPathsContext(remaining), path[1:] // include path and method
		}

	case "webhooks":
		return ContextWebhook, path[1:]
	}

	return ContextUnknown, path
}

// detectPathsContext determines context from within a path item.
func detectPathsContext(remaining SchemaPath) SchemaContext {
	if len(remaining) == 0 {
		return ContextUnknown
	}

	switch remaining[0] {
	case "parameters":
		return ContextParameter
	case "requestBody":
		return ContextRequestBody
	case "responses":
		return ContextResponse
	case "callbacks":
		return ContextCallback
	}

	return ContextUnknown
}

// buildComponentSchemaName builds a name for a component schema.
// e.g., ["Cat"] -> "Cat", ["Cat", "properties", "name"] -> "CatName"
func buildComponentSchemaName(parts []string, converter *NameConverter) string {
	if len(parts) == 0 {
		return "Schema"
	}

	var nameParts []string
	nameParts = append(nameParts, parts[0]) // schema name

	// Track trailing structural elements (only add suffix if they're at the end)
	trailingSuffix := ""

	// Process nested parts
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		switch part {
		case "properties":
			// Skip, but next part (property name) will be added
			// Clear trailing suffix since we're going deeper
			trailingSuffix = ""
			continue
		case "items":
			// Accumulate Item suffix (for nested arrays)
			trailingSuffix += "Item"
			continue
		case "additionalProperties":
			// Set Value suffix
			trailingSuffix = "Value"
			continue
		case "allOf", "anyOf", "oneOf":
			// Include the composition type and index
			trailingSuffix = "" // Clear since we're adding meaningful content
			if i+1 < len(parts) {
				nameParts = append(nameParts, part+parts[i+1])
				i++ // Skip the index
			} else {
				nameParts = append(nameParts, part)
			}
		case "not", "prefixItems", "contains", "if", "then", "else":
			// Include these structural keywords
			trailingSuffix = ""
			nameParts = append(nameParts, part)
		default:
			// Include meaningful parts (property names, indices)
			// Clear trailing suffix since we have a meaningful name part
			trailingSuffix = ""
			nameParts = append(nameParts, part)
		}
	}

	name := converter.ToTypeName(strings.Join(nameParts, "_"))

	// Add trailing structural suffix if still present
	if trailingSuffix != "" {
		name += trailingSuffix
	}

	return name
}

// buildParameterName builds a name for a parameter schema.
func buildParameterName(parts []string, converter *NameConverter) string {
	// parts could be:
	// - from components/parameters: [paramName, "schema"]
	// - from paths: [path, method, "parameters", index, "schema"]

	var baseName string
	if len(parts) >= 2 && parts[0] != "" {
		// Try to extract operation-style name
		baseName = buildOperationName(parts, converter)
	}
	if baseName == "" && len(parts) > 0 {
		baseName = converter.ToTypeName(parts[0])
	}
	if baseName == "" {
		baseName = "Param"
	}

	// Always add Parameter suffix
	if !strings.HasSuffix(baseName, "Parameter") {
		baseName += "Parameter"
	}
	return baseName
}

// buildRequestBodyName builds a name for a request body schema.
func buildRequestBodyName(parts []string, converter *NameConverter) string {
	var baseName string
	if len(parts) >= 2 {
		baseName = buildOperationName(parts, converter)
	}
	if baseName == "" && len(parts) > 0 {
		baseName = converter.ToTypeName(parts[0])
	}
	if baseName == "" {
		baseName = "Request"
	}

	// Always add Request suffix
	if !strings.HasSuffix(baseName, "Request") {
		baseName += "Request"
	}
	return baseName
}

// buildResponseName builds a name for a response schema.
func buildResponseName(parts []string, converter *NameConverter) string {
	var baseName string
	var statusCode string

	if len(parts) >= 4 {
		// paths: [path, method, "responses", code, ...]
		baseName = buildOperationName(parts[:2], converter)
		// Find status code
		for i, p := range parts {
			if p == "responses" && i+1 < len(parts) {
				statusCode = parts[i+1]
				break
			}
		}
	}
	if baseName == "" && len(parts) > 0 {
		baseName = converter.ToTypeName(parts[0])
	}
	if baseName == "" {
		baseName = "Response"
	}

	// Add status code if present
	if statusCode != "" && statusCode != "default" {
		baseName += statusCode
	}

	// Always add Response suffix
	if !strings.HasSuffix(baseName, "Response") {
		baseName += "Response"
	}
	return baseName
}

// buildHeaderName builds a name for a header schema.
func buildHeaderName(parts []string, converter *NameConverter) string {
	if len(parts) == 0 {
		return "Header"
	}
	baseName := converter.ToTypeName(parts[0])
	if !strings.HasSuffix(baseName, "Header") {
		baseName += "Header"
	}
	return baseName
}

// buildCallbackName builds a name for a callback schema.
func buildCallbackName(parts []string, converter *NameConverter) string {
	if len(parts) == 0 {
		return "Callback"
	}
	return converter.ToTypeName(parts[0]) + "Callback"
}

// buildWebhookName builds a name for a webhook schema.
func buildWebhookName(parts []string, converter *NameConverter) string {
	if len(parts) == 0 {
		return "Webhook"
	}
	return converter.ToTypeName(parts[0]) + "Webhook"
}

// buildOperationName builds a name from path and method.
// e.g., ["/pets", "get"] -> "GetPets"
func buildOperationName(parts []string, converter *NameConverter) string {
	if len(parts) < 2 {
		return ""
	}

	pathStr := parts[0]
	method := parts[1]

	// Convert method to title case
	methodName := converter.ToTypeName(method)

	// Convert path to name parts
	// /pets/{petId}/toys -> PetsPetIdToys
	pathName := pathToName(pathStr, converter)

	return methodName + pathName
}

// pathToName converts an API path to a name component.
// e.g., "/pets/{petId}" -> "PetsPetId"
func pathToName(path string, converter *NameConverter) string {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Split by slash
	segments := strings.Split(path, "/")

	var parts []string
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		// Remove braces from path parameters
		seg = strings.TrimPrefix(seg, "{")
		seg = strings.TrimSuffix(seg, "}")
		parts = append(parts, seg)
	}

	return converter.ToTypeName(strings.Join(parts, "_"))
}

// buildFallbackName creates a name from the full path as a last resort.
func buildFallbackName(path SchemaPath, converter *NameConverter) string {
	var parts []string
	for _, p := range path {
		// Skip common structural elements
		switch p {
		case "components", "schemas", "paths", "properties",
			"items", "schema", "content", "application/json":
			continue
		default:
			parts = append(parts, p)
		}
	}

	if len(parts) == 0 {
		return "Schema"
	}

	return converter.ToTypeName(strings.Join(parts, "_"))
}

// buildOperationRequestName builds a name for a request body using the operationId.
// e.g., operationId="addPet" -> "AddPetJSONRequest"
func buildOperationRequestName(operationID, contentType string, converter *NameConverter, contentTypeNamer *ContentTypeShortNamer) string {
	baseName := converter.ToTypeName(operationID)

	// Add content type short name if available
	if contentType != "" && contentTypeNamer != nil {
		baseName += contentTypeNamer.ShortName(contentType)
	}

	return baseName + "Request"
}

// buildOperationResponseName builds a name for a response using the operationId.
// e.g., operationId="findPets", statusCode="200" -> "FindPetsJSONResponse"
// e.g., operationId="findPets", statusCode="404" -> "FindPets404JSONResponse"
// e.g., operationId="findPets", statusCode="default" -> "FindPetsDefaultJSONResponse"
func buildOperationResponseName(operationID, statusCode, contentType string, converter *NameConverter, contentTypeNamer *ContentTypeShortNamer) string {
	baseName := converter.ToTypeName(operationID)

	// Add status code, skipping only for 200 (the common success case)
	if statusCode != "" && statusCode != "200" {
		if statusCode == "default" {
			baseName += "Default"
		} else {
			baseName += statusCode
		}
	}

	// Add content type short name if available
	if contentType != "" && contentTypeNamer != nil {
		baseName += contentTypeNamer.ShortName(contentType)
	}

	return baseName + "Response"
}

// extractStatusCode extracts the HTTP status code from path parts.
// Looks for "responses" followed by the status code.
func extractStatusCode(parts []string) string {
	for i, p := range parts {
		if p == "responses" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// isDirectBodySchema returns true if the schema path represents the direct
// schema of a request body or response (i.e., content/{type}/schema),
// not a nested schema (items, properties, allOf members, etc.).
func isDirectBodySchema(path SchemaPath) bool {
	// Find the position of "schema" after "content"
	schemaIdx := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == "schema" {
			schemaIdx = i
			break
		}
	}
	if schemaIdx == -1 {
		return false
	}

	// Check that "schema" is directly after a content type (content/{type}/schema)
	// and there are no structural elements after it
	if schemaIdx < 2 {
		return false
	}
	if path[schemaIdx-2] != "content" {
		return false
	}

	// If schema is at the end of the path, it's a direct body schema
	return schemaIdx == len(path)-1
}
