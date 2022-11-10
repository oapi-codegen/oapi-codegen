package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	tests := map[string]struct {
		opts OutputOptions
		fn   func(t *testing.T, code string)
	}{
		"include tags": {
			opts: OutputOptions{
				IncludeTags: []string{"hippo", "giraffe", "cat"},
			},
			fn: func(t *testing.T, code string) {
				t.Helper()
				assert.NotContains(t, code, `"/test/:name"`)
				assert.Contains(t, code, `"/cat"`)
			},
		},
		"exclude tags": {
			opts: OutputOptions{
				ExcludeTags: []string{"hippo", "giraffe", "cat"},
			},
			fn: func(t *testing.T, code string) {
				t.Helper()
				assert.Contains(t, code, `"/test/:name"`)
				assert.NotContains(t, code, `"/cat"`)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := Configuration{
				PackageName: packageName,
				Generate: GenerateOptions{
					EchoServer:   true,
					Client:       true,
					Models:       true,
					EmbeddedSpec: true,
				},
				OutputOptions: tt.opts,
			}

			// Get a spec from the test definition in this file:
			swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinition))
			assert.NoError(t, err)

			// Run our code generation:
			code, err := Generate(swagger, opts)
			assert.NoError(t, err)
			assert.NotEmpty(t, code)
			tt.fn(t, code)
		})
	}
}
