package grab_import_names

import (
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestLineComments(t *testing.T) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	code, err := codegen.Generate(swagger, "grab_import_names", opts)
	require.NoError(t, err)
	require.NotContains(t, code, `"openapi_types"`)
}
