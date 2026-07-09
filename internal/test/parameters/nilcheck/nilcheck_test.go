package parametersnilcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// From issue-2238: a nil optional array header param must not be sent; a non-nil one must be.
func TestNilHeaderArrayParam(t *testing.T) {
	t.Run("nil header array param is not sent", func(t *testing.T) {
		params := GetHeaderParams{
			XTags: nil,
		}

		req, err := NewGetHeaderRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Header.Values("X-Tags"))
	})

	t.Run("non-nil header array param is sent", func(t *testing.T) {
		params := GetHeaderParams{
			XTags: []string{"a", "b"},
		}

		req, err := NewGetHeaderRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.NotEmpty(t, req.Header.Values("X-Tags"))
	})
}

// From issue-2238: a nil optional array cookie param must not be sent; a non-nil one must be.
func TestNilCookieArrayParam(t *testing.T) {
	t.Run("nil cookie array param is not sent", func(t *testing.T) {
		params := GetCookieParams{
			Tags: nil,
		}

		req, err := NewGetCookieRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.Cookies())
	})

	t.Run("non-nil cookie array param is sent", func(t *testing.T) {
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

// From issue-2031: under prefer-skip-optional-pointer, an optional array query
// param is a bare slice. A nil slice must be omitted entirely, but an
// explicitly initialised empty slice must still be serialized — the client
// must preserve the nil-vs-empty distinction.
func TestNilQueryArrayParam(t *testing.T) {
	t.Run("does not add the user_ids[] parameter if zero value", func(t *testing.T) {
		params := GetQueryParams{}

		req, err := NewGetQueryRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/query", req.URL.String())
	})

	t.Run("does not add the user_ids[] parameter if nil", func(t *testing.T) {
		params := GetQueryParams{
			UserIds: nil,
		}

		req, err := NewGetQueryRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/query", req.URL.String())
	})

	t.Run("adds the user_ids[] parameter if an explicitly initialised empty array", func(t *testing.T) {
		params := GetQueryParams{
			UserIds: []int{},
		}

		req, err := NewGetQueryRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/query?user_ids%5B%5D=", req.URL.String())
	})

	t.Run("adds the user_ids[] parameter if array contains a value", func(t *testing.T) {
		params := GetQueryParams{
			UserIds: []int{1},
		}

		req, err := NewGetQueryRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/query?user_ids%5B%5D=1", req.URL.String())
	})

	t.Run("handles multiple user_ids[] parameters", func(t *testing.T) {
		params := GetQueryParams{
			UserIds: []int{1, 100},
		}

		req, err := NewGetQueryRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/query?user_ids%5B%5D=1&user_ids%5B%5D=100", req.URL.String())
	})
}
