package codegen

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schemaWithExtension builds a minimal *openapi3.T whose single schema
// component carries the supplied extension map.
func schemaWithExtension(ext map[string]any) *openapi3.T {
	s := openapi3.NewSchema()
	s.Type = &openapi3.Types{"string"}
	s.Extensions = ext
	ref := &openapi3.SchemaRef{Value: s}
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "0"},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{"S": ref},
		},
	}
}

func TestValidateSpec_XGoTypeImport_Valid(t *testing.T) {
	cases := []struct {
		name string
		ext  map[string]any
	}{
		{
			name: "identifier alias",
			ext: map[string]any{
				"x-go-type": "mypkg.MyType",
				"x-go-type-import": map[string]any{
					"name": "mypkg",
					"path": "example.com/mymodule/mypkg",
				},
			},
		},
		{
			name: "blank alias",
			ext: map[string]any{
				"x-go-type": "something.T",
				"x-go-type-import": map[string]any{
					"name": "_",
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "dot alias",
			ext: map[string]any{
				"x-go-type": "T",
				"x-go-type-import": map[string]any{
					"name": ".",
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "no alias",
			ext: map[string]any{
				"x-go-type": "pkg.T",
				"x-go-type-import": map[string]any{
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "empty alias",
			ext: map[string]any{
				"x-go-type": "pkg.T",
				"x-go-type-import": map[string]any{
					"name": "",
					"path": "example.com/mod",
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSpec(schemaWithExtension(tc.ext))
			assert.NoError(t, err)
		})
	}
}

func TestValidateSpec_XGoTypeImport_NameInjection(t *testing.T) {
	// GHSA-9c2f-gr95-7wqw: x-go-type-import.name is emitted unquoted into the
	// import block; a newline closes the block and allows arbitrary code injection.
	injected := "_ \"unsafe\"\n)\nfunc init() { panic(\"injected\") }\nvar (\n_ ="
	ext := map[string]any{
		"x-go-type": "os.FileInfo",
		"x-go-type-import": map[string]any{
			"name": injected,
			"path": "ignored",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_NameWithNewline(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "alias\ninjected",
			"path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_NameWithBacktick(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "alias`bad",
			"path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_PathWithControlChar(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "mypkg",
			"path": "example.com/mod\x00injected",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import path")
}

func TestValidateSpec_XGoTypeImport_CaseInsensitiveKeys(t *testing.T) {
	// Matches the case-insensitive key handling in ParseGoImportExtension.
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"Name": "mypkg",
			"Path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	assert.NoError(t, err)
}

// Regression: pre-existing extensions must still be validated.
func TestValidateSpec_XGoName_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-name": "bad\nname",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-name")
}

func TestValidateSpec_XGoType_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T\nfunc init() {}",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type")
}

func TestValidateSpec_XGoTypeName_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-type-name": "Bad\nName",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-name")
}

func TestValidateSpec_XOapiCodegenExtraTags_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-oapi-codegen-extra-tags": map[string]any{
			"bad\ntag": "value",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "x-oapi-codegen-extra-tags"))
}
