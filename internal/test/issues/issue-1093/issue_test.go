package issue1093

import (
	_ "embed"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/openapiv3"
)

//go:embed child.api.yaml
var spec []byte

func TestIssue(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	opts := openapiv3.Configuration{
		PackageName: "issue1093",
		Generate: openapiv3.GenerateOptions{
			GinServer:    true,
			Strict:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
		ImportMapping: map[string]string{
			"parent.api.yaml": "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1093/api/parent",
		},
	}

	_, err = codegen.Generate(swagger, opts)
	require.NoError(t, err)
}
