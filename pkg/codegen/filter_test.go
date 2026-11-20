package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("include tags", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: OutputOptions{
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
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: OutputOptions{
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
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: OutputOptions{
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
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				EchoServer:   true,
				Client:       true,
				Models:       true,
				EmbeddedSpec: true,
			},
			OutputOptions: OutputOptions{
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

func TestFilterRemovesPathsWithNoOperationsLeft(t *testing.T) {
	const spec = `
openapi: 3.0.3
info:
  title: filter
  version: "1"
paths:
  /kept:
    get:
      operationId: getKept
      tags: [kept]
      responses:
        "200":
          description: ok
  /dropped:
    get:
      operationId: getDropped
      tags: [dropped]
      responses:
        "200":
          description: ok
`
	load := func(t *testing.T) *openapi3.T {
		t.Helper()
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
		assert.NoError(t, err)
		return swagger
	}

	t.Run("by tag", func(t *testing.T) {
		swagger := load(t)
		filterOperationsByTag(swagger, Configuration{OutputOptions: OutputOptions{IncludeTags: []string{"kept"}}})
		assert.Equal(t, []string{"/kept"}, swagger.Paths.Keys())
	})

	t.Run("by operation id", func(t *testing.T) {
		swagger := load(t)
		filterOperationsByOperationID(swagger, Configuration{OutputOptions: OutputOptions{ExcludeOperationIDs: []string{"getDropped"}}})
		assert.Equal(t, []string{"/kept"}, swagger.Paths.Keys())
	})
}
