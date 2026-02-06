package client

import (
	"context"
	"net/http"
	"testing"

	petstore "github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClientTypes verifies all generated types exist and have correct structure.
// If this test compiles, the type generation is correct.
func TestClientTypes(t *testing.T) {
	// Core client types
	var _ *Client
	var _ *SimpleClient
	var _ ClientInterface
	var _ ClientOption
	var _ RequestEditorFn
	var _ HttpRequestDoer

	// Param types
	var _ *FindPetsParams

	// Request body type alias
	var _ addPetJSONRequestBody

	// Error type with generic parameter
	var _ *ClientHttpError[petstore.Error]
}

// TestClientStructure verifies Client struct has expected fields by accessing them.
func TestClientStructure(t *testing.T) {
	c := &Client{}

	// Access fields - compiler validates they exist with correct types
	var _ = c.Server
	var _ = c.Client
	var _ = c.RequestEditors
}

// TestClientImplementsInterface verifies Client implements ClientInterface.
func TestClientImplementsInterface(t *testing.T) {
	var _ ClientInterface = (*Client)(nil)
}

// TestClientInterfaceMethods verifies ClientInterface methods exist with correct signatures.
// The fact that Client implements ClientInterface (tested above) validates all method signatures.
// Here we use method expressions to verify exact signatures without needing an instance.
func TestClientInterfaceMethods(t *testing.T) {
	// Method expressions verify signatures at compile time
	var _ = (*Client).FindPets
	var _ = (*Client).AddPetWithBody
	var _ = (*Client).AddPet
	var _ = (*Client).DeletePet
	var _ = (*Client).FindPetByID
}

// TestSimpleClientMethods verifies SimpleClient methods return typed responses.
func TestSimpleClientMethods(t *testing.T) {
	// Use method expressions to verify signatures without needing an instance
	// Compiler validates return types - these use the petstore package types
	var _ = (*SimpleClient).FindPets
	var _ = (*SimpleClient).AddPet
	var _ = (*SimpleClient).FindPetByID
}

// TestFindPetsParamsStructure verifies param struct fields.
func TestFindPetsParamsStructure(t *testing.T) {
	p := &FindPetsParams{}

	// Access fields - compiler validates they exist with correct types
	var _ = p.Tags
	var _ = p.Limit
}

// TestRequestBodyTypeAlias verifies the type alias points to correct type.
func TestRequestBodyTypeAlias(t *testing.T) {
	// Compiler validates the alias is compatible with petstore.NewPet
	var body addPetJSONRequestBody
	var newPet petstore.NewPet

	// Bidirectional assignment proves they're the same type
	body = newPet
	newPet = body
	_ = body
	_ = newPet
}

// TestPetstoreTypeAliases verifies that short name aliases match their underlying types.
func TestPetstoreTypeAliases(t *testing.T) {
	// Pet alias should be assignable to/from PetSchemaComponent
	var pet petstore.Pet
	var petComponent petstore.PetSchemaComponent
	pet = petComponent
	petComponent = pet
	_ = pet
	_ = petComponent

	// NewPet alias should be assignable to/from NewPetSchemaComponent
	var newPet petstore.NewPet
	var newPetComponent petstore.NewPetSchemaComponent
	newPet = newPetComponent
	newPetComponent = newPet
	_ = newPet
	_ = newPetComponent

	// Error alias should be assignable to/from ErrorSchemaComponent
	var errType petstore.Error
	var errComponent petstore.ErrorSchemaComponent
	errType = errComponent
	errComponent = errType
	_ = errType
	_ = errComponent
}

// TestClientHttpErrorImplementsError verifies ClientHttpError implements error interface.
func TestClientHttpErrorImplementsError(t *testing.T) {
	var _ error = (*ClientHttpError[petstore.Error])(nil)
}

// TestClientHttpErrorStructure verifies error type fields.
func TestClientHttpErrorStructure(t *testing.T) {
	e := &ClientHttpError[petstore.Error]{}

	// Access fields - compiler validates they exist with correct types
	var _ = e.StatusCode
	var _ = e.Body
	var _ = e.RawBody
}

// TestRequestBuilders verifies request builder functions exist with correct signatures.
func TestRequestBuilders(t *testing.T) {
	var _ = NewFindPetsRequest
	var _ = NewAddPetRequestWithBody
	var _ = NewAddPetRequest
	var _ = NewDeletePetRequest
	var _ = NewFindPetByIDRequest
}

// TestNewClientConstructor verifies the constructor works correctly.
func TestNewClientConstructor(t *testing.T) {
	client, err := NewClient("https://api.example.com")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify trailing slash is added
	assert.Equal(t, "https://api.example.com/", client.Server)
	// Verify default http.Client is created
	assert.NotNil(t, client.Client)
}

// TestClientOptions verifies client options work correctly.
func TestClientOptions(t *testing.T) {
	customClient := &http.Client{}

	client, err := NewClient("https://api.example.com",
		WithHTTPClient(customClient),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-Custom", "value")
			return nil
		}),
	)
	require.NoError(t, err)

	assert.Equal(t, customClient, client.Client)
	assert.Len(t, client.RequestEditors, 1)
}

// TestClientHttpErrorMessage verifies the error message format.
func TestClientHttpErrorMessage(t *testing.T) {
	err := &ClientHttpError[petstore.Error]{
		StatusCode: 404,
		Body:       petstore.Error{Code: 404, Message: "Not Found"},
	}

	assert.Contains(t, err.Error(), "404")
}
