package issue52

import (
	_ "embed"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/openapiv3"
	"github.com/stretchr/testify/require"
)

//go:embed spec.yaml
var spec []byte

func TestIssue(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(spec)
	require.NoError(t, err)

	opts := openapiv3.Configuration{
		PackageName: "issue52",
		Generate: openapiv3.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	_, err = codegen.Generate(swagger, opts)
	require.NoError(t, err)
}
