package issue983

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/require"
)

func TestPR983_NestedEnumTypeNameCollision(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Configuration{
		PackageName: "issue983",
		Generate: codegen.GenerateOptions{
			Models: true,
		},
	}

	code, err := codegen.Generate(swagger, opts)

	// PR 983 reports that two properties "_field" and "field" on the same
	// object both normalize to the enum type name "FooField", producing:
	//   duplicate typename 'FooField' detected, can't auto-rename, please
	//   use x-go-name to specify your own name for one of them
	//
	// Log whatever happens so we can see the current behavior.
	if err != nil {
		t.Logf("Generate returned error: %v", err)
	} else {
		t.Logf("Generate succeeded, code length: %d bytes", len(code))
		t.Logf("Generated code:\n%s", code)
	}

	// The generation should succeed without error.
	require.NoError(t, err, "Generate should succeed — the x-go-name on 'field' should disambiguate the enum type names")
	require.NotEmpty(t, code)
}
