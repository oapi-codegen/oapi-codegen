package serverurls

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerUrlTheProductionAPIServer(t *testing.T) {
	t.Run("when an empty value is provided for an enum-typed variable, it errors", func(t *testing.T) {
		// `port` is enum-typed; the empty string is not in {"443", "8443"}.
		_, err := NewServerUrlTheProductionAPIServer("", "", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("when a value not in the enum is provided, it errors", func(t *testing.T) {
		invalidPort := ServerUrlTheProductionAPIServerPortVariable("12345")
		_, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			invalidPort,
			ServerUrlTheProductionAPIServerUsernameVariableDefault,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("when default values are provided, it does not error", func(t *testing.T) {
		serverUrl, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			ServerUrlTheProductionAPIServerPortVariableDefaultValue,
			ServerUrlTheProductionAPIServerUsernameVariableDefault,
		)
		require.NoError(t, err)

		assert.Equal(t, "https://demo.gigantic-server.com:8443/v2", serverUrl)

		_, err = url.Parse(serverUrl)
		require.NoError(t, err)
	})

	t.Run("a valid non-default enum value is accepted", func(t *testing.T) {
		serverUrl, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			ServerUrlTheProductionAPIServerPortVariableN443,
			ServerUrlTheProductionAPIServerUsernameVariableDefault,
		)
		require.NoError(t, err)

		assert.Equal(t, "https://demo.gigantic-server.com:443/v2", serverUrl)
	})
}

func TestXGoName(t *testing.T) {
	t.Run("x-go-name overrides the auto-generated server name", func(t *testing.T) {
		assert.Equal(t, "https://api.example.com/v2", MyCustomAPIServer)
	})
}

// Regression test for #2003: an `enum` value `default` no longer
// collides with the default-pointer constant — both are emitted and
// the typed default-pointer correctly references the enum constant.
func TestServerUrlConflictingDefaultEnum(t *testing.T) {
	t.Run("the default-pointer references the enum constant for `default`", func(t *testing.T) {
		assert.Equal(t,
			ServerUrlConflictingDefaultEnumPortVariable("default"),
			ServerUrlConflictingDefaultEnumPortVariableDefaultValue,
		)
	})

	t.Run("New… accepts the default and the other enum value, errors on others", func(t *testing.T) {
		got, err := NewServerUrlConflictingDefaultEnum(ServerUrlConflictingDefaultEnumPortVariableDefaultValue)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/default", got)

		got, err = NewServerUrlConflictingDefaultEnum(ServerUrlConflictingDefaultEnumPortVariableN443)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/443", got)

		_, err = NewServerUrlConflictingDefaultEnum("nope")
		require.Error(t, err)
	})
}

// Regression test for #2005: a `{placeholder}` in the URL with no
// matching entry in `variables` is generated as a plain `string`
// parameter so the function returns a usable URL instead of always
// erroring on the trailing `{` / `}` check.
func TestServerUrlUndeclaredPlaceholderServer(t *testing.T) {
	got, err := NewServerUrlUndeclaredPlaceholderServer("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.api.example.com", got)
}
