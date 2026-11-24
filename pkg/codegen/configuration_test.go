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

// TestConfigurationWarningsClientWithoutModels verifies that a `client`
// generator with `generate.models: false` warns, because the emitted code
// references schema types this config never declares (#1567). A warning,
// not an error, because a sibling config emitting `models: true` into the
// same package is a legitimate setup.
func TestConfigurationWarningsClientWithoutModels(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Client: true},
	}
	assert.Contains(t, cfg.Warnings()["client"], "generate.models")

	cfg.Generate.Models = true
	assert.NotContains(t, cfg.Warnings(), "client")
}

// TestConfigurationWarningsStrictWithoutServer verifies that
// `strict-server` without a server generator warns, since the strict
// wrapper adapts to a server interface no selected generator emits (#1558).
func TestConfigurationWarningsStrictWithoutServer(t *testing.T) {
	cfg := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Strict: true, Models: true},
	}
	assert.Contains(t, cfg.Warnings()["strict-server"], "server")

	cfg.Generate.ChiServer = true
	assert.NotContains(t, cfg.Warnings(), "strict-server")
}
