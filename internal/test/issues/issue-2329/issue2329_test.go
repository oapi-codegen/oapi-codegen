package issue2329

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewListThingsRequest verifies that map- and slice-typed optional query
// parameters marked with `x-go-type-skip-optional-pointer: true` produce
// client request code that compiles. Before the fix, the client template
// emitted `*params.Tags` / `*params.Labels`, which does not compile because
// the fields are declared as `map[string]string` and `[]string`.
func TestNewListThingsRequest(t *testing.T) {
	t.Run("nil map and slice query params are not sent", func(t *testing.T) {
		params := ListThingsParams{}

		req, err := NewListThingsRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Empty(t, req.URL.RawQuery)
	})

	t.Run("non-nil map query param (deepObject) is serialized", func(t *testing.T) {
		params := ListThingsParams{
			Tags: map[string]string{"color": "blue"},
		}

		req, err := NewListThingsRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Contains(t, req.URL.RawQuery, "tags[color]=blue")
	})

	t.Run("non-nil slice query param (form, explode) is serialized", func(t *testing.T) {
		params := ListThingsParams{
			Labels: []string{"a", "b"},
		}

		req, err := NewListThingsRequest("https://localhost", &params)
		require.NoError(t, err)

		assert.Contains(t, req.URL.RawQuery, "labels=a")
		assert.Contains(t, req.URL.RawQuery, "labels=b")
	})
}
