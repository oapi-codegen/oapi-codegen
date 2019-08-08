package issue_f5_3231

import (
	"io/ioutil"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"github.com/weberr13/oapi-codegen/pkg/codegen"
)

func TestIssue(t *testing.T) {
	// load spec from testdata identified by spec
	bytes, err := ioutil.ReadFile(`spec.yaml`)
	if err != nil {
		t.Fatal(err)
	}
	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(bytes)
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient: true,
		GenerateServer: true,
		GenerateTypes:  true,
		EmbedSpec:      true,
	}

	_, err = codegen.Generate(swagger, "issue_f5_3231", opts)
	require.NoError(t, err)
}
