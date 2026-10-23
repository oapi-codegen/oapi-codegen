package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalOperationSpec builds a document with one GET /x operation whose
// operationId and responses are set by the caller.
func minimalOperationSpec(operationID string, responseKeys ...string) *openapi3.T {
	responses := openapi3.NewResponses()
	for _, key := range responseKeys {
		desc := "d"
		responses.Set(key, &openapi3.ResponseRef{Value: &openapi3.Response{Description: &desc}})
	}
	op := &openapi3.Operation{OperationID: operationID, Responses: responses}
	paths := openapi3.NewPaths()
	paths.Set("/x", &openapi3.PathItem{Get: op})
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "0"},
		Paths:   paths,
	}
}

func TestValidateSpec_OperationId(t *testing.T) {
	// The raw operationId is copied, quoted, into the generated code as the
	// Echo per-operation middleware key, so quotes are fine but a control
	// character is not.
	require.NoError(t, ValidateSpec(minimalOperationSpec(`x"]...); panic("code-injection"); //`, "200")))

	err := ValidateSpec(minimalOperationSpec("x\nfunc init() {}", "200"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operationId")
}

func TestValidateSpec_ResponseKeys(t *testing.T) {
	for _, key := range []string{"100", "200", "404", "599", "999", "1XX", "2XX", "5XX", "default"} {
		assert.NoError(t, ValidateSpec(minimalOperationSpec("op", key)), key)
	}
	for _, key := range []string{"200 || os.Exit(1) == 0", "2xx", "20", "2000", "abc", "", "Default", "0XX", "6XX"} {
		err := ValidateSpec(minimalOperationSpec("op", key))
		require.Error(t, err, "%q", key)
		assert.Contains(t, err.Error(), "response key", "%q", key)
	}
}

func TestValidateSpec_ComponentNamesWithControlCharacters(t *testing.T) {
	// Component names reach single-line doc comments ("X defines model for
	// <name>.") verbatim, so a newline would end the comment and start code.
	bad := "Name\nfunc init() { panic(\"injected\") }\n//"
	desc := "d"
	schema := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	cases := map[string]*openapi3.Components{
		"schema name":                 {Schemas: openapi3.Schemas{bad: schema}},
		"component parameter name":    {Parameters: openapi3.ParametersMap{bad: {Value: &openapi3.Parameter{Name: "p", In: "query", Schema: schema}}}},
		"component request body name": {RequestBodies: openapi3.RequestBodies{bad: {Value: &openapi3.RequestBody{}}}},
		"component response name":     {Responses: openapi3.ResponseBodies{bad: {Value: &openapi3.Response{Description: &desc}}}},
		"component header name":       {Headers: openapi3.Headers{bad: {Value: &openapi3.Header{}}}},
		"component callback name":     {Callbacks: openapi3.Callbacks{bad: {Value: &openapi3.Callback{}}}},
		"security scheme name":        {SecuritySchemes: openapi3.SecuritySchemes{bad: {Value: openapi3.NewSecurityScheme()}}},
	}
	for what, components := range cases {
		t.Run(what, func(t *testing.T) {
			spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Components: components}
			err := ValidateSpec(spec)
			require.Error(t, err)
			assert.Contains(t, err.Error(), what)
			assert.Contains(t, err.Error(), "newline")
		})
	}

	// Quotes in a component name are fine: the name only ever reaches a
	// comment verbatim, and a Go identifier after sanitization.
	spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"},
		Components: &openapi3.Components{Schemas: openapi3.Schemas{`Quoted"Name`: schema}}}
	require.NoError(t, ValidateSpec(spec))
}

func TestValidateSpec_ResponseHeaderStyle(t *testing.T) {
	// A response header's style is copied into a string literal by the strict
	// server templates, like a parameter's style is, so it is checked the same way.
	desc := "d"
	responses := openapi3.NewResponses()
	responses.Set("200", &openapi3.ResponseRef{Value: &openapi3.Response{
		Description: &desc,
		Headers: openapi3.Headers{"X-Trace": {Value: &openapi3.Header{Parameter: openapi3.Parameter{
			Style:  `simple"); panic("injected`,
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}}}},
	}})
	paths := openapi3.NewPaths()
	paths.Set("/x", &openapi3.PathItem{Get: &openapi3.Operation{OperationID: "op", Responses: responses}})
	spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Paths: paths}

	err := ValidateSpec(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "style of header")
}

func TestValidateServerURLs(t *testing.T) {
	serverSpec := func(server *openapi3.Server) *openapi3.T {
		return &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Servers: openapi3.Servers{server}}
	}

	t.Run("a well-formed server passes", func(t *testing.T) {
		spec := serverSpec(&openapi3.Server{
			URL:        "https://{region}.example.com/{basePath}/{undeclared}",
			Extensions: map[string]any{"x-go-name": "Production"},
			Variables: map[string]*openapi3.ServerVariable{
				"region":   {Enum: []string{"eu", "usEast", "Üst", "8443", "_x"}, Default: "eu"},
				"basePath": {Default: "v1"},
			},
		})
		require.NoError(t, ValidateServerURLs(spec))
	})

	t.Run("x-go-name must be an identifier", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL:        "https://example.com",
			Extensions: map[string]any{"x-go-name": "Prod; func init() { panic(1) }; var _"},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "x-go-name")
	})

	t.Run("variable names become parameters and must be identifiers", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL:       "https://example.com/{api-version}",
			Variables: map[string]*openapi3.ServerVariable{"api-version": {Default: "1"}},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server variable name")
	})

	t.Run("undeclared placeholders become parameters and must be identifiers", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL: `https://example.com/{a string); panic(1); var _ = "}`,
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "undeclared URL placeholder")
	})

	t.Run("enum values are appended to the type name and must be identifier characters", func(t *testing.T) {
		for _, bad := range []string{`x = 1; func init() { panic(1) }; var _`, "us-east", "a b", ""} {
			err := ValidateServerURLs(serverSpec(&openapi3.Server{
				URL:       "https://{region}.example.com",
				Variables: map[string]*openapi3.ServerVariable{"region": {Enum: []string{"eu", bad}, Default: "eu"}},
			}))
			require.Error(t, err, "%q", bad)
			assert.Contains(t, err.Error(), "enum value of server variable")
		}
	})

	t.Run("ValidateSpec itself does not impose the identifier rules", func(t *testing.T) {
		// Hyphenated variable names are fine when server URLs are not generated.
		spec := serverSpec(&openapi3.Server{
			URL:       "https://example.com/{api-version}",
			Variables: map[string]*openapi3.ServerVariable{"api-version": {Default: "1"}},
		})
		require.NoError(t, ValidateSpec(spec))
	})
}
