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

