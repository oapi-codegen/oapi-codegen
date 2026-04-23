package issue2031

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGetTestRequest(t *testing.T) {
	t.Run("does not add the user_ids[] parameter if zero value", func(t *testing.T) {
		params := GetTestParams{}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/test", req.URL.String())
	})

	t.Run("does not add the user_ids[] parameter if nil", func(t *testing.T) {
		params := GetTestParams{
			UserIds: nil,
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/test", req.URL.String())
	})

	t.Run("adds the user_ids[] parameter if an explicitly initialised empty array", func(t *testing.T) {
		params := GetTestParams{
			UserIds: []int{},
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/test?user_ids%5B%5D=", req.URL.String())
	})

	t.Run("adds the user_ids[] parameter if array contains a value", func(t *testing.T) {
		params := GetTestParams{
			UserIds: []int{1},
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/test?user_ids%5B%5D=1", req.URL.String())
	})

	t.Run("handles multiple user_ids[] parameters", func(t *testing.T) {
		params := GetTestParams{
			UserIds: []int{1, 100},
		}

		req, err := NewGetTestRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Equal(t, "https://localhost/test?user_ids%5B%5D=1&user_ids%5B%5D=100", req.URL.String())
	})
}
