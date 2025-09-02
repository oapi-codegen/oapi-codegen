package issue1530_test

import (
	"testing"

	issue1530 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1530"
	"github.com/stretchr/testify/require"
)

func TestIssue1530(t *testing.T) {
	httpConfigTypes := []string{
		"another_server",
		"apache_server",
		"web_server",
	}

	for _, configType := range httpConfigTypes {
		t.Run("http-"+configType, func(t *testing.T) {
			saveReq := issue1530.ConfigSaveReq{}
			err := saveReq.FromConfigHttp(issue1530.ConfigHttp{
				ConfigType: configType,
				Host:       "example.com",
			})
			require.NoError(t, err)

			cfg, err := saveReq.AsConfigHttp()
			require.NoError(t, err)
			require.Equal(t, configType, cfg.ConfigType)

			cfgByDiscriminator, err := saveReq.ValueByDiscriminator()
			require.NoError(t, err)
			require.Equal(t, cfg, cfgByDiscriminator)
		})
	}

	t.Run("ssh", func(t *testing.T) {
		saveReq := issue1530.ConfigSaveReq{}
		err := saveReq.FromConfigSsh(issue1530.ConfigSsh{
			ConfigType: "ssh_server",
		})
		require.NoError(t, err)

		cfg, err := saveReq.AsConfigSsh()
		require.NoError(t, err)
		require.Equal(t, "ssh_server", cfg.ConfigType)

		cfgByDiscriminator, err := saveReq.ValueByDiscriminator()
		require.NoError(t, err)
		require.Equal(t, cfg, cfgByDiscriminator)
	})
}
