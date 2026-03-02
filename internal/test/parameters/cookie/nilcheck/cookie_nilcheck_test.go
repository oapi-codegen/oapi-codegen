package cookienilcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilCookieArrayParam(t *testing.T) {
	t.Run("nil array param is not sent", func(t *testing.T) {
		params := GetCookieParams{
			Tags: nil,
		}

		req, err := NewGetCookieRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Cookies())
	})

	t.Run("non-nil array param is sent", func(t *testing.T) {
		params := GetCookieParams{
			Tags: []string{"a", "b"},
		}

		req, err := NewGetCookieRequest("https://localhost", &params)
		require.NoError(t, err)

		cookies := req.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "tags", cookies[0].Name)
	})
}
