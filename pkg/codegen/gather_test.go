package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherSchemas_ComponentSchemas(t *testing.T) {
	spec := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Pet": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
				},
				"Owner": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	require.Len(t, schemas, 2)

	// Sorted order: Owner, Pet
	assert.Equal(t, SchemaPath{"components", "schemas", "Owner"}, schemas[0].Path)
	assert.Equal(t, ContextComponentSchema, schemas[0].Context)
	assert.Equal(t, "Owner", schemas[0].ComponentName)

	assert.Equal(t, SchemaPath{"components", "schemas", "Pet"}, schemas[1].Path)
	assert.Equal(t, ContextComponentSchema, schemas[1].Context)
	assert.Equal(t, "Pet", schemas[1].ComponentName)
}

func TestGatherSchemas_ComponentParameters(t *testing.T) {
	spec := &openapi3.T{
		Components: &openapi3.Components{
			Parameters: openapi3.ParametersMap{
				"Limit": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name: "limit",
						In:   "query",
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
						},
					},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	require.Len(t, schemas, 1)
	assert.Equal(t, SchemaPath{"components", "parameters", "Limit"}, schemas[0].Path)
	assert.Equal(t, ContextComponentParameter, schemas[0].Context)
	assert.Equal(t, "Limit", schemas[0].ComponentName)
}

func TestGatherSchemas_ComponentResponses(t *testing.T) {
	spec := &openapi3.T{
		Components: &openapi3.Components{
			Responses: openapi3.ResponseBodies{
				"Error": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
								},
							},
						},
					},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	require.Len(t, schemas, 1)
	assert.Equal(t, SchemaPath{"components", "responses", "Error", "content", "application/json"}, schemas[0].Path)
	assert.Equal(t, ContextComponentResponse, schemas[0].Context)
	assert.Equal(t, "Error", schemas[0].ComponentName)
	assert.Equal(t, "application/json", schemas[0].ContentType)
}

func TestGatherSchemas_ComponentRequestBodies(t *testing.T) {
	spec := &openapi3.T{
		Components: &openapi3.Components{
			RequestBodies: openapi3.RequestBodies{
				"CreatePet": &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
								},
							},
						},
					},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	require.Len(t, schemas, 1)
	assert.Equal(t, SchemaPath{"components", "requestBodies", "CreatePet", "content", "application/json"}, schemas[0].Path)
	assert.Equal(t, ContextComponentRequestBody, schemas[0].Context)
	assert.Equal(t, "CreatePet", schemas[0].ComponentName)
}

func TestGatherSchemas_ComponentHeaders(t *testing.T) {
	spec := &openapi3.T{
		Components: &openapi3.Components{
			Headers: openapi3.Headers{
				"X-Rate-Limit": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
							},
						},
					},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	require.Len(t, schemas, 1)
	assert.Equal(t, SchemaPath{"components", "headers", "X-Rate-Limit"}, schemas[0].Path)
	assert.Equal(t, ContextComponentHeader, schemas[0].Context)
}

func TestGatherSchemas_ClientResponseWrappers(t *testing.T) {
	paths := openapi3.NewPaths()
	paths.Set("/pets", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "listPets",
		},
		Post: &openapi3.Operation{
			OperationID: "createPet",
		},
	})

	spec := &openapi3.T{
		Paths: paths,
	}

	// Without client generation, no wrappers
	opts := Configuration{Generate: GenerateOptions{Client: false}}
	schemas := GatherSchemas(spec, opts)
	assert.Len(t, schemas, 0)

	// With client generation, wrappers are gathered
	opts = Configuration{Generate: GenerateOptions{Client: true}}
	schemas = GatherSchemas(spec, opts)
	assert.Len(t, schemas, 2)

	// Check they're sorted by operationID
	assert.Equal(t, ContextClientResponseWrapper, schemas[0].Context)
	assert.Equal(t, "createPet", schemas[0].OperationID)
	assert.Equal(t, ContextClientResponseWrapper, schemas[1].Context)
	assert.Equal(t, "listPets", schemas[1].OperationID)
}

func TestGatherSchemas_AllSections(t *testing.T) {
	// Spec with "Bar" in schemas, parameters, responses, requestBodies, headers
	// This is the issue #200 scenario (cross-section collision)
	paths := openapi3.NewPaths()
	spec := &openapi3.T{
		Paths: paths,
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Bar": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
				},
			},
			Parameters: openapi3.ParametersMap{
				"Bar": &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name: "Bar",
						In:   "query",
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
						},
					},
				},
			},
			Responses: openapi3.ResponseBodies{
				"Bar": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
								},
							},
						},
					},
				},
			},
			RequestBodies: openapi3.RequestBodies{
				"Bar": &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{Type: &openapi3.Types{"object"}},
								},
							},
						},
					},
				},
			},
			Headers: openapi3.Headers{
				"Bar": &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}},
							},
						},
					},
				},
			},
		},
	}

	opts := Configuration{}
	schemas := GatherSchemas(spec, opts)

	// Should have 5 entries: schema, parameter, response, requestBody, header
	assert.Len(t, schemas, 5)

	// Verify contexts are all different
	contexts := make(map[SchemaContext]bool)
	for _, s := range schemas {
		contexts[s.Context] = true
	}
	assert.True(t, contexts[ContextComponentSchema])
	assert.True(t, contexts[ContextComponentParameter])
	assert.True(t, contexts[ContextComponentResponse])
	assert.True(t, contexts[ContextComponentRequestBody])
	assert.True(t, contexts[ContextComponentHeader])
}
