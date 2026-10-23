package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
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
// allow-remote-refs-from entries are document URLs without a fragment, or
// "*", so a misconfiguration fails loudly instead of silently allowing
// nothing.
func TestConfigurationValidateAllowRemoteRefsFrom(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
	}
	for _, good := range [][]string{
		nil, {}, {"*"},
		{"https://api.example.com/specs/common.yaml"},
		{"http://localhost:8080/common.yaml", " https://api.example.com/v1/api.yaml "},
	} {
		cfg.AllowRemoteRefsFrom = good
		assert.NoError(t, cfg.Validate(), "%q", good)
	}
	for _, bad := range []string{
		"", "api.example.com", "api.example.com/specs/common.yaml", "/specs/common.yaml",
		"https://api.example.com/specs/common.yaml#/components/schemas/X",
	} {
		cfg.AllowRemoteRefsFrom = []string{bad}
		err := cfg.Validate()
		require.Error(t, err, "%q", bad)
		assert.Contains(t, err.Error(), "allow-remote-refs-from")
	}
}

// TestConfigurationUpdateDefaultsAllowRemoteRefsFrom verifies the 2.8 release
// line default: an omitted allow-remote-refs-from becomes ["*"], while an
// explicit empty list in the YAML stays empty and refuses remote references.
func TestConfigurationUpdateDefaultsAllowRemoteRefsFrom(t *testing.T) {
	var omitted Configuration
	require.NoError(t, yaml.Unmarshal([]byte("package: api\n"), &omitted))
	assert.Equal(t, []string{"*"}, omitted.UpdateDefaults().AllowRemoteRefsFrom)

	var empty Configuration
	require.NoError(t, yaml.Unmarshal([]byte("package: api\nallow-remote-refs-from: []\n"), &empty))
	require.NotNil(t, empty.AllowRemoteRefsFrom)
	assert.Empty(t, empty.UpdateDefaults().AllowRemoteRefsFrom)

	var listed Configuration
	require.NoError(t, yaml.Unmarshal([]byte("package: api\nallow-remote-refs-from: [api.example.com]\n"), &listed))
	assert.Equal(t, []string{"api.example.com"}, listed.UpdateDefaults().AllowRemoteRefsFrom)
}
