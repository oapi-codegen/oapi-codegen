package codegen

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// MCPToolDefinition represents an MCP tool generated from an OpenAPI operation
type MCPToolDefinition struct {
	OperationID  string // The operation ID used as the tool name
	Description  string // Tool description from the operation
	InputSchema  string // JSON schema for the tool input (structured with path, query, header, cookie, body)
	OutputSchema string // JSON schema for the tool output
	Operation    OperationDefinition
}

// filterOperationsForMCP filters operations based on the x-mcp extension and inclusion mode
func filterOperationsForMCP(ops []OperationDefinition, inclusionMode string) ([]OperationDefinition, error) {
	// Default mode is "include" if not specified
	if inclusionMode == "" {
		inclusionMode = "include"
	}

	filtered := make([]OperationDefinition, 0, len(ops))

	for _, op := range ops {
		// Check for x-mcp extension
		xMCPValue, hasXMCP := getXMCPExtension(op)

		switch inclusionMode {
		case "include":
			// Include by default unless x-mcp is explicitly false
			if !hasXMCP || xMCPValue {
				filtered = append(filtered, op)
			}
		case "exclude":
			// Exclude by default unless x-mcp is explicitly true
			if hasXMCP && xMCPValue {
				filtered = append(filtered, op)
			}
		case "explicit":
			// Require explicit x-mcp field on all operations
			if !hasXMCP {
				return nil, fmt.Errorf("operation %s is missing required x-mcp extension field (mcp-inclusion-mode is set to 'explicit')", op.OperationId)
			}
			if xMCPValue {
				filtered = append(filtered, op)
			}
		default:
			return nil, fmt.Errorf("invalid mcp-inclusion-mode: %s (valid values: include, exclude, explicit)", inclusionMode)
		}
	}

	return filtered, nil
}

// getXMCPExtension retrieves the x-mcp extension value from an operation
// Returns (value, hasExtension)
func getXMCPExtension(op OperationDefinition) (bool, bool) {
	if op.Spec == nil || op.Spec.Extensions == nil {
		return false, false
	}

	xMCP, ok := op.Spec.Extensions["x-mcp"]
	if !ok {
		return false, false
	}

	// Try to convert to bool
	if boolVal, ok := xMCP.(bool); ok {
		return boolVal, true
	}

	// Extension exists but is not a boolean, treat as false
	return false, true
}

// GenerateMCPServer generates MCP server code from operations
func GenerateMCPServer(t *template.Template, ops []OperationDefinition, inclusionMode string) (string, error) {
	// Filter operations based on x-mcp extension and inclusion mode
	filteredOps, err := filterOperationsForMCP(ops, inclusionMode)
	if err != nil {
		return "", fmt.Errorf("error filtering operations for MCP: %w", err)
	}

	tools := make([]MCPToolDefinition, 0, len(filteredOps))

	for _, op := range filteredOps {
		tool, err := operationToMCPTool(op)
		if err != nil {
			return "", fmt.Errorf("error converting operation %s to MCP tool: %w", op.OperationId, err)
		}
		tools = append(tools, tool)
	}

	toolContext := struct {
		Tools []MCPToolDefinition
	}{
		Tools: tools,
	}

	// Generate the generic ServerInterface
	interfaceCode, err := GenerateTemplates([]string{"mcp/mcp-interface.tmpl"}, t, filteredOps)
	if err != nil {
		return "", fmt.Errorf("error generating MCP interface: %w", err)
	}

	// Generate the registration function
	registerCode, err := GenerateTemplates([]string{"mcp/mcp-register.tmpl"}, t, toolContext)
	if err != nil {
		return "", fmt.Errorf("error generating MCP register: %w", err)
	}

	return interfaceCode + "\n" + registerCode, nil
}

// GenerateStrictMCPServer generates strict MCP server adapter from operations
func GenerateStrictMCPServer(t *template.Template, ops []OperationDefinition, inclusionMode string) (string, error) {
	// Filter operations based on x-mcp extension and inclusion mode
	filteredOps, err := filterOperationsForMCP(ops, inclusionMode)
	if err != nil {
		return "", fmt.Errorf("error filtering operations for MCP: %w", err)
	}

	// Generate the strict MCP adapter that bridges StrictServerInterface to ServerInterface
	return GenerateTemplates([]string{"strict/strict-mcp.tmpl"}, t, filteredOps)
}

// operationToMCPTool converts an OpenAPI operation to an MCP tool definition
func operationToMCPTool(op OperationDefinition) (MCPToolDefinition, error) {
	tool := MCPToolDefinition{
		OperationID: op.OperationId,
		Description: op.Summary,
		Operation:   op,
	}

	// Build input schema
	inputSchema, err := buildMCPInputSchema(op)
	if err != nil {
		return tool, fmt.Errorf("error building input schema: %w", err)
	}
	inputSchemaJSON, err := json.Marshal(inputSchema)
	if err != nil {
		return tool, fmt.Errorf("error marshaling input schema: %w", err)
	}
	// Escape for use in Go string literal (replace backslashes and quotes)
	tool.InputSchema = strings.ReplaceAll(strings.ReplaceAll(string(inputSchemaJSON), `\`, `\\`), `"`, `\"`)

	// Build output schema
	outputSchema, err := buildMCPOutputSchema(op)
	if err != nil {
		return tool, fmt.Errorf("error building output schema: %w", err)
	}
	if outputSchema != nil {
		outputSchemaJSON, err := json.Marshal(outputSchema)
		if err != nil {
			return tool, fmt.Errorf("error marshaling output schema: %w", err)
		}
		// Escape for use in Go string literal (replace backslashes and quotes)
		tool.OutputSchema = strings.ReplaceAll(strings.ReplaceAll(string(outputSchemaJSON), `\`, `\\`), `"`, `\"`)
	}

	return tool, nil
}

// buildMCPInputSchema creates a structured input schema for an MCP tool
// The schema has separate sections for path, query, header, cookie, and body parameters
func buildMCPInputSchema(op OperationDefinition) (map[string]any, error) {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	properties := schema["properties"].(map[string]any)
	required := []string{}

	// Add path parameters
	if len(op.PathParams) > 0 {
		pathSchema, pathRequired := buildParameterGroupSchema(op.PathParams)
		properties["path"] = pathSchema
		if len(pathRequired) > 0 {
			required = append(required, "path")
		}
	}

	// Add query parameters
	if len(op.QueryParams) > 0 {
		querySchema, _ := buildParameterGroupSchema(op.QueryParams)
		properties["query"] = querySchema
	}

	// Add header parameters
	if len(op.HeaderParams) > 0 {
		headerSchema, _ := buildParameterGroupSchema(op.HeaderParams)
		properties["header"] = headerSchema
	}

	// Add cookie parameters
	if len(op.CookieParams) > 0 {
		cookieSchema, _ := buildParameterGroupSchema(op.CookieParams)
		properties["cookie"] = cookieSchema
	}

	// Add request body
	if op.HasBody() && len(op.Bodies) > 0 {
		// Find the JSON body if available
		var jsonBody *RequestBodyDefinition
		for i := range op.Bodies {
			body := &op.Bodies[i]
			if util.IsMediaTypeJson(body.ContentType) {
				jsonBody = body
				break
			}
		}

		if jsonBody != nil {
			bodySchema, err := schemaToJSONSchema(&jsonBody.Schema)
			if err != nil {
				return nil, fmt.Errorf("error converting body schema: %w", err)
			}
			properties["body"] = bodySchema
			if op.BodyRequired {
				required = append(required, "body")
			}
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// buildParameterGroupSchema builds a JSON schema for a group of parameters (path, query, header, or cookie)
func buildParameterGroupSchema(params []ParameterDefinition) (map[string]any, []string) {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
	properties := schema["properties"].(map[string]any)
	required := []string{}

	for _, param := range params {
		paramSchema, err := schemaToJSONSchema(&param.Schema)
		if err != nil {
			// If we can't convert, create a basic string schema
			paramSchema = map[string]any{"type": "string"}
		}

		// Add description if available
		if param.Spec.Description != "" {
			paramSchema["description"] = param.Spec.Description
		}

		properties[param.ParamName] = paramSchema

		if param.Required {
			required = append(required, param.ParamName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, required
}

// buildMCPOutputSchema extracts the output schema from operation responses
func buildMCPOutputSchema(op OperationDefinition) (map[string]any, error) {
	if op.Spec == nil || op.Spec.Responses == nil {
		return nil, nil
	}

	// Look for success responses (200, 201, 204)
	successCodes := []string{"200", "201", "204"}
	for _, code := range successCodes {
		responseRef := op.Spec.Responses.Value(code)
		if responseRef == nil || responseRef.Value == nil {
			continue
		}

		response := responseRef.Value

		// Look for JSON content
		for mediaType, content := range response.Content {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}

			if content.Schema != nil {
				return openAPISchemaToJSONSchema(content.Schema.Value)
			}
		}
	}

	// No success response with JSON content found
	return nil, nil
}

// schemaToJSONSchema converts a codegen Schema to a JSON Schema map
func schemaToJSONSchema(s *Schema) (map[string]any, error) {
	if s.OAPISchema != nil {
		return openAPISchemaToJSONSchema(s.OAPISchema)
	}

	// Fallback: create a basic schema from the Go type
	schema := map[string]any{}

	// Map Go types to JSON Schema types
	goType := s.GoType
	switch goType {
	case "string", "time.Time":
		schema["type"] = "string"
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		schema["type"] = "integer"
	case "float32", "float64":
		schema["type"] = "number"
	case "bool":
		schema["type"] = "boolean"
	default:
		// For complex types, default to object
		schema["type"] = "object"
	}

	return schema, nil
}

// openAPISchemaToJSONSchema converts an OpenAPI schema to a JSON Schema map
func openAPISchemaToJSONSchema(s *openapi3.Schema) (map[string]any, error) {
	if s == nil {
		return map[string]any{"type": "object"}, nil
	}

	schema := map[string]any{}

	// Type
	if s.Type.Is("") {
		schema["type"] = "object"
	} else if s.Type.Is("array") {
		schema["type"] = "array"
		if s.Items != nil && s.Items.Value != nil {
			items, err := openAPISchemaToJSONSchema(s.Items.Value)
			if err != nil {
				return nil, err
			}
			schema["items"] = items
		}
	} else {
		// Get the first type
		types := s.Type.Slice()
		if len(types) > 0 {
			schema["type"] = types[0]
		}
	}

	// Description
	if s.Description != "" {
		schema["description"] = s.Description
	}

	// Properties (for objects)
	if len(s.Properties) > 0 {
		props := map[string]any{}
		for name, propRef := range s.Properties {
			if propRef.Value != nil {
				propSchema, err := openAPISchemaToJSONSchema(propRef.Value)
				if err != nil {
					return nil, err
				}
				props[name] = propSchema
			}
		}
		schema["properties"] = props
	}

	// Required fields
	if len(s.Required) > 0 {
		schema["required"] = s.Required
	}

	// Enum values
	if len(s.Enum) > 0 {
		schema["enum"] = s.Enum
	}

	// Additional properties
	if s.AdditionalProperties.Has != nil && !*s.AdditionalProperties.Has {
		schema["additionalProperties"] = false
	} else if s.AdditionalProperties.Schema != nil && s.AdditionalProperties.Schema.Value != nil {
		addlProps, err := openAPISchemaToJSONSchema(s.AdditionalProperties.Schema.Value)
		if err != nil {
			return nil, err
		}
		schema["additionalProperties"] = addlProps
	}

	// Format
	if s.Format != "" {
		schema["format"] = s.Format
	}

	// Min/Max for numbers
	if s.Min != nil {
		schema["minimum"] = *s.Min
	}
	if s.Max != nil {
		schema["maximum"] = *s.Max
	}

	// Min/Max length for strings
	if s.MinLength > 0 {
		schema["minLength"] = s.MinLength
	}
	if s.MaxLength != nil {
		schema["maxLength"] = *s.MaxLength
	}

	// Pattern
	if s.Pattern != "" {
		schema["pattern"] = s.Pattern
	}

	return schema, nil
}
