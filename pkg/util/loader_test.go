package util

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestYamlMarshal(t *testing.T) {
	r := require.New(t)

	t.Run("check that yaml.v3 workaround for multiline strings with leading spaces still works", func(t *testing.T) {
		spec := openapi3.T{}
		spec.AddOperation("info", "GET", &openapi3.Operation{
			Summary: "Get info",
			Parameters: openapi3.Parameters{
				{
					Value: &openapi3.Parameter{
						Description: "     a\nb",
					},
				},
			},
		})
		spec.Paths.Extensions = map[string]interface{}{}

		contents, err := yamlMarshal(&spec)
		r.NoError(err)

		unmarshalledSpec, err := openapi3.NewLoader().LoadFromData(contents)
		r.NoError(err)
		r.Equal(spec, *unmarshalledSpec)
	})
}
