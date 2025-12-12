package util

import (
	"bytes"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestYamlMarshal(t *testing.T) {
	spec := openapi3.T{}
	spec.AddOperation("info", "GET", &openapi3.Operation{
		Summary: "Get info",
		Parameters: openapi3.Parameters{
			{
				Value: &openapi3.Parameter{
					Name:        "param1",
					Description: " a\nb",
				},
			},
			{
				Value: &openapi3.Parameter{
					Name:        "param2",
					Description: "   a\n  b\n",
				},
			},
		},
	})
	spec.Paths.Extensions = map[string]interface{}{}

	t.Run("check marshal/unmarshal via workaround: should not modify spec", func(t *testing.T) {
		r := require.New(t)
		contents, err := yamlMarshal(&spec)
		r.NoError(err)

		unmarshalledSpec, err := openapi3.NewLoader().LoadFromData(contents)
		r.NoError(err)
		r.Equal(spec, *unmarshalledSpec)
	})

	t.Run("check that yaml.v3 bug is still present and needs a workaround", func(t *testing.T) {
		r := require.New(t)
		contents, err := yaml.Marshal(&spec)
		r.NoError(err)

		var node yaml.Node
		err = yaml.NewDecoder(bytes.NewReader(contents)).Decode(&node)
		r.Error(err, "expected this to fail due to a bug in go-yaml.v3 library. if it's no longer here, maybe the workaround is no longer needed.")
		r.EqualError(err, "yaml: line 6: did not find expected key")
	})
}
