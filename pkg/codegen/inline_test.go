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

func TestMarshalInlinedSpecPreservesSchemaRefSiblingsWithoutOrigin(t *testing.T) {
	spec := loadSchemaFromData(t, `openapi: 3.1.0
info:
  title: API
  version: "1.0"
paths: {}
components:
  schemas:
    Base:
      type: string
      maxLength: 100
    Request:
      type: object
      properties:
        value:
          $ref: '#/components/schemas/Base'
          minLength: 50
`)
	valueRef := spec.Components.Schemas["Request"].Value.Properties["value"]
	require.Nil(t, valueRef.Origin)

	raw, err := marshalInlinedSpec(spec)
	require.NoError(t, err)

	property := embeddedValueProperty(t, raw)
	require.NotContains(t, property, "$ref")
	require.Equal(t, float64(50), property["minLength"])
}

func TestMarshalInlinedSpecVisitsOpenAPI31SchemaRefFields(t *testing.T) {
	spec := loadSchemaFromData(t, `openapi: 3.1.0
info:
  title: API
  version: "1.0"
paths: {}
components:
  schemas:
    Base:
      type: string
      maxLength: 100
    Container:
      prefixItems:
        - $ref: '#/components/schemas/Base'
          minLength: 1
      contains:
        $ref: '#/components/schemas/Base'
        minLength: 2
      patternProperties:
        '^x':
          $ref: '#/components/schemas/Base'
          minLength: 3
      dependentSchemas:
        dependency:
          $ref: '#/components/schemas/Base'
          minLength: 4
      propertyNames:
        $ref: '#/components/schemas/Base'
        minLength: 5
      unevaluatedItems:
        $ref: '#/components/schemas/Base'
        minLength: 6
      unevaluatedProperties:
        $ref: '#/components/schemas/Base'
        minLength: 7
      if:
        $ref: '#/components/schemas/Base'
        minLength: 8
      then:
        $ref: '#/components/schemas/Base'
        minLength: 9
      else:
        $ref: '#/components/schemas/Base'
        minLength: 10
      $defs:
        decorated:
          $ref: '#/components/schemas/Base'
          minLength: 11
      contentSchema:
        $ref: '#/components/schemas/Base'
        minLength: 12
`)

	raw, err := marshalInlinedSpec(spec)
	require.NoError(t, err)

	container := embeddedSchema(t, raw, "Container")
	checks := []struct {
		name      string
		value     map[string]any
		minLength float64
	}{
		{"prefixItems", container["prefixItems"].([]any)[0].(map[string]any), 1},
		{"contains", container["contains"].(map[string]any), 2},
		{"patternProperties", container["patternProperties"].(map[string]any)["^x"].(map[string]any), 3},
		{"dependentSchemas", container["dependentSchemas"].(map[string]any)["dependency"].(map[string]any), 4},
		{"propertyNames", container["propertyNames"].(map[string]any), 5},
		{"unevaluatedItems", container["unevaluatedItems"].(map[string]any), 6},
		{"unevaluatedProperties", container["unevaluatedProperties"].(map[string]any), 7},
		{"if", container["if"].(map[string]any), 8},
		{"then", container["then"].(map[string]any), 9},
		{"else", container["else"].(map[string]any), 10},
		{"$defs", container["$defs"].(map[string]any)["decorated"].(map[string]any), 11},
		{"contentSchema", container["contentSchema"].(map[string]any), 12},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			require.NotContains(t, check.value, "$ref")
			require.Equal(t, check.minLength, check.value["minLength"])
		})
	}
}

func TestMarshalInlinedSpecKeepsRefBoundaryForCyclicSchema(t *testing.T) {
	spec := loadSchemaFromData(t, `openapi: 3.1.0
info:
  title: API
  version: "1.0"
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        parent:
          $ref: '#/components/schemas/Node'
          minProperties: 1
`)
	parentRef := spec.Components.Schemas["Node"].Value.Properties["parent"]
	// kin-openapi does not apply sibling fields while resolving a self-reference,
	// so model the shallow-copy shape which reaches this function when the
	// sibling has been applied to a cyclic resolved value.
	resolved := *parentRef.Value
	resolved.MinProps = 1
	parentRef.Value = &resolved

	raw, err := marshalInlinedSpec(spec)
	require.NoError(t, err)

	node := embeddedSchema(t, raw, "Node")
	parent := node["properties"].(map[string]any)["parent"].(map[string]any)
	require.Contains(t, parent, "$ref")
	require.Equal(t, float64(1), parent["minProperties"])
}

func TestMarshalInlinedSpecDoesNotMutateReferencedSchemaExtensions(t *testing.T) {
	spec := loadSchemaFromData(t, `openapi: 3.1.0
info:
  title: API
  version: "1.0"
paths: {}
components:
  schemas:
    Base:
      type: string
      x-component: true
    Request:
      type: object
      properties:
        value:
          $ref: '#/components/schemas/Base'
          minLength: 1
          x-use-site: true
`)
	baseExtensions := spec.Components.Schemas["Base"].Value.Extensions

	raw, err := marshalInlinedSpec(spec)
	require.NoError(t, err)

	require.Equal(t, map[string]any{"x-component": true}, baseExtensions)
	require.NotContains(t, embeddedSchema(t, raw, "Base"), "x-use-site")
	property := embeddedValueProperty(t, raw)
	require.Equal(t, true, property["x-component"])
	require.Equal(t, true, property["x-use-site"])
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
	request := embeddedSchema(t, raw, "Request")
	properties := request["properties"].(map[string]any)
	return properties["value"].(map[string]any)
}

func embeddedSchema(t *testing.T, raw []byte, name string) map[string]any {
	t.Helper()

	var spec map[string]any
	require.NoError(t, json.Unmarshal(raw, &spec))
	components := spec["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	return schemas[name].(map[string]any)
}

func loadSchemaFromData(t *testing.T, data string) *openapi3.T {
	t.Helper()
	spec, err := openapi3.NewLoader().LoadFromData([]byte(data))
	require.NoError(t, err)
	return spec
}
