package codegen

import (
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const filterTestSpec = `openapi: "3.1.0"
info:
  title: Filter Test API
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      tags:
        - users
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
  /pets:
    get:
      operationId: listPets
      tags:
        - pets
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
  /admin/settings:
    get:
      operationId: getSettings
      tags:
        - admin
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Settings'
    put:
      operationId: updateSettings
      tags:
        - admin
      responses:
        "200":
          description: OK
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
    Pet:
      type: object
      properties:
        name:
          type: string
    Settings:
      type: object
      properties:
        theme:
          type: string
`

func gatherTestOps(t *testing.T) []*OperationDescriptor {
	t.Helper()
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	ctx := NewCodegenContext()
	ops, err := GatherOperations(doc, ctx, NewContentTypeMatcher(DefaultContentTypes()))
	require.NoError(t, err)
	return ops
}

func opIDs(ops []*OperationDescriptor) []string {
	ids := make([]string, len(ops))
	for i, op := range ops {
		ids[i] = op.OperationID
	}
	return ids
}

func TestFilterOperationsByTag_Include(t *testing.T) {
	ops := gatherTestOps(t)
	require.Len(t, ops, 4)

	filtered := FilterOperationsByTag(ops, OutputOptions{
		IncludeTags: []string{"users"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "listUsers")
	assert.NotContains(t, ids, "listPets")
	assert.NotContains(t, ids, "getSettings")
	assert.NotContains(t, ids, "updateSettings")
}

func TestFilterOperationsByTag_Exclude(t *testing.T) {
	ops := gatherTestOps(t)

	filtered := FilterOperationsByTag(ops, OutputOptions{
		ExcludeTags: []string{"admin"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "listUsers")
	assert.Contains(t, ids, "listPets")
	assert.NotContains(t, ids, "getSettings")
	assert.NotContains(t, ids, "updateSettings")
}

func TestFilterOperationsByTag_IncludeMultiple(t *testing.T) {
	ops := gatherTestOps(t)

	filtered := FilterOperationsByTag(ops, OutputOptions{
		IncludeTags: []string{"users", "pets"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "listUsers")
	assert.Contains(t, ids, "listPets")
	assert.NotContains(t, ids, "getSettings")
	assert.Len(t, filtered, 2)
}

func TestFilterOperationsByTag_Empty(t *testing.T) {
	ops := gatherTestOps(t)

	filtered := FilterOperationsByTag(ops, OutputOptions{})
	assert.Len(t, filtered, 4)
}

func TestFilterOperationsByOperationID_Include(t *testing.T) {
	ops := gatherTestOps(t)

	filtered := FilterOperationsByOperationID(ops, OutputOptions{
		IncludeOperationIDs: []string{"listPets"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "listPets")
	assert.Len(t, filtered, 1)
}

func TestFilterOperationsByOperationID_Exclude(t *testing.T) {
	ops := gatherTestOps(t)

	filtered := FilterOperationsByOperationID(ops, OutputOptions{
		ExcludeOperationIDs: []string{"getSettings", "updateSettings"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "listUsers")
	assert.Contains(t, ids, "listPets")
	assert.NotContains(t, ids, "getSettings")
	assert.NotContains(t, ids, "updateSettings")
	assert.Len(t, filtered, 2)
}

func TestFilterOperations_Combined(t *testing.T) {
	ops := gatherTestOps(t)

	// Include only admin, then exclude updateSettings
	filtered := FilterOperations(ops, OutputOptions{
		IncludeTags:         []string{"admin"},
		ExcludeOperationIDs: []string{"updateSettings"},
	})

	ids := opIDs(filtered)
	assert.Contains(t, ids, "getSettings")
	assert.NotContains(t, ids, "updateSettings")
	assert.Len(t, filtered, 1)
}

func TestFilterSchemasByName(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	matcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, matcher, OutputOptions{})
	require.NoError(t, err)

	// Count component schemas
	var componentNames []string
	for _, s := range schemas {
		if len(s.Path) == 3 && s.Path[0] == "components" && s.Path[1] == "schemas" {
			componentNames = append(componentNames, s.Path[2])
		}
	}
	assert.Contains(t, componentNames, "User")
	assert.Contains(t, componentNames, "Pet")
	assert.Contains(t, componentNames, "Settings")

	// Exclude Pet
	filtered := FilterSchemasByName(schemas, []string{"Pet"})

	var filteredNames []string
	for _, s := range filtered {
		if len(s.Path) == 3 && s.Path[0] == "components" && s.Path[1] == "schemas" {
			filteredNames = append(filteredNames, s.Path[2])
		}
	}
	assert.Contains(t, filteredNames, "User")
	assert.NotContains(t, filteredNames, "Pet")
	assert.Contains(t, filteredNames, "Settings")
}

func TestFilterSchemasByName_Empty(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	matcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, matcher, OutputOptions{})
	require.NoError(t, err)

	filtered := FilterSchemasByName(schemas, nil)
	assert.Equal(t, len(schemas), len(filtered))
}

func TestFilterIntegration_GenerateWithIncludeTags(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		Generation: GenerationOptions{
			Client: true,
		},
		OutputOptions: OutputOptions{
			IncludeTags: []string{"users"},
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)
	// Client interface should only contain the included operation
	assert.Contains(t, code, "ListUsers(ctx context.Context")
	assert.NotContains(t, code, "ListPets(ctx context.Context")
	assert.NotContains(t, code, "GetSettings(ctx context.Context")
}

func TestFilterIntegration_GenerateWithExcludeTags(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		Generation: GenerationOptions{
			Client: true,
		},
		OutputOptions: OutputOptions{
			ExcludeTags: []string{"admin"},
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)
	// Client interface should include users and pets but not admin operations
	assert.Contains(t, code, "ListUsers(ctx context.Context")
	assert.Contains(t, code, "ListPets(ctx context.Context")
	assert.NotContains(t, code, "GetSettings(ctx context.Context")
	assert.NotContains(t, code, "UpdateSettings(ctx context.Context")
}

func TestFilterIntegration_GenerateWithExcludeSchemas(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		OutputOptions: OutputOptions{
			ExcludeSchemas: []string{"Pet"},
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)
	// User and Settings types should still exist
	assert.Contains(t, code, "type User struct")
	assert.Contains(t, code, "type Settings struct")
	// Pet type should be excluded
	assert.NotContains(t, code, "type Pet struct")
}

// TestFilterIntegration_IncludeTagsFiltersSchemas verifies that when include-tags is used,
// only schemas referenced by the included operations are generated.
// This is the behavioral test for https://github.com/oapi-codegen/oapi-codegen-exp/issues/3
func TestFilterIntegration_IncludeTagsFiltersSchemas(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		Generation: GenerationOptions{
			Client: true,
		},
		OutputOptions: OutputOptions{
			IncludeTags:              []string{"users"},
			PruneUnreferencedSchemas: true,
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)

	t.Logf("Generated code:\n%s", code)

	// Operations: only ListUsers should be included
	assert.Contains(t, code, "ListUsers(ctx context.Context")
	assert.NotContains(t, code, "ListPets(ctx context.Context")
	assert.NotContains(t, code, "GetSettings(ctx context.Context")

	// Schemas: only User (used by listUsers) should be generated.
	// Pet and Settings are only used by excluded operations and should NOT be generated.
	assert.Contains(t, code, "type User struct")
	assert.NotContains(t, code, "type Pet struct", "Pet schema should not be generated when only 'users' tag is included")
	assert.NotContains(t, code, "type Settings struct", "Settings schema should not be generated when only 'users' tag is included")
}

// TestFilterIntegration_ExcludeTagsFiltersSchemas verifies that when exclude-tags is used,
// schemas only referenced by excluded operations are not generated.
func TestFilterIntegration_ExcludeTagsFiltersSchemas(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		Generation: GenerationOptions{
			Client: true,
		},
		OutputOptions: OutputOptions{
			ExcludeTags:              []string{"admin"},
			PruneUnreferencedSchemas: true,
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)

	t.Logf("Generated code:\n%s", code)

	// Operations: users and pets should be included, admin should not
	assert.Contains(t, code, "ListUsers(ctx context.Context")
	assert.Contains(t, code, "ListPets(ctx context.Context")
	assert.NotContains(t, code, "GetSettings(ctx context.Context")

	// Schemas: User and Pet should be generated, Settings (admin-only) should not
	assert.Contains(t, code, "type User struct")
	assert.Contains(t, code, "type Pet struct")
	assert.NotContains(t, code, "type Settings struct", "Settings schema should not be generated when 'admin' tag is excluded")
}

func TestFilterIntegration_ServerWithIncludeTags(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(filterTestSpec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
		Generation: GenerationOptions{
			Server: ServerTypeStdHTTP,
		},
		OutputOptions: OutputOptions{
			IncludeTags: []string{"pets"},
		},
	}

	code, err := Generate(doc, []byte(filterTestSpec), cfg)
	require.NoError(t, err)
	// Server interface should only contain the included operation
	assert.Contains(t, code, "ListPets(w http.ResponseWriter")
	assert.NotContains(t, code, "ListUsers(w http.ResponseWriter")
	assert.NotContains(t, code, "GetSettings(w http.ResponseWriter")
}
