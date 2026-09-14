package codegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

func TestMarshalInlinedSpecPreservesSchemaRefSiblingsForOpenAPI31AndLater(t *testing.T) {
	for _, version := range []string{"3.1.0", "3.2.0"} {
		t.Run(version, func(t *testing.T) {
			spec := loadExternalSchemaRefSiblingSpec(t, version)

			raw, err := marshalInlinedSpec(spec)
			require.NoError(t, err)

			property := embeddedValueProperty(t, raw)
			require.NotContains(t, property, "$ref")
			require.Equal(t, float64(50), property["minLength"])

			embedded, err := openapi3.NewLoader().LoadFromData(raw)
			require.NoError(t, err)
			valueSchema := embedded.Components.Schemas["Request"].Value.Properties["value"].Value
			require.Equal(t, uint64(50), valueSchema.MinLength)
			require.Error(t, valueSchema.VisitJSON(strings.Repeat("x", 49)))
		})
	}
}

func TestMarshalInlinedSpecDoesNotApplySchemaRefSiblingsToOpenAPI30(t *testing.T) {
	spec := loadExternalSchemaRefSiblingSpec(t, "3.0.3")

	raw, err := marshalInlinedSpec(spec)
	require.NoError(t, err)

	property := embeddedValueProperty(t, raw)
	require.Contains(t, property, "$ref")
	require.NotContains(t, property, "minLength")

	embedded, err := openapi3.NewLoader().LoadFromData(raw)
	require.NoError(t, err)
	valueSchema := embedded.Components.Schemas["Request"].Value.Properties["value"].Value
	require.Zero(t, valueSchema.MinLength)
	require.NoError(t, valueSchema.VisitJSON(strings.Repeat("x", 49)))
}

func loadExternalSchemaRefSiblingSpec(t *testing.T, version string) *openapi3.T {
	t.Helper()

	dir := t.TempDir()
	commonPath := filepath.Join(dir, "common.yaml")
	apiPath := filepath.Join(dir, "api.yaml")

	common := fmt.Sprintf(`openapi: %s
info:
  title: Common
  version: "1.0"
paths: {}
components:
  schemas:
    BaseString:
      type: string
      maxLength: 1000
`, version)
	api := fmt.Sprintf(`openapi: %s
info:
  title: API
  version: "1.0"
paths: {}
components:
  schemas:
    Request:
      type: object
      properties:
        value:
          $ref: './common.yaml#/components/schemas/BaseString'
          minLength: 50
`, version)

	require.NoError(t, os.WriteFile(commonPath, []byte(common), 0o600))
	require.NoError(t, os.WriteFile(apiPath, []byte(api), 0o600))

	spec, err := util.LoadSwagger(apiPath)
	require.NoError(t, err)
	return spec
}

func embeddedValueProperty(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec))
	components := spec["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	request := schemas["Request"].(map[string]any)
	properties := request["properties"].(map[string]any)
	return properties["value"].(map[string]any)
}
