package issue_52

import (
	_ "embed"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
)

//go:embed spec.yaml
var spec []byte

func TestIssue(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(spec)
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	_, err = codegen.Generate(swagger, "issue_52", opts)
	require.NoError(t, err)
}
