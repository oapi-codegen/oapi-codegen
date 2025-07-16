package preferskipoptionalpointer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Run("zero value (empty string) on Name is not omitted", func(t *testing.T) {
		client := Client{
			Name: "",
		}

		b, err := json.Marshal(client)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "name"))
	})

	t.Run("value on Name is not omitted", func(t *testing.T) {
		client := Client{
			Name: "some value",
		}

		b, err := json.Marshal(client)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "name"))
	})

	t.Run("zero value (0.0) on Id is omitted (as `omitempty` flags it as empty)", func(t *testing.T) {
		client := Client{
			Id: 0.0,
		}

		b, err := json.Marshal(client)
		require.NoError(t, err)

		assert.False(t, jsonContainsKey(b, "id"))
	})

	t.Run("value on Id is not omitted", func(t *testing.T) {
		client := Client{
			Id: 3.142,
		}

		b, err := json.Marshal(client)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "id"))
	})
}

func TestNestedType(t *testing.T) {
	t.Run("zero value (empty struct) on Client is not omitted", func(t *testing.T) {
		nestedType := NestedType{
			Client: Client{},
		}

		b, err := json.Marshal(nestedType)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "client"))
	})

	t.Run("value on Client is not omitted", func(t *testing.T) {
		nestedType := NestedType{
			Client: Client{
				Name: "foo",
			},
		}

		b, err := json.Marshal(nestedType)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "client"))
	})
}

// jsonContainsKey checks if the given JSON object contains the specified key at the top level.
func jsonContainsKey(b []byte, key string) bool {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
