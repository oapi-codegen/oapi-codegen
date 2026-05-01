package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLPlaceholders(t *testing.T) {
	t.Run("returns nil for a URL with no placeholders", func(t *testing.T) {
		assert.Nil(t, urlPlaceholders("https://api.example.com/v1"))
	})

	t.Run("extracts a single placeholder", func(t *testing.T) {
		got := urlPlaceholders("https://{host}.example.com/v1")
		assert.Len(t, got, 1)
		assert.Contains(t, got, "host")
	})

	t.Run("extracts multiple placeholders and dedupes repeats", func(t *testing.T) {
		got := urlPlaceholders("https://{host}.example.com:{port}/{base}/{host}")
		assert.Len(t, got, 3)
		assert.Contains(t, got, "host")
		assert.Contains(t, got, "port")
		assert.Contains(t, got, "base")
	})

	t.Run("does not span across `/`", func(t *testing.T) {
		// "{a/b}" must not be treated as a single placeholder named "a/b".
		assert.Nil(t, urlPlaceholders("https://{a/b}.example.com"))
	})
}

func TestUsedAndUndeclaredVariables(t *testing.T) {
	srv := ServerObjectDefinition{
		GoName: "ServerUrlExample",
		OAPISchema: &openapi3.Server{
			URL: "https://{host}.example.com:{port}/{path}",
			Variables: map[string]*openapi3.ServerVariable{
				"host":   {Default: "demo"},
				"port":   {Default: "443", Enum: []string{"443", "8443"}},
				"unused": {Default: "x"}, // declared, but not referenced in URL
				// "path" is referenced in URL but not declared
			},
		},
	}

	used := srv.UsedVariables()
	assert.Len(t, used, 2)
	assert.Contains(t, used, "host")
	assert.Contains(t, used, "port")
	assert.NotContains(t, used, "unused", "declared-but-unused must be filtered (#2004)")

	undeclared := srv.UndeclaredPlaceholders()
	assert.Equal(t, []string{"path"}, undeclared, "URL placeholder not in variables must be reported (#2005)")
}

func TestBuildServerURLTypeDefinitions(t *testing.T) {
	t.Run("synthesises one TypeDefinition per enum-typed used variable", func(t *testing.T) {
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL: "https://api.example.com:{port}",
					Variables: map[string]*openapi3.ServerVariable{
						"port": {Default: "443", Enum: []string{"443", "8443"}},
					},
				},
			},
		}
		defs, err := BuildServerURLTypeDefinitions(spec)
		require.NoError(t, err)
		require.Len(t, defs, 1)
		assert.True(t, defs[0].ForceEnumPrefix, "server-URL enum types must keep prefixed identifiers")
		assert.Equal(t, "string", defs[0].Schema.GoType)
		assert.Len(t, defs[0].Schema.EnumValues, 2)
	})

	t.Run("skips non-enum variables", func(t *testing.T) {
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL: "https://{host}.example.com",
					Variables: map[string]*openapi3.ServerVariable{
						"host": {Default: "demo"}, // no enum
					},
				},
			},
		}
		defs, err := BuildServerURLTypeDefinitions(spec)
		require.NoError(t, err)
		assert.Empty(t, defs)
	})

	t.Run("skips declared-but-unused variables (#2004)", func(t *testing.T) {
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL: "https://api.example.com",
					Variables: map[string]*openapi3.ServerVariable{
						"unused": {Default: "443", Enum: []string{"443", "8443"}},
					},
				},
			},
		}
		defs, err := BuildServerURLTypeDefinitions(spec)
		require.NoError(t, err)
		assert.Empty(t, defs)
	})

	t.Run("errors when default is not in enum (#2007)", func(t *testing.T) {
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL:         "https://api.example.com:{port}",
					Description: "Production API server",
					Variables: map[string]*openapi3.ServerVariable{
						"port": {Default: "12345", Enum: []string{"443", "8443"}},
					},
				},
			},
		}
		_, err := BuildServerURLTypeDefinitions(spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port")
		assert.Contains(t, err.Error(), "12345")
	})

	t.Run("an enum value 'default' does not collide with the default-pointer (#2003)", func(t *testing.T) {
		// The fix routes enum values through GenerateEnums (ForceEnumPrefix=true)
		// and renames the default-pointer constant emitted by server-urls.tmpl
		// to ...VariableDefaultValue, so an enum literal "default" no longer
		// produces a duplicate const declaration with the default pointer.
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL: "https://api.example.com:{port}",
					Variables: map[string]*openapi3.ServerVariable{
						"port": {Default: "default", Enum: []string{"default", "443"}},
					},
				},
			},
		}
		defs, err := BuildServerURLTypeDefinitions(spec)
		require.NoError(t, err)
		require.Len(t, defs, 1)
		// Both enum values are present in the synthesized EnumValues
		// without the "Default" identifier alone shadowing the
		// default-pointer constant — that lives outside this map and
		// is named ...VariableDefaultValue in the template.
		assert.Len(t, defs[0].Schema.EnumValues, 2)
	})
}
