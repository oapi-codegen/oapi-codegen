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

// TestConfigurationValidateAllowRemoteRefsFrom verifies that
// allow-remote-refs-from entries are host names (optionally with a port) or
// "*", not URLs or paths, so a misconfiguration fails loudly instead of
// silently allowing nothing.
func TestConfigurationValidateAllowRemoteRefsFrom(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
	}
	for _, good := range [][]string{nil, {}, {"*"}, {"api.example.com"}, {"localhost:8080", " api.example.com "}} {
		cfg.AllowRemoteRefsFrom = good
		assert.NoError(t, cfg.Validate(), "%q", good)
	}
	for _, bad := range []string{"", "https://api.example.com", "api.example.com/specs", "api.example.com?x", "api.example.com#x"} {
		cfg.AllowRemoteRefsFrom = []string{bad}
		err := cfg.Validate()
		require.Error(t, err, "%q", bad)
		assert.Contains(t, err.Error(), "allow-remote-refs-from")
	}
}
