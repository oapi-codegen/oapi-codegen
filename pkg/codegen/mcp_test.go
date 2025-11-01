package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationToMCPTool(t *testing.T) {
	// Create a simple operation definition
	spec := &openapi3.Operation{
		Summary: "Get a pet by ID",
		Responses: openapi3.NewResponses(
			openapi3.WithStatus(200, &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: openapi3.Schemas{
							"id": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
								},
							},
							"name": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
								},
							},
						},
					}),
				},
			}),
		),
	}

	op := OperationDefinition{
		OperationId: "GetPet",
		Summary:     "Get a pet by ID",
		Method:      "GET",
		Path:        "/pets/{petId}",
		PathParams: []ParameterDefinition{
			{
				ParamName: "petId",
				Required:  true,
				Spec: &openapi3.Parameter{
					Name:        "petId",
					In:          "path",
					Required:    true,
					Description: "The pet ID",
				},
				Schema: Schema{
					GoType: "string",
					OAPISchema: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
		},
		QueryParams: []ParameterDefinition{
			{
				ParamName: "include",
				Required:  false,
				Spec: &openapi3.Parameter{
					Name:        "include",
					In:          "query",
					Required:    false,
					Description: "What to include",
				},
				Schema: Schema{
					GoType: "string",
					OAPISchema: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
		},
		Spec: spec,
	}

	tool, err := operationToMCPTool(op)
	require.NoError(t, err)
	assert.Equal(t, "GetPet", tool.OperationID)
	assert.Equal(t, "Get a pet by ID", tool.Description)
	assert.NotEmpty(t, tool.InputSchema)
	assert.NotEmpty(t, tool.OutputSchema)
}

func TestBuildMCPInputSchema(t *testing.T) {
	tests := []struct {
		name     string
		op       OperationDefinition
		wantKeys []string
	}{
		{
			name: "path parameters only",
			op: OperationDefinition{
				Spec: &openapi3.Operation{},
				PathParams: []ParameterDefinition{
					{
						ParamName: "id",
						Required:  true,
						Spec: &openapi3.Parameter{
							Name:     "id",
							In:       "path",
							Required: true,
						},
						Schema: Schema{
							GoType: "string",
							OAPISchema: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
			},
			wantKeys: []string{"path"},
		},
		{
			name: "multiple parameter types",
			op: OperationDefinition{
				Spec: &openapi3.Operation{},
				PathParams: []ParameterDefinition{
					{
						ParamName: "id",
						Required:  true,
						Spec: &openapi3.Parameter{
							Name:     "id",
							In:       "path",
							Required: true,
						},
						Schema: Schema{
							GoType: "string",
							OAPISchema: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
				QueryParams: []ParameterDefinition{
					{
						ParamName: "filter",
						Required:  false,
						Spec: &openapi3.Parameter{
							Name:     "filter",
							In:       "query",
							Required: false,
						},
						Schema: Schema{
							GoType: "string",
							OAPISchema: &openapi3.Schema{
								Type: &openapi3.Types{"string"},
							},
						},
					},
				},
			},
			wantKeys: []string{"path", "query"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := buildMCPInputSchema(tt.op)
			require.NoError(t, err)
			assert.Equal(t, "object", schema["type"])

			props, ok := schema["properties"].(map[string]any)
			require.True(t, ok)

			for _, key := range tt.wantKeys {
				assert.Contains(t, props, key)
			}
		})
	}
}

func TestBuildMCPOutputSchema(t *testing.T) {
	tests := []struct {
		name    string
		op      OperationDefinition
		wantNil bool
	}{
		{
			name: "with 200 response",
			op: OperationDefinition{
				Spec: &openapi3.Operation{
					Responses: openapi3.NewResponses(
						openapi3.WithStatus(200, &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
									Type: &openapi3.Types{"object"},
									Properties: openapi3.Schemas{
										"id": &openapi3.SchemaRef{
											Value: &openapi3.Schema{
												Type: &openapi3.Types{"string"},
											},
										},
									},
								}),
							},
						}),
					),
				},
			},
			wantNil: false,
		},
		{
			name: "no responses",
			op: OperationDefinition{
				Spec: &openapi3.Operation{},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := buildMCPOutputSchema(tt.op)
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, schema)
			} else {
				assert.NotNil(t, schema)
			}
		})
	}
}

func TestOpenAPISchemaToJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		schema   *openapi3.Schema
		wantType string
	}{
		{
			name: "string type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
			wantType: "string",
		},
		{
			name: "integer type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"integer"},
			},
			wantType: "integer",
		},
		{
			name: "array type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
			wantType: "array",
		},
		{
			name: "object type with properties",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"name": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
				Required: []string{"name"},
			},
			wantType: "object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := openAPISchemaToJSONSchema(tt.schema)
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, result["type"])
		})
	}
}

func TestGetXMCPExtension(t *testing.T) {
	tests := []struct {
		name          string
		op            OperationDefinition
		wantValue     bool
		wantHasExtension bool
	}{
		{
			name: "x-mcp true",
			op: OperationDefinition{
				Spec: &openapi3.Operation{
					Extensions: map[string]interface{}{
						"x-mcp": true,
					},
				},
			},
			wantValue:     true,
			wantHasExtension: true,
		},
		{
			name: "x-mcp false",
			op: OperationDefinition{
				Spec: &openapi3.Operation{
					Extensions: map[string]interface{}{
						"x-mcp": false,
					},
				},
			},
			wantValue:     false,
			wantHasExtension: true,
		},
		{
			name: "no x-mcp extension",
			op: OperationDefinition{
				Spec: &openapi3.Operation{
					Extensions: map[string]interface{}{},
				},
			},
			wantValue:     false,
			wantHasExtension: false,
		},
		{
			name: "nil spec",
			op: OperationDefinition{
				Spec: nil,
			},
			wantValue:     false,
			wantHasExtension: false,
		},
		{
			name: "x-mcp with non-boolean value",
			op: OperationDefinition{
				Spec: &openapi3.Operation{
					Extensions: map[string]interface{}{
						"x-mcp": "true",
					},
				},
			},
			wantValue:     false,
			wantHasExtension: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, hasExtension := getXMCPExtension(tt.op)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantHasExtension, hasExtension)
		})
	}
}

func TestFilterOperationsForMCP(t *testing.T) {
	// Create test operations
	opWithXMCPTrue := OperationDefinition{
		OperationId: "GetResource",
		Spec: &openapi3.Operation{
			Extensions: map[string]interface{}{
				"x-mcp": true,
			},
		},
	}

	opWithXMCPFalse := OperationDefinition{
		OperationId: "CreateResource",
		Spec: &openapi3.Operation{
			Extensions: map[string]interface{}{
				"x-mcp": false,
			},
		},
	}

	opWithoutXMCP := OperationDefinition{
		OperationId: "UpdateResource",
		Spec: &openapi3.Operation{
			Extensions: map[string]interface{}{},
		},
	}

	tests := []struct {
		name          string
		ops           []OperationDefinition
		inclusionMode MCPInclusionMode
		wantOps       []string // operation IDs that should be included
		wantErr       bool
	}{
		{
			name:          "include mode - includes all by default",
			ops:           []OperationDefinition{opWithXMCPTrue, opWithXMCPFalse, opWithoutXMCP},
			inclusionMode: MCPInclusionModeInclude,
			wantOps:       []string{"GetResource", "UpdateResource"}, // excludes only opWithXMCPFalse
			wantErr:       false,
		},
		{
			name:          "include mode (default) - empty string",
			ops:           []OperationDefinition{opWithXMCPTrue, opWithXMCPFalse, opWithoutXMCP},
			inclusionMode: "",
			wantOps:       []string{"GetResource", "UpdateResource"},
			wantErr:       false,
		},
		{
			name:          "exclude mode - excludes all by default",
			ops:           []OperationDefinition{opWithXMCPTrue, opWithXMCPFalse, opWithoutXMCP},
			inclusionMode: MCPInclusionModeExclude,
			wantOps:       []string{"GetResource"}, // includes only opWithXMCPTrue
			wantErr:       false,
		},
		{
			name:          "explicit mode - requires x-mcp on all operations",
			ops:           []OperationDefinition{opWithXMCPTrue, opWithXMCPFalse},
			inclusionMode: MCPInclusionModeExplicit,
			wantOps:       []string{"GetResource"}, // includes only opWithXMCPTrue
			wantErr:       false,
		},
		{
			name:          "explicit mode - error when x-mcp missing",
			ops:           []OperationDefinition{opWithXMCPTrue, opWithoutXMCP},
			inclusionMode: MCPInclusionModeExplicit,
			wantOps:       nil,
			wantErr:       true, // should error because opWithoutXMCP doesn't have x-mcp
		},
		{
			name:          "invalid mode",
			ops:           []OperationDefinition{opWithXMCPTrue},
			inclusionMode: "invalid",
			wantOps:       nil,
			wantErr:       true,
		},
		{
			name:          "include mode - only x-mcp true",
			ops:           []OperationDefinition{opWithXMCPTrue},
			inclusionMode: MCPInclusionModeInclude,
			wantOps:       []string{"GetResource"},
			wantErr:       false,
		},
		{
			name:          "include mode - only x-mcp false",
			ops:           []OperationDefinition{opWithXMCPFalse},
			inclusionMode: MCPInclusionModeInclude,
			wantOps:       []string{}, // excluded
			wantErr:       false,
		},
		{
			name:          "exclude mode - only x-mcp false",
			ops:           []OperationDefinition{opWithXMCPFalse},
			inclusionMode: MCPInclusionModeExclude,
			wantOps:       []string{}, // excluded
			wantErr:       false,
		},
		{
			name:          "exclude mode - only without x-mcp",
			ops:           []OperationDefinition{opWithoutXMCP},
			inclusionMode: MCPInclusionModeExclude,
			wantOps:       []string{}, // excluded by default
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := filterOperationsForMCP(tt.ops, tt.inclusionMode)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check that we got the right number of operations
			assert.Equal(t, len(tt.wantOps), len(filtered), "unexpected number of filtered operations")

			// Check that we got the right operations
			gotIDs := make([]string, len(filtered))
			for i, op := range filtered {
				gotIDs[i] = op.OperationId
			}

			for _, wantID := range tt.wantOps {
				assert.Contains(t, gotIDs, wantID, "expected operation %s to be included", wantID)
			}
		})
	}
}

