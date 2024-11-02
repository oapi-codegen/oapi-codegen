package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/openapiv3"
	"github.com/stretchr/testify/assert"
)

// All tests using the Generate function are integration tests and not unit tests. Cannot be moved to subpackages without causing issues
func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("include tags", func(t *testing.T) {
		opts := openapiv3.Configuration{
			PackageName: packageName,
			Generate: openapiv3.GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: openapiv3.OutputOptions{
				IncludeTags: []string{"hippo", "giraffe", "cat"},
			},
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true

		// Get a spec from the test definition in this file:
		swagger, err := loader.LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code, `"/test/:name"`)
		assert.Contains(t, code, `"/cat"`)
	})

	t.Run("exclude tags", func(t *testing.T) {
		opts := openapiv3.Configuration{
			PackageName: packageName,
			Generate: openapiv3.GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: openapiv3.OutputOptions{
				ExcludeTags: []string{"hippo", "giraffe", "cat"},
			},
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true

		// Get a spec from the test definition in this file:
		swagger, err := loader.LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, code, `"/test/:name"`)
		assert.NotContains(t, code, `"/cat"`)
	})
}

func TestFilterOperationsByOperationID(t *testing.T) {
	packageName := "testswagger"
	t.Run("include operation ids", func(t *testing.T) {
		opts := openapiv3.Configuration{
			PackageName: packageName,
			Generate: openapiv3.GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: openapiv3.OutputOptions{
				IncludeOperationIDs: []string{"getCatStatus"},
			},
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true

		// Get a spec from the test definition in this file:
		swagger, err := loader.LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code, `"/test/:name"`)
		assert.Contains(t, code, `"/cat"`)
	})

	t.Run("exclude operation ids", func(t *testing.T) {
		opts := openapiv3.Configuration{
			PackageName: packageName,
			Generate: openapiv3.GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: openapiv3.OutputOptions{
				ExcludeOperationIDs: []string{"getCatStatus"},
			},
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true

		// Get a spec from the test definition in this file:
		swagger, err := loader.LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, code, `"/test/:name"`)
		assert.NotContains(t, code, `"/cat"`)
	})
}
