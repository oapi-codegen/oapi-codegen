package issue1373

import (
	_ "embed"

	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/require"
)

//go:embed spec.yaml
var spec []byte

// Test treatment additionalProperties in mergeOpenapiSchemas()
func TestIssue(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(spec)
	require.NoError(t, err)

	opts := codegen.Configuration{
		PackageName: "issue1373",
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}

	_, err = codegen.Generate(swagger, opts)
	require.NoError(t, err)
}
