package issue2329

import (
	"encoding/json"
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

// TestThingMarshalJSON verifies the body-schema custom-marshal path. The
// Thing schema has additionalProperties, so codegen emits a custom
// MarshalJSON that walks named properties one at a time. Map- and
// slice-typed properties marked with `x-go-type-skip-optional-pointer: true`
// must both be guarded by a nil-check there — otherwise an unset map
// property serialises as `"tags":null` while an unset slice property is
// omitted, producing inconsistent output for the same OpenAPI flag.
func TestThingMarshalJSON(t *testing.T) {
	t.Run("zero-value Thing omits both nil map and nil slice properties", func(t *testing.T) {
		b, err := json.Marshal(Thing{})
		require.NoError(t, err)

		var got map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(b, &got))

		assert.NotContains(t, got, "tags", "nil map property must be omitted, not serialised as null")
		assert.NotContains(t, got, "labels", "nil slice property must be omitted, not serialised as null")
	})

	t.Run("populated Thing serialises both map and slice properties", func(t *testing.T) {
		thing := Thing{
			Tags:   map[string]string{"color": "blue"},
			Labels: []string{"a", "b"},
		}
		b, err := json.Marshal(thing)
		require.NoError(t, err)

		assert.Contains(t, string(b), `"tags":{"color":"blue"}`)
		assert.Contains(t, string(b), `"labels":["a","b"]`)
	})
}
