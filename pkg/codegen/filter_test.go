package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("include tags", func(t *testing.T) {
		opts := Options{
			GenerateClient:     true,
			GenerateEchoServer: true,
			GenerateTypes:      true,
			EmbedSpec:          true,
			IncludeTags:        []string{"hippo", "giraffe", "cat"},
		}

		// Get a spec from the test definition in this file:
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, packageName, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code, `"/test/:name"`)
		assert.Contains(t, code, `"/cat"`)
	})

	t.Run("exclude tags", func(t *testing.T) {
		opts := Options{
			GenerateClient:     true,
			GenerateEchoServer: true,
			GenerateTypes:      true,
			EmbedSpec:          true,
			ExcludeTags:        []string{"hippo", "giraffe", "cat"},
		}

		// Get a spec from the test definition in this file:
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, packageName, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, code, `"/test/:name"`)
		assert.NotContains(t, code, `"/cat"`)
	})
}
