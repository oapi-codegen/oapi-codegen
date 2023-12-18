package issue1093

import (
	_ "embed"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/v2/pkg/codegen/openapi"
)

//go:embed child.api.yaml
var spec []byte

func TestIssue(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err := openapi.LoadFromData(spec)
	require.NoError(t, err)

	opts := openapi.Configuration{
		PackageName: "issue1093",
		Generate: openapi.GenerateOptions{
			GinServer:    true,
			Strict:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
		ImportMapping: map[string]string{
			"parent.api.yaml": "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1093/api/parent",
		},
	}

	_, err = openapi.Generate(swagger, opts)
	require.NoError(t, err)
}
