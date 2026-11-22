package extensionsskipoptionalpointer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issue #2031: an optional, non-nullable array must be OMITTED when nil, never
// marshaled as `null` (which is invalid per the schema). additionalProperties:true
// on ArrayContainer forces the custom MarshalJSON path that exposed the bug.
func TestMarshal(t *testing.T) {
	value := ArrayContainer{}
	content, err := json.Marshal(value)
	require.NoError(t, err)
	assert.Equal(t, "{}", string(content))
}

// issue #1561: with prefer-skip-optional-pointer-on-container-types enabled, container
// fields (slices/maps/[]byte) are emitted as non-pointer values; a property-level
// x-go-type-skip-optional-pointer:false re-introduces the pointer (BytesWithOverride).
func TestResponseBody_DoesNotHaveOptionalPointerToContainerTypes(t *testing.T) {
	pong0 := Pong{
		Ping: "0th",
	}

	pong1 := Pong{
		Ping: "1th",
	}

	slice := []Pong{
		pong0,
		pong1,
	}

	m := map[string]Pong{
		"0": pong0,
		"1": pong1,
	}

	byteData := []byte("some bytes")

	body := ResponseBody{
		RequiredSlice:             slice,
		ASlice:                    slice,
		AMap:                      m,
		UnknownObject:             map[string]any{},
		AdditionalProps:           m,
		ASliceWithAdditionalProps: []map[string]Pong{m},
		Bytes:                     byteData,
		BytesWithOverride:         &byteData,
	}

	assert.NotNil(t, body.RequiredSlice)
	assert.NotZero(t, body.RequiredSlice)

	assert.NotNil(t, body.ASlice)
	assert.NotZero(t, body.ASlice)

	assert.NotNil(t, body.AMap)
	assert.NotZero(t, body.AMap)

	assert.NotNil(t, body.UnknownObject)
	assert.Empty(t, body.UnknownObject)

	assert.NotNil(t, body.AdditionalProps)
	assert.NotZero(t, body.AdditionalProps)

	assert.NotNil(t, body.ASliceWithAdditionalProps)
	assert.NotZero(t, body.ASliceWithAdditionalProps)

	assert.NotNil(t, body.Bytes)
	assert.NotZero(t, body.Bytes)

	assert.NotNil(t, body.BytesWithOverride)
	assert.NotZero(t, body.BytesWithOverride)
}

// issue #2329: map- and slice-typed optional query parameters marked with
// x-go-type-skip-optional-pointer:true must produce client request code that compiles
// (the fields are `map[string]string` / `[]string`, not pointers), and nil params
// must not be sent.
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

// issue #2329: the Thing schema has additionalProperties, so codegen emits a custom
// MarshalJSON that walks named properties one at a time. Map- and slice-typed
// properties marked skip-optional-pointer must both be nil-guarded there — otherwise
// an unset map serialises as `"tags":null` while an unset slice is omitted, producing
// inconsistent output for the same flag.
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
