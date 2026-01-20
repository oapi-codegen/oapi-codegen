package codegen

import (
	"go/format"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestEcho5CodeGeneration(t *testing.T) {
	// Input vars for code generation:
	packageName := "testswagger"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			Echo5Server:  true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
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

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package testswagger")

	// Check imports
	assert.Contains(t, code, "\"github.com/labstack/echo/v5\"")

	// Check that the interface is generated correctly:
	assert.Contains(t, code, "type ServerInterface interface {")

	// Check that we use PathParam instead of Param
	assert.Contains(t, code, "ctx.PathParam(\"name\")")
	assert.NotContains(t, code, "ctx.Param(\"name\")")

	// Check that the register function is generated correctly:
	assert.Contains(t, code, "func RegisterHandlers(router EchoRouter, si ServerInterface) {")
}
