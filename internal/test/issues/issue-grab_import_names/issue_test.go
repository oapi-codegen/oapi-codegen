package grab_import_names

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
)

func TestLineComments(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Configuration{
		PackageName: "grab_import_names",
		Generate: codegen.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	code, err := codegen.Generate(swagger, opts)
	require.NoError(t, err)
	require.NotContains(t, code, `"openapi_types"`)
}
