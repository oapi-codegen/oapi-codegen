package serverurls

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerUrlTheProductionAPIServer(t *testing.T) {
	t.Run("when no values are provided, it does not error", func(t *testing.T) {
		serverUrl, err := NewServerUrlTheProductionAPIServer("", "", "", "")
		require.NoError(t, err)

		assert.Equal(t, "https://.gigantic-server.com:/", serverUrl)

		// NOTE that ideally this should fail as it doesn't /seem/ to provide a valid URL, but it does seem to be valid
		_, err = url.Parse(serverUrl)
		require.NoError(t, err)
	})

	// TODO:when we validate enums, this will need more testing https://github.com/oapi-codegen/oapi-codegen/issues/2006
	t.Run("when values that are not part of the enum are provided, it does not error", func(t *testing.T) {
		invalidPort := ServerUrlTheProductionAPIServerPortVariable("12345")
		serverUrl, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			ServerUrlTheProductionAPIServerNoDefaultVariable(""),
			invalidPort,
			ServerUrlTheProductionAPIServerUsernameVariableDefault,
		)
		require.NoError(t, err)

		assert.Equal(t, "https://demo.gigantic-server.com:12345/v2", serverUrl)
	})

	t.Run("when default values are provided, it does not error", func(t *testing.T) {
		serverUrl, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			ServerUrlTheProductionAPIServerNoDefaultVariable(""),
			ServerUrlTheProductionAPIServerPortVariableDefault,
			ServerUrlTheProductionAPIServerUsernameVariableDefault,
		)
		require.NoError(t, err)

		assert.Equal(t, "https://demo.gigantic-server.com:8443/v2", serverUrl)
	})
}
