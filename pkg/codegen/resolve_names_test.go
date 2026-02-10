package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestResolveNames_NoCollisions(t *testing.T) {
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "schemas", "Pet"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "Pet",
		},
		{
			Path:          SchemaPath{"components", "schemas", "Owner"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "Owner",
		},
	}

	result := ResolveNames(schemas)

	assert.Equal(t, "Pet", result["components/schemas/Pet"])
	assert.Equal(t, "Owner", result["components/schemas/Owner"])
}

func TestResolveNames_Issue200_CrossSectionCollisions(t *testing.T) {
	// "Bar" appears in schemas, parameters, responses, requestBodies, headers
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "schemas", "Bar"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "Bar",
		},
		{
			Path:          SchemaPath{"components", "parameters", "Bar"},
			Context:       ContextComponentParameter,
			Schema:        &openapi3.Schema{},
			ComponentName: "Bar",
		},
		{
			Path:          SchemaPath{"components", "responses", "Bar", "content", "application/json"},
			Context:       ContextComponentResponse,
			Schema:        &openapi3.Schema{},
			ComponentName: "Bar",
			ContentType:   "application/json",
		},
		{
			Path:          SchemaPath{"components", "requestBodies", "Bar", "content", "application/json"},
			Context:       ContextComponentRequestBody,
			Schema:        &openapi3.Schema{},
			ComponentName: "Bar",
			ContentType:   "application/json",
		},
		{
			Path:          SchemaPath{"components", "headers", "Bar"},
			Context:       ContextComponentHeader,
			Schema:        &openapi3.Schema{},
			ComponentName: "Bar",
		},
	}

	result := ResolveNames(schemas)

	// Component schema is privileged — keeps bare name
	assert.Equal(t, "Bar", result["components/schemas/Bar"])
	// Others get context suffixes
	assert.Equal(t, "BarParameter", result["components/parameters/Bar"])
	assert.Equal(t, "BarResponse", result["components/responses/Bar/content/application/json"])
	assert.Equal(t, "BarRequestBody", result["components/requestBodies/Bar/content/application/json"])
	assert.Equal(t, "BarHeader", result["components/headers/Bar"])
}

func TestResolveNames_Issue1474_ClientWrapperCollision(t *testing.T) {
	// Schema named "CreateChatCompletionResponse" collides with
	// client wrapper for operation "createChatCompletion" which
	// would generate "CreateChatCompletionResponse".
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "schemas", "CreateChatCompletionResponse"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "CreateChatCompletionResponse",
		},
		{
			Path:        SchemaPath{"paths", "/chat/completions", "POST", "x-client-response-wrapper"},
			Context:     ContextClientResponseWrapper,
			OperationID: "createChatCompletion",
		},
	}

	result := ResolveNames(schemas)

	// Component schema is privileged — keeps its name
	assert.Equal(t, "CreateChatCompletionResponse", result["components/schemas/CreateChatCompletionResponse"])
	// Client wrapper gets a suffix to avoid collision
	wrapperName := result["paths//chat/completions/POST/x-client-response-wrapper"]
	assert.NotEqual(t, "CreateChatCompletionResponse", wrapperName,
		"client wrapper should not collide with component schema")
	assert.Contains(t, wrapperName, "Response",
		"client wrapper should still contain 'Response'")
}

func TestResolveNames_PrivilegedComponentSchema(t *testing.T) {
	// When exactly one collision member is a component schema,
	// it keeps the bare name
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "schemas", "Foo"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "Foo",
		},
		{
			Path:          SchemaPath{"components", "parameters", "Foo"},
			Context:       ContextComponentParameter,
			Schema:        &openapi3.Schema{},
			ComponentName: "Foo",
		},
	}

	result := ResolveNames(schemas)

	assert.Equal(t, "Foo", result["components/schemas/Foo"])
	assert.Equal(t, "FooParameter", result["components/parameters/Foo"])
}

func TestResolveNames_NoComponentSchema_AllGetSuffixes(t *testing.T) {
	// When no member is a component schema, all get suffixed
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "parameters", "Foo"},
			Context:       ContextComponentParameter,
			Schema:        &openapi3.Schema{},
			ComponentName: "Foo",
		},
		{
			Path:          SchemaPath{"components", "responses", "Foo", "content", "application/json"},
			Context:       ContextComponentResponse,
			Schema:        &openapi3.Schema{},
			ComponentName: "Foo",
			ContentType:   "application/json",
		},
	}

	result := ResolveNames(schemas)

	assert.Equal(t, "FooParameter", result["components/parameters/Foo"])
	assert.Equal(t, "FooResponse", result["components/responses/Foo/content/application/json"])
}

func TestResolveNames_NumericFallback(t *testing.T) {
	// Two schemas with same context that can't be disambiguated
	// by context suffix (both are component schemas)
	schemas := []*GatheredSchema{
		{
			Path:          SchemaPath{"components", "schemas", "Foo"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "Foo",
		},
		{
			// Hypothetical: same candidate name from a different path
			// This shouldn't normally happen with real specs, but tests the fallback
			Path:          SchemaPath{"components", "schemas", "foo"},
			Context:       ContextComponentSchema,
			Schema:        &openapi3.Schema{},
			ComponentName: "foo",
		},
	}

	result := ResolveNames(schemas)

	names := make(map[string]bool)
	for _, name := range result {
		names[name] = true
	}
	// Both should have unique names
	assert.Len(t, names, 2, "should have two unique names")
}

func TestContentTypeSuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"application/json", "JSON"},
		{"application/xml", "XML"},
		{"application/x-www-form-urlencoded", "Form"},
		{"text/plain", "Text"},
		{"application/octet-stream", "Binary"},
		{"application/yaml", "YAML"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, contentTypeSuffix(tt.input))
		})
	}
}
