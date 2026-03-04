package issue2238

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGetTestRequest(t *testing.T) {
	t.Run("nil header array param is not sent", func(t *testing.T) {
		params := GetTestParams{
			XTags: nil,
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Header.Values("X-Tags"))
	})

	t.Run("non-nil header array param is sent", func(t *testing.T) {
		params := GetTestParams{
			XTags: []string{"a", "b"},
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.NotEmpty(t, req.Header.Values("X-Tags"))
	})

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
