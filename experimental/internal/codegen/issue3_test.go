package codegen

import (
	"os"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for https://github.com/oapi-codegen/oapi-codegen-exp/issues/3

// TestApplyDefaults_ExternalTypes verifies that generated ApplyDefaults methods
// do not blindly call .ApplyDefaults() on external types. External types may not
// have this method, and calling it would cause compilation errors.
//
// The fix should use reflection to call ApplyDefaults only if it exists on the
// external type.
func TestApplyDefaults_ExternalTypes(t *testing.T) {
	specData, err := os.ReadFile("test/external_ref/spec.yaml")
	require.NoError(t, err)

	docConfig := datamodel.NewDocumentConfiguration()
	docConfig.SkipExternalRefResolution = true

	doc, err := libopenapi.NewDocumentWithConfiguration(specData, docConfig)
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "externalref",
		ImportMapping: map[string]string{
			"./packagea/spec.yaml": "github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/test/external_ref/packagea",
			"./packageb/spec.yaml": "github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/test/external_ref/packageb",
		},
	}

	code, err := Generate(doc, nil, cfg)
	require.NoError(t, err)

	t.Logf("Generated code:\n%s", code)

	// The Container struct should exist and reference external types
	assert.Contains(t, code, "type Container struct")

	// ApplyDefaults should be generated for Container
	assert.Contains(t, code, "func (s *Container) ApplyDefaults()")

	// The ApplyDefaults method should NOT contain direct calls to .ApplyDefaults()
	// on external types, because we cannot know at codegen time whether external
	// types have this method. Instead, it should use reflection or a type assertion
	// to check if the method exists before calling it.
	//
	// Current (broken) behavior generates:
	//   if s.ObjectA != nil { s.ObjectA.ApplyDefaults() }
	//   if s.ObjectB != nil { s.ObjectB.ApplyDefaults() }
	//
	// This causes compilation errors when external types don't have ApplyDefaults().
	// The fix should use reflection to call it conditionally.
	applyDefaultsSection := extractApplyDefaultsMethod(t, code, "Container")
	require.NotEmpty(t, applyDefaultsSection, "should have found ApplyDefaults method for Container")

	// It should NOT directly call .ApplyDefaults() on external type fields
	assert.NotContains(t, applyDefaultsSection, "s.ObjectA.ApplyDefaults()",
		"should not directly call ApplyDefaults() on external type ObjectA - use reflection instead")
	assert.NotContains(t, applyDefaultsSection, "s.ObjectB.ApplyDefaults()",
		"should not directly call ApplyDefaults() on external type ObjectB - use reflection instead")

	// It SHOULD use reflection to conditionally call ApplyDefaults
	assert.Contains(t, code, "reflect", "should import reflect package for calling ApplyDefaults on external types")
}

// TestRefProperties_NoRedeclaration verifies that response models composed of
// properties with $ref do not cause type redeclarations.
func TestRefProperties_NoRedeclaration(t *testing.T) {
	const spec = `openapi: "3.1.0"
info:
  title: Ref Redeclaration Test
  version: "1.0"
paths:
  /order:
    get:
      operationId: getOrder
      tags:
        - orders
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  buyer:
                    $ref: '#/components/schemas/Person'
                  seller:
                    $ref: '#/components/schemas/Person'
                  item:
                    $ref: '#/components/schemas/Item'
components:
  schemas:
    Person:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
    Item:
      type: object
      properties:
        title:
          type: string
        price:
          type: number
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
	}

	code, err := Generate(doc, nil, cfg)
	require.NoError(t, err)

	t.Logf("Generated code:\n%s", code)

	// Person and Item component schemas should each be declared exactly once
	personCount := strings.Count(code, "type Person struct")
	assert.Equal(t, 1, personCount,
		"Person struct should be declared exactly once, got %d", personCount)

	itemCount := strings.Count(code, "type Item struct")
	assert.Equal(t, 1, itemCount,
		"Item struct should be declared exactly once, got %d", itemCount)

	// The response type for getOrder should exist
	// (the exact name depends on naming conventions, but should contain "GetOrder" and "Response")
	assert.Contains(t, code, "GetOrder")

	// The code should compile (basic check: no duplicate type declarations)
	// Count all "type X struct" declarations
	typeDecls := countTypeDeclarations(code)
	for typeName, count := range typeDecls {
		assert.Equal(t, 1, count,
			"type %s should be declared exactly once, got %d", typeName, count)
	}
}

// TestRefProperties_SharedSchemaAcrossEndpoints verifies that when multiple
// endpoints reference the same schema, it's only declared once.
func TestRefProperties_SharedSchemaAcrossEndpoints(t *testing.T) {
	const spec = `openapi: "3.1.0"
info:
  title: Shared Schema Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  users:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      operationId: getUser
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
	}

	code, err := Generate(doc, nil, cfg)
	require.NoError(t, err)

	t.Logf("Generated code:\n%s", code)

	// User struct should be declared exactly once even though it's referenced
	// from multiple endpoints
	userCount := strings.Count(code, "type User struct")
	assert.Equal(t, 1, userCount,
		"User struct should be declared exactly once, got %d", userCount)
}

// extractApplyDefaultsMethod extracts the ApplyDefaults method body for a given type name.
func extractApplyDefaultsMethod(t *testing.T, code, typeName string) string {
	t.Helper()

	// Look for "func (s *TypeName) ApplyDefaults()" and extract until the closing "}"
	marker := "func (s *" + typeName + ") ApplyDefaults()"
	idx := strings.Index(code, marker)
	if idx == -1 {
		// Also try with "u" receiver for union types
		marker = "func (u *" + typeName + ") ApplyDefaults()"
		idx = strings.Index(code, marker)
	}
	if idx == -1 {
		return ""
	}

	// Find the matching closing brace by counting braces
	rest := code[idx:]
	braceCount := 0
	for i, ch := range rest {
		if ch == '{' {
			braceCount++
		} else if ch == '}' {
			braceCount--
			if braceCount == 0 {
				return rest[:i+1]
			}
		}
	}
	return ""
}

// countTypeDeclarations counts occurrences of "type X struct" declarations.
func countTypeDeclarations(code string) map[string]int {
	counts := make(map[string]int)
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type ") && strings.Contains(line, " struct") {
			// Extract the type name
			parts := strings.Fields(line)
			if len(parts) >= 3 && parts[2] == "struct" {
				counts[parts[1]]++
			}
		}
	}
	return counts
}
