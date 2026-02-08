package codegen

import (
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/require"
)

func TestClientGenerator(t *testing.T) {
	// Read the petstore spec
	specPath := "../../examples/petstore-expanded/petstore-expanded.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read petstore spec")

	// Parse the spec
	doc, err := libopenapi.NewDocument(specData)
	require.NoError(t, err, "Failed to parse petstore spec")

	// Gather schemas to build schema index
	contentTypeMatcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, contentTypeMatcher)
	require.NoError(t, err, "Failed to gather schemas")

	// Compute names for schemas
	converter := NewNameConverter(NameMangling{}, NameSubstitutions{})
	contentTypeNamer := NewContentTypeShortNamer(DefaultContentTypeShortNames())
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	// Build schema index - key by Path.String() for component schemas
	schemaIndex := make(map[string]*SchemaDescriptor)
	for _, s := range schemas {
		schemaIndex[s.Path.String()] = s
	}

	// Create param tracker
	paramTracker := NewParamUsageTracker()

	// Gather operations
	ops, err := GatherOperations(doc, paramTracker, NewContentTypeMatcher(DefaultContentTypes()))
	require.NoError(t, err, "Failed to gather operations")
	require.Len(t, ops, 4, "Expected 4 operations")

	// Log operations for debugging
	// Verify we have the expected operations
	operationIDs := make([]string, 0, len(ops))
	for _, op := range ops {
		operationIDs = append(operationIDs, op.GoOperationID)
	}
	t.Logf("Operations: %v", operationIDs)

	// Generate client code
	gen, err := NewClientGenerator(schemaIndex, true, nil)
	require.NoError(t, err, "Failed to create client generator")

	clientCode, err := gen.GenerateClient(ops)
	require.NoError(t, err, "Failed to generate client code")
	require.NotEmpty(t, clientCode, "Generated client code should not be empty")

	t.Logf("Generated client code:\n%s", clientCode)

	// Verify key components are present
	require.Contains(t, clientCode, "type Client struct")
	require.Contains(t, clientCode, "NewClient")
	require.Contains(t, clientCode, "type ClientInterface interface")
	require.Contains(t, clientCode, "FindPets")
	require.Contains(t, clientCode, "AddPet")
	require.Contains(t, clientCode, "DeletePet")
	require.Contains(t, clientCode, "FindPetByID")

	// Verify request builders
	require.Contains(t, clientCode, "NewFindPetsRequest")
	require.Contains(t, clientCode, "NewAddPetRequest")
	require.Contains(t, clientCode, "NewDeletePetRequest")
	require.Contains(t, clientCode, "NewFindPetByIDRequest")

	// Verify SimpleClient
	require.Contains(t, clientCode, "type SimpleClient struct")
	require.Contains(t, clientCode, "NewSimpleClient")
}

func TestClientGenerator_FormEncoded(t *testing.T) {
	// Read the comprehensive spec which includes form-encoded bodies
	specPath := "test/files/comprehensive.yaml"
	specData, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read comprehensive spec")

	doc, err := libopenapi.NewDocument(specData)
	require.NoError(t, err, "Failed to parse comprehensive spec")

	contentTypeMatcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, contentTypeMatcher)
	require.NoError(t, err, "Failed to gather schemas")

	converter := NewNameConverter(NameMangling{}, NameSubstitutions{})
	contentTypeNamer := NewContentTypeShortNamer(DefaultContentTypeShortNames())
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	schemaIndex := make(map[string]*SchemaDescriptor)
	for _, s := range schemas {
		schemaIndex[s.Path.String()] = s
	}

	paramTracker := NewParamUsageTracker()
	ops, err := GatherOperations(doc, paramTracker, contentTypeMatcher)
	require.NoError(t, err, "Failed to gather operations")

	// Verify we have an operation with a form-encoded body
	var hasFormBody bool
	for _, op := range ops {
		for _, body := range op.Bodies {
			if body.IsFormEncoded && body.GenerateTyped {
				hasFormBody = true
				break
			}
		}
	}
	require.True(t, hasFormBody, "Expected at least one operation with a form-encoded typed body")

	// Generate client code
	gen, err := NewClientGenerator(schemaIndex, true, nil)
	require.NoError(t, err, "Failed to create client generator")

	clientCode, err := gen.GenerateClient(ops)
	require.NoError(t, err, "Failed to generate client code")

	t.Logf("Generated client code:\n%s", clientCode)

	// Verify form-encoded body methods reference marshalForm
	require.Contains(t, clientCode, "marshalForm(body)")

	// Verify we generate the form helper when needed
	formHelper, err := generateFormHelper(ops)
	require.NoError(t, err, "Failed to generate form helper")
	require.NotEmpty(t, formHelper, "Form helper should be generated when form-encoded bodies exist")
	require.Contains(t, formHelper, "func marshalForm(")
	require.Contains(t, formHelper, "func marshalFormImpl(")
	require.Contains(t, formHelper, "reflect.Value")

	// Verify it generates WithFormdataBody method
	require.Contains(t, clientCode, "WithFormdataBody")
}

func TestIsSimpleOperation(t *testing.T) {
	tests := []struct {
		name     string
		op       *OperationDescriptor
		expected bool
	}{
		{
			name: "simple operation with single JSON 200 response",
			op: &OperationDescriptor{
				Responses: []*ResponseDescriptor{
					{
						StatusCode: "200",
						Contents: []*ResponseContentDescriptor{
							{ContentType: "application/json", IsJSON: true},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "not simple - multiple success responses",
			op: &OperationDescriptor{
				Responses: []*ResponseDescriptor{
					{
						StatusCode: "200",
						Contents: []*ResponseContentDescriptor{
							{ContentType: "application/json", IsJSON: true},
						},
					},
					{
						StatusCode: "201",
						Contents: []*ResponseContentDescriptor{
							{ContentType: "application/json", IsJSON: true},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "not simple - multiple content types",
			op: &OperationDescriptor{
				Responses: []*ResponseDescriptor{
					{
						StatusCode: "200",
						Contents: []*ResponseContentDescriptor{
							{ContentType: "application/json", IsJSON: true},
							{ContentType: "application/xml", IsJSON: false},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "not simple - no JSON content",
			op: &OperationDescriptor{
				Responses: []*ResponseDescriptor{
					{
						StatusCode: "200",
						Contents: []*ResponseContentDescriptor{
							{ContentType: "text/plain", IsJSON: false},
						},
					},
				},
			},
			expected: false,
		},
		{
			name:     "not simple - no responses",
			op:       &OperationDescriptor{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isSimpleOperation(tc.op)
			if result != tc.expected {
				t.Errorf("isSimpleOperation() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestPathFmt(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/pets", "/pets"},
		{"/pets/{petId}", "/pets/%s"},
		{"/pets/{petId}/photos/{photoId}", "/pets/%s/photos/%s"},
		{"/users/{userId}/posts/{postId}/comments/{commentId}", "/users/%s/posts/%s/comments/%s"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := pathFmt(tc.path)
			if result != tc.expected {
				t.Errorf("pathFmt(%q) = %q, expected %q", tc.path, result, tc.expected)
			}
		})
	}
}
