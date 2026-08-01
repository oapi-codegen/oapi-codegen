package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigurationValidateImportMappingKeys verifies that import-mapping
// keys which are JSON pointers are rejected: keys must be the path or URL of
// a $ref'd document, and same-document references can never be remapped.
// Please see https://github.com/oapi-codegen/oapi-codegen/issues/2459
func TestConfigurationValidateImportMappingKeys(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
		ImportMapping: map[string]string{
			"#/components/schemas": "example.com/mymodule/dto",
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JSON pointer")
	assert.Contains(t, err.Error(), `"#/components/schemas"`)

	// Document paths and URLs are fine, as is the current-package mapping.
	cfg.ImportMapping = map[string]string{
		"../common/api.yaml":           "example.com/mymodule/common",
		"https://example.com/api.yaml": "example.com/mymodule/remote",
		"./sibling.yaml":               "-",
	}
	require.NoError(t, cfg.Validate())
}
