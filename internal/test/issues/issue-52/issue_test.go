package issue52

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/v2/pkg/openapi"
)

//go:embed spec.yaml
var spec []byte

func TestIssue(t *testing.T) {
	swagger, err := openapi.LoadFromData(spec)
	require.NoError(t, err)

	opts := openapi.Configuration{
		PackageName: "issue52",
		Generate: openapi.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	_, err = openapi.Generate(swagger, opts)
	require.NoError(t, err)
}
