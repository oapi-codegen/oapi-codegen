package issue2074

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGetClientRequest(t *testing.T) {
	t.Run("when params are nil, it does not error", func(t *testing.T) {
		req, err := NewGetClientRequest("http://localhost:8080", nil)

		require.NoError(t, err)
		assert.NotNil(t, req)
	})

	t.Run("when params is an empty struct, it does not error", func(t *testing.T) {
		req, err := NewGetClientRequest("http://localhost:8080", &GetClientParams{})

		require.NoError(t, err)
		assert.NotNil(t, req)
	})

	t.Run("when ParentTag is set, it does not error", func(t *testing.T) {
		parentTag := "parentTag"

		req, err := NewGetClientRequest("http://localhost:8080", &GetClientParams{
			ParentTag: &parentTag,
		})

		require.NoError(t, err)

		assert.NotNil(t, req)
	})
}
