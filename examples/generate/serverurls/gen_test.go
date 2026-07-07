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
		// The default-pointer keeps its historical name on this server
		// because the enum doesn't collide (no enum value folds to
		// "Default") — the asymmetric rename for #2003 only kicks in
		// when collision is detected.
		serverUrl, err := NewServerUrlTheProductionAPIServer(
			ServerUrlTheProductionAPIServerBasePathVariableDefault,
			ServerUrlTheProductionAPIServerPortVariableDefault,
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
			ServerUrlTheProductionAPIServerPortVariable443,
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

		got, err = NewServerUrlConflictingDefaultEnum(ServerUrlConflictingDefaultEnumPortVariable443)
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

// Regression test for the case-only-different-enum scenario raised in
// PR #2358 review: `enum: [foo, Foo]` produces two values that
// `ucFirst`-fold to the same identifier `Foo`. The synthesizer dedups
// with a numeric suffix (`Foo` and `Foo1`) and the typed default-pointer
// references the post-suffix const that actually exists. Under the old
// codegen this would have emitted two `const ...VariableFoo`
// declarations and failed to compile.
func TestServerUrlCaseOnlyEnumCollision(t *testing.T) {
	t.Run("default-pointer references the post-dedup constant", func(t *testing.T) {
		// The spec sets default: "Foo"; the post-dedup const is
		// ...VariableFoo1, which is exactly what the pointer must
		// resolve to.
		assert.Equal(t,
			ServerUrlCaseOnlyEnumCollisionModeVariable("Foo"),
			ServerUrlCaseOnlyEnumCollisionModeVariableDefault,
		)
		assert.Equal(t,
			ServerUrlCaseOnlyEnumCollisionModeVariableDefault,
			ServerUrlCaseOnlyEnumCollisionModeVariableFoo1,
		)
	})

	t.Run("both case-variant constants are distinct and accepted", func(t *testing.T) {
		got, err := NewServerUrlCaseOnlyEnumCollision(ServerUrlCaseOnlyEnumCollisionModeVariableFoo)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/foo", got)

		got, err = NewServerUrlCaseOnlyEnumCollision(ServerUrlCaseOnlyEnumCollisionModeVariableFoo1)
		require.NoError(t, err)
		assert.Equal(t, "https://api.example.com/Foo", got)
	})
}
