package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerGenerator(t *testing.T) {
	// Create a simple operation for testing
	ops := []*OperationDescriptor{
		{
			OperationID:   "getUser",
			GoOperationID: "GetUser",
			Method:        "GET",
			Path:          "/users/{userId}",
			Summary:       "Get a user by ID",
			PathParams: []*ParameterDescriptor{
				{
					Name:      "userId",
					GoName:    "UserId",
					Location:  "path",
					Required:  true,
					Style:     "simple",
					Explode:   false,
					TypeDecl:  "string",
					IsStyled:  true,
					StyleFunc: "StyleSimpleParam",
					BindFunc:  "BindSimpleParam",
				},
			},
			QueryParams:    []*ParameterDescriptor{},
			HeaderParams:   []*ParameterDescriptor{},
			CookieParams:   []*ParameterDescriptor{},
			HasBody:        false,
			HasParams:      false,
			ParamsTypeName: "GetUserParams",
		},
		{
			OperationID:   "listUsers",
			GoOperationID: "ListUsers",
			Method:        "GET",
			Path:          "/users",
			Summary:       "List all users",
			PathParams:    []*ParameterDescriptor{},
			QueryParams: []*ParameterDescriptor{
				{
					Name:      "limit",
					GoName:    "Limit",
					Location:  "query",
					Required:  false,
					Style:     "form",
					Explode:   true,
					TypeDecl:  "int",
					IsStyled:  true,
					StyleFunc: "StyleFormExplodeParam",
					BindFunc:  "BindFormExplodeParam",
				},
				{
					Name:      "offset",
					GoName:    "Offset",
					Location:  "query",
					Required:  false,
					Style:     "form",
					Explode:   true,
					TypeDecl:  "int",
					IsStyled:  true,
					StyleFunc: "StyleFormExplodeParam",
					BindFunc:  "BindFormExplodeParam",
				},
			},
			HeaderParams:   []*ParameterDescriptor{},
			CookieParams:   []*ParameterDescriptor{},
			HasBody:        false,
			HasParams:      true,
			ParamsTypeName: "ListUsersParams",
		},
		{
			OperationID:    "createUser",
			GoOperationID:  "CreateUser",
			Method:         "POST",
			Path:           "/users",
			Summary:        "Create a new user",
			PathParams:     []*ParameterDescriptor{},
			QueryParams:    []*ParameterDescriptor{},
			HeaderParams:   []*ParameterDescriptor{},
			CookieParams:   []*ParameterDescriptor{},
			HasBody:        true,
			HasParams:      false,
			ParamsTypeName: "CreateUserParams",
		},
	}

	gen, err := NewServerGenerator()
	require.NoError(t, err)

	t.Run("GenerateInterface", func(t *testing.T) {
		result, err := gen.GenerateInterface(ops)
		require.NoError(t, err)

		t.Log("Generated interface:\n", result)

		// Check interface definition
		assert.Contains(t, result, "type ServerInterface interface")
		assert.Contains(t, result, "GetUser(w http.ResponseWriter, r *http.Request, userId string)")
		assert.Contains(t, result, "ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams)")
		assert.Contains(t, result, "CreateUser(w http.ResponseWriter, r *http.Request)")

		// Check comments
		assert.Contains(t, result, "// Get a user by ID")
		assert.Contains(t, result, "// (GET /users/{userId})")
	})

	t.Run("GenerateHandler", func(t *testing.T) {
		result, err := gen.GenerateHandler(ops)
		require.NoError(t, err)

		t.Log("Generated handler:\n", result)

		// Check Handler functions
		assert.Contains(t, result, "func Handler(si ServerInterface) http.Handler")
		assert.Contains(t, result, "func HandlerWithOptions(si ServerInterface, options StdHTTPServerOptions) http.Handler")
		assert.Contains(t, result, "type ServeMux interface")
		assert.Contains(t, result, "type StdHTTPServerOptions struct")

		// Check route registration
		assert.Contains(t, result, `"GET "+options.BaseURL+"/users/{userId}"`)
		assert.Contains(t, result, `"GET "+options.BaseURL+"/users"`)
		assert.Contains(t, result, `"POST "+options.BaseURL+"/users"`)
	})

	t.Run("GenerateWrapper", func(t *testing.T) {
		result, err := gen.GenerateWrapper(ops)
		require.NoError(t, err)

		t.Log("Generated wrapper:\n", result)

		// Check wrapper struct
		assert.Contains(t, result, "type ServerInterfaceWrapper struct")
		assert.Contains(t, result, "Handler")
		assert.Contains(t, result, "ServerInterface")
		assert.Contains(t, result, "HandlerMiddlewares")
		assert.Contains(t, result, "MiddlewareFunc")
		assert.Contains(t, result, "ErrorHandlerFunc")

		// Check wrapper methods
		assert.Contains(t, result, "func (siw *ServerInterfaceWrapper) GetUser(w http.ResponseWriter, r *http.Request)")
		assert.Contains(t, result, "func (siw *ServerInterfaceWrapper) ListUsers(w http.ResponseWriter, r *http.Request)")
		assert.Contains(t, result, "func (siw *ServerInterfaceWrapper) CreateUser(w http.ResponseWriter, r *http.Request)")

		// Check path parameter extraction
		assert.Contains(t, result, `r.PathValue("userId")`)
		assert.Contains(t, result, "BindSimpleParam")

		// Check query parameter extraction
		assert.Contains(t, result, "var params ListUsersParams")
		assert.Contains(t, result, "BindFormExplodeParam")
	})

	t.Run("GenerateErrors", func(t *testing.T) {
		result, err := gen.GenerateErrors()
		require.NoError(t, err)

		t.Log("Generated errors:\n", result)

		// Check error types
		assert.Contains(t, result, "type UnescapedCookieParamError struct")
		assert.Contains(t, result, "type UnmarshalingParamError struct")
		assert.Contains(t, result, "type RequiredParamError struct")
		assert.Contains(t, result, "type RequiredHeaderError struct")
		assert.Contains(t, result, "type InvalidParamFormatError struct")
		assert.Contains(t, result, "type TooManyValuesForParamError struct")
	})

	t.Run("GenerateParamTypes", func(t *testing.T) {
		result, err := gen.GenerateParamTypes(ops)
		require.NoError(t, err)

		t.Log("Generated param types:\n", result)

		// Check param type for ListUsers
		assert.Contains(t, result, "type ListUsersParams struct")
		assert.Contains(t, result, "Limit")
		assert.Contains(t, result, "Offset")
	})

	t.Run("GenerateServer_StdHTTP", func(t *testing.T) {
		result, err := gen.GenerateServer(ServerTypeStdHTTP, ops)
		require.NoError(t, err)

		t.Log("Generated complete server:\n", result)

		// Should contain all components
		assert.Contains(t, result, "type ServerInterface interface")
		assert.Contains(t, result, "type ServerInterfaceWrapper struct")
		assert.Contains(t, result, "func Handler(si ServerInterface) http.Handler")
		assert.Contains(t, result, "type UnescapedCookieParamError struct")

		// Should be reasonably well-formatted
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 50, "Should generate substantial code")
	})

	t.Run("GenerateServer_Empty", func(t *testing.T) {
		result, err := gen.GenerateServer("", ops)
		require.NoError(t, err)
		assert.Empty(t, result, "Empty server type should produce no output")
	})

	t.Run("GenerateServer_Unsupported", func(t *testing.T) {
		_, err := gen.GenerateServer("unsupported-server", ops)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported server type")
		assert.Contains(t, err.Error(), "unsupported-server")
	})
}

func TestServerGeneratorEmptyOperations(t *testing.T) {
	gen, err := NewServerGenerator()
	require.NoError(t, err)

	ops := []*OperationDescriptor{}

	t.Run("empty operations", func(t *testing.T) {
		iface, err := gen.GenerateInterface(ops)
		require.NoError(t, err)
		assert.Contains(t, iface, "type ServerInterface interface")
	})
}
