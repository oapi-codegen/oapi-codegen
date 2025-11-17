package codegen

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("group by tags", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				StdHTTPServer: true,
				Client:        true,
				Models:        true,
				EmbeddedSpec:  true,
			},
			OutputOptions: OutputOptions{
				GroupByTag: true,
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
		assert.Contains(t, code, "TestAPI")
		assert.Contains(t, code, "CatAPI")
		serverInterface := "type ServerInterface interface {\n\tCatAPI\n\n\tEnumAPI\n\n\tMergeAllOfAPI\n\n\tTestAPI\n}"
		assert.Contains(t, code, serverInterface)
	})

	t.Run("group by tags with include tags filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				StdHTTPServer: true,
				Client:        true,
				Models:        true,
				EmbeddedSpec:  true,
			},
			OutputOptions: OutputOptions{
				GroupByTag:  true,
				IncludeTags: []string{"cat"},
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
		assert.Contains(t, code, "CatAPI")
	})

	t.Run("group by tags with exclude tags filter", func(t *testing.T) {
		opts := Configuration{
			PackageName: packageName,
			Generate: GenerateOptions{
				StdHTTPServer: true,
				Client:        true,
				Models:        true,
				EmbeddedSpec:  true,
			},
			OutputOptions: OutputOptions{
				GroupByTag:  true,
				ExcludeTags: []string{"cat"},
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
		assert.Contains(t, code, "TestAPI")
		assert.NotContains(t, code, "CatAPI")
	})
}
