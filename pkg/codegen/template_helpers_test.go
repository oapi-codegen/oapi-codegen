package codegen

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCustomHTTPMethod(t *testing.T) {
	for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"} {
		assert.False(t, isCustomHTTPMethod(m), m)
	}
	// OpenAPI 3.2's fixed `query` field and additionalOperations keys.
	for _, m := range []string{"QUERY", "PURGE", "LINK", "get"} {
		assert.True(t, isCustomHTTPMethod(m), m)
	}
}

func TestHasPrefixHasSuffixTemplateFuncs(t *testing.T) {
	tmpl := template.Must(template.New("t").Funcs(TemplateFunctions).Parse(
		`{{ if hasPrefix . "Get" }}p{{ end }}{{ if hasSuffix . "Response" }}s{{ end }}`))

	for in, want := range map[string]string{
		"GetPetResponse": "ps",
		"GetPet":         "p",
		"PetResponse":    "s",
		"Pet":            "",
	} {
		var buf bytes.Buffer
		require.NoError(t, tmpl.Execute(&buf, in))
		assert.Equal(t, want, buf.String(), in)
	}
}

func TestCustomHTTPMethods(t *testing.T) {
	// Standard-only specs yield nil so templates gated on the result emit
	// nothing and existing generated code stays byte-identical.
	assert.Nil(t, customHTTPMethods([]OperationDefinition{{Method: "GET"}, {Method: "POST"}}))

	got := customHTTPMethods([]OperationDefinition{
		{Method: "QUERY"},
		{Method: "GET"},
		{Method: "PURGE"},
		{Method: "QUERY"},
	})
	assert.Equal(t, []string{"PURGE", "QUERY"}, got)
}
