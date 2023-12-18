package grabimportnames

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/v2/pkg/openapi"
)

func TestLineComments(t *testing.T) {
	swagger, err := openapi.LoadOpenAPI("spec.yaml")
	require.NoError(t, err)

	opts := openapi.Configuration{
		PackageName: "grabimportnames",
		Generate: openapi.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	code, err := openapi.Generate(swagger, opts)
	require.NoError(t, err)
	require.NotContains(t, code, `"openapi_types"`)
}
