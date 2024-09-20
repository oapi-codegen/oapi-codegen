package codegen

import (
	"go/format"
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnyOfInlineSchema(t *testing.T) {
	opts := codegen.Configuration{
		PackageName: "api",
		Generate: codegen.GenerateOptions{
			Models:     true,
			EchoServer: true,
			Client:     true,
		},
	}
	swagger, err := util.LoadSwagger("anyof-inline.yaml")
	assert.NoError(t, err)

	// Generate code
	code, err := codegen.Generate(swagger, opts)

	validateGeneratedCode(t, err, code)
}

func TestAnyOfRefSchema(t *testing.T) {
	opts := codegen.Configuration{
		PackageName: "api",
		Generate: codegen.GenerateOptions{
			Models:     true,
			EchoServer: true,
			Client:     true,
		},
	}
	swagger, err := util.LoadSwagger("anyof-ref-schema.yaml")
	require.NoError(t, err)

	// Generate code
	code, err := codegen.Generate(swagger, opts)
	assert.NoError(t, err)

	validateGeneratedCode(t, err, code)
}

func validateGeneratedCode(t *testing.T, err error, code string) {
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")
}
