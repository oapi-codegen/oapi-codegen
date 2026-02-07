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

// TestSkipExternalRefResolution verifies that we can parse a spec containing
// external $ref references without resolving them, and still generate correct
// code using import mappings. This uses libopenapi's SkipExternalRefResolution
// flag (added in v0.33.4, pb33f/libopenapi#519).
func TestSkipExternalRefResolution(t *testing.T) {
	specData, err := os.ReadFile("test/external_ref/spec.yaml")
	require.NoError(t, err)

	// Parse WITHOUT BasePath or AllowFileReferences — the external spec files
	// won't be read. Instead, we rely on SkipExternalRefResolution to leave
	// external $refs unresolved while still building an iterable model.
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

	// The generated code should contain the Container struct with external type references
	assert.Contains(t, code, "type Container struct")
	assert.Contains(t, code, "ObjectA")
	assert.Contains(t, code, "ObjectB")

	// Should reference the external packages via hashed aliases
	assert.Contains(t, code, "ext_934ff11d")
	assert.Contains(t, code, "ext_b892eff9")

	// Should contain the import declarations
	assert.Contains(t, code, `"github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/test/external_ref/packagea"`)
	assert.Contains(t, code, `"github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/test/external_ref/packageb"`)

	// Should NOT contain "any" as a fallback type for the external refs
	// (which would indicate the refs weren't properly detected)
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ObjectA") && strings.Contains(line, "any") {
			t.Errorf("ObjectA resolved to 'any' instead of external type: %s", line)
		}
		if strings.Contains(line, "ObjectB") && strings.Contains(line, "any") {
			t.Errorf("ObjectB resolved to 'any' instead of external type: %s", line)
		}
	}

	t.Logf("Generated code:\n%s", code)
}
