package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodegenContext_Params(t *testing.T) {
	t.Run("empty context has no params", func(t *testing.T) {
		ctx := NewCodegenContext()
		assert.False(t, ctx.HasAnyParams())
		assert.Empty(t, ctx.GetRequiredParamTemplates())
		assert.Empty(t, ctx.GetRequiredParamImports())
	})

	t.Run("records param (style and bind)", func(t *testing.T) {
		ctx := NewCodegenContext()
		ctx.NeedParam("simple", false)

		assert.True(t, ctx.HasAnyParams())
		keys := ctx.RequiredParams()
		assert.Contains(t, keys, "style_simple")
		assert.Contains(t, keys, "bind_simple")
	})

	t.Run("records param with explode", func(t *testing.T) {
		ctx := NewCodegenContext()
		ctx.NeedParam("form", true)

		keys := ctx.RequiredParams()
		assert.Contains(t, keys, "style_form_explode")
		assert.Contains(t, keys, "bind_form_explode")
	})

	t.Run("returns helpers template first", func(t *testing.T) {
		ctx := NewCodegenContext()
		ctx.NeedParam("simple", false)

		templates := ctx.GetRequiredParamTemplates()
		require.NotEmpty(t, templates)
		assert.Equal(t, "helpers", templates[0].Name)
	})

	t.Run("aggregates imports", func(t *testing.T) {
		ctx := NewCodegenContext()
		ctx.NeedParam("simple", false)
		ctx.NeedParam("form", true)

		imports := ctx.GetRequiredParamImports()
		assert.NotEmpty(t, imports)

		// Check that common imports are included
		paths := make([]string, len(imports))
		for i, imp := range imports {
			paths[i] = imp.Path
		}
		assert.Contains(t, paths, "reflect")
		assert.Contains(t, paths, "strings")
	})
}

func TestDefaultParamStyle(t *testing.T) {
	tests := []struct {
		location string
		expected string
	}{
		{"path", "simple"},
		{"header", "simple"},
		{"query", "form"},
		{"cookie", "form"},
		{"unknown", "form"},
	}

	for _, tc := range tests {
		t.Run(tc.location, func(t *testing.T) {
			assert.Equal(t, tc.expected, DefaultParamStyle(tc.location))
		})
	}
}

func TestDefaultParamExplode(t *testing.T) {
	tests := []struct {
		location string
		expected bool
	}{
		{"path", false},
		{"header", false},
		{"query", true},
		{"cookie", true},
		{"unknown", false},
	}

	for _, tc := range tests {
		t.Run(tc.location, func(t *testing.T) {
			assert.Equal(t, tc.expected, DefaultParamExplode(tc.location))
		})
	}
}

func TestValidateParamStyle(t *testing.T) {
	validCases := []struct {
		style    string
		location string
	}{
		{"simple", "path"},
		{"label", "path"},
		{"matrix", "path"},
		{"form", "query"},
		{"spaceDelimited", "query"},
		{"pipeDelimited", "query"},
		{"deepObject", "query"},
		{"simple", "header"},
		{"form", "cookie"},
	}

	for _, tc := range validCases {
		t.Run(tc.style+"_in_"+tc.location, func(t *testing.T) {
			err := ValidateParamStyle(tc.style, tc.location)
			assert.NoError(t, err)
		})
	}

	invalidCases := []struct {
		style    string
		location string
	}{
		{"deepObject", "path"},
		{"matrix", "query"},
		{"label", "header"},
		{"simple", "cookie"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.style+"_in_"+tc.location+"_invalid", func(t *testing.T) {
			err := ValidateParamStyle(tc.style, tc.location)
			assert.Error(t, err)
		})
	}

	t.Run("unknown location", func(t *testing.T) {
		err := ValidateParamStyle("simple", "body")
		assert.Error(t, err)
	})
}
