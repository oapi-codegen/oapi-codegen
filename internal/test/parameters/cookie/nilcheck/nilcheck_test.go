package parameterscookienilcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issue #2238: a nil optional array cookie param must not be sent; a non-nil one must be.
func TestNewGetTestRequest(t *testing.T) {
	t.Run("nil cookie array param is not sent", func(t *testing.T) {
		params := GetTestParams{
			Tags: nil,
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Cookies())
	})

	t.Run("non-nil cookie array param is sent", func(t *testing.T) {
		params := GetTestParams{
			Tags: []string{"a", "b"},
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		cookies := req.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "tags", cookies[0].Name)
	})
}
