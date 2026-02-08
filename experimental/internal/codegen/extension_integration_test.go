package codegen

import (
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestExtensionIntegration(t *testing.T) {
	spec := `
openapi: "3.1.0"
info:
  title: Extension Test API
  version: "1.0"
paths: {}
components:
  schemas:
    # Test type-name-override
    MySchema:
      type: object
      x-oapi-codegen-type-name-override: CustomTypeName
      properties:
        id:
          type: string

    # Test type-override at schema level
    ExternalType:
      type: string
      x-oapi-codegen-type-override: "uuid.UUID;github.com/google/uuid"

    # Test field-level extensions
    User:
      type: object
      properties:
        # Test name override
        user_id:
          type: string
          x-oapi-codegen-name-override: UserID
        # Test type override
        created_at:
          type: string
          x-oapi-codegen-type-override: "time.Time;time"
        # Test skip optional pointer
        description:
          type: string
          x-oapi-codegen-skip-optional-pointer: true
        # Test omitempty control
        status:
          type: string
          x-oapi-codegen-omitempty: false
        # Test omitzero
        count:
          type: integer
          x-oapi-codegen-omitzero: true
        # Test order (count should come before status)
        age:
          type: integer
          x-oapi-codegen-order: 1
        name:
          type: string
          x-oapi-codegen-order: 0
        # Test deprecated reason
        old_field:
          type: string
          x-oapi-codegen-deprecated-reason: "Use new_field instead"

    # Test enum with custom var names
    Status:
      type: string
      enum:
        - active
        - inactive
        - pending
      x-oapi-codegen-enum-varnames:
        - Active
        - Inactive
        - Pending
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	cfg := Configuration{
		PackageName: "output",
	}

	code, err := Generate(doc, nil, cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	// Verify type-name-override
	if !strings.Contains(code, "type CustomTypeName") {
		t.Error("Expected CustomTypeName type from type-name-override")
	}

	// Verify type-override at schema level creates alias
	if !strings.Contains(code, "= uuid.UUID") {
		t.Error("Expected type alias to uuid.UUID from type-override")
	}
	if !strings.Contains(code, `"github.com/google/uuid"`) {
		t.Error("Expected uuid import")
	}

	// Verify name override
	if !strings.Contains(code, "UserID") {
		t.Error("Expected UserID field from name-override")
	}

	// Verify type override on field
	if !strings.Contains(code, "time.Time") {
		t.Error("Expected time.Time from field type-override")
	}
	if !strings.Contains(code, `"time"`) {
		t.Error("Expected time import")
	}

	// Verify skip-optional-pointer (description should be string, not *string)
	// The field should appear as just "string", not "*string"
	if strings.Contains(code, "Description *string") || strings.Contains(code, "Description  *string") {
		t.Error("Expected description to not be a pointer due to skip-optional-pointer")
	}
	if !strings.Contains(code, "Description string") && !strings.Contains(code, "Description  string") {
		t.Error("Expected description to be a non-pointer string")
	}

	// Verify omitzero
	if !strings.Contains(code, "omitzero") {
		t.Error("Expected omitzero in struct tags")
	}

	// Verify deprecated reason in doc
	if !strings.Contains(code, "Deprecated:") {
		t.Error("Expected Deprecated: in documentation")
	}

	// Verify enum with custom var names (no collision, so no type prefix).
	// After goimports formatting, columns may be tab-aligned, so check
	// that the constant name and value appear without a type-name prefix.
	if !strings.Contains(code, `= "active"`) || strings.Contains(code, "StatusActive") {
		t.Error("Expected unprefixed Active constant from custom enum var names")
	}
	if !strings.Contains(code, `= "inactive"`) || strings.Contains(code, "StatusInactive") {
		t.Error("Expected unprefixed Inactive constant from custom enum var names")
	}
	if !strings.Contains(code, `= "pending"`) || strings.Contains(code, "StatusPending") {
		t.Error("Expected unprefixed Pending constant from custom enum var names")
	}
}

func TestLegacyExtensionIntegration(t *testing.T) {
	spec := `
openapi: "3.1.0"
info:
  title: Legacy Extension Test API
  version: "1.0"
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        # Test legacy x-go-type
        id:
          type: string
          x-go-type: mypackage.ID
        # Test legacy x-go-name
        user_name:
          type: string
          x-go-name: Username
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	cfg := Configuration{
		PackageName: "output",
	}

	code, err := Generate(doc, nil, cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	t.Logf("Generated code:\n%s", code)

	// Verify legacy x-go-type works
	if !strings.Contains(code, "mypackage.ID") {
		t.Error("Expected mypackage.ID from legacy x-go-type")
	}

	// Verify legacy x-go-name works
	if !strings.Contains(code, "Username") {
		t.Error("Expected Username from legacy x-go-name")
	}
}
