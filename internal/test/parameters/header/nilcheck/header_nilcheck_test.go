package headernilcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNilHeaderArrayParam(t *testing.T) {
	t.Run("nil array param is not sent", func(t *testing.T) {
		params := GetHeaderParams{
			XTags: nil,
		}

		req, err := NewGetHeaderRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Header.Values("X-Tags"))
	})

	t.Run("non-nil array param is sent", func(t *testing.T) {
		params := GetHeaderParams{
			XTags: []string{"a", "b"},
		}

		req, err := NewGetHeaderRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.NotEmpty(t, req.Header.Values("X-Tags"))
	})
}
