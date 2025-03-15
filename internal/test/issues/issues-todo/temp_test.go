package issuestodo

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"os"
	"testing"

	invopop_yaml "github.com/invopop/yaml"
)

func TestLoadYaml(t *testing.T) {
	contents, err := os.ReadFile("random-crap.yaml")
	require.NoError(t, err)

	something := struct {
		Paths map[string]any `yaml:"paths"`
	}{}

	t.Run("unmarshal yaml", func(t *testing.T) {
		require.NoError(t, yaml.Unmarshal(contents, &something))
	})

	t.Run("unmarshaln invopop", func(t *testing.T) {
		require.NoError(t, invopop_yaml.Unmarshal(contents, &something))
	})

}
