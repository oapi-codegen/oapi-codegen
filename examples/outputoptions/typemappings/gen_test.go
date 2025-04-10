package typemappings

import (
	"testing"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	var c Client

	t.Run("go-type", func(t *testing.T) {
		// Email should have the overridden go type
		var expected *string
		require.Equal(t, expected, c.Email)
	})

	t.Run("skip-optional-pointer", func(t *testing.T) {
		// Uuid should have no pointer receiver, even though it's an optional field.
		require.Equal(t, types.UUID{}, c.Uuid)
	})

	t.Run("defined-via-alias", func(t *testing.T) {
		// StringAlias should be defined via an alias.
		require.Equal(t, StringAlias(""), "")
	})

	t.Run("not-defined-via-alias", func(t *testing.T) {
		// StringType should not be defined via an alias.
		require.NotEqual(t, StringType(""), "")
	})

}
