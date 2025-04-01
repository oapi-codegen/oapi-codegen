package xomitzero

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_WithOmitEmpty(t *testing.T) {
	t.Run("with a `string`, without `omitempty`", func(t *testing.T) {
		t.Run("zero value Name does not get omitted", func(t *testing.T) {
			client := Client{
				Name: "",
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "name"))
		})

		t.Run("value Name does not get omitted", func(t *testing.T) {
			client := Client{
				Name: "some value",
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "name"))
		})
	})

	t.Run("with a `*float32` with `omitempty`", func(t *testing.T) {
		var zeroValue float32

		t.Run("nil pointer ID gets omitted", func(t *testing.T) {
			client := Client{
				Id: nil,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.False(t, jsonContainsKey(b, "id"))
		})

		t.Run("pointer to zero value ID does not get omitted", func(t *testing.T) {
			client := Client{
				Id: &zeroValue,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "id"))
		})

		t.Run("pointer to value ID does not get omitted", func(t *testing.T) {
			client := Client{
				Id: &zeroValue,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "id"))
		})
	})
}

func TestClientWithExtension_WithOmitZero(t *testing.T) {
	t.Run("with a `string`, without `omitzero`", func(t *testing.T) {
		t.Run("zero value Name does not get omitted", func(t *testing.T) {
			client := ClientWithExtension{
				Name: "",
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "name"))
		})

		t.Run("value Name does not get omitted", func(t *testing.T) {
			client := ClientWithExtension{
				Name: "some value",
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "name"))
		})
	})

	t.Run("with a `*float32` with `omitzero`", func(t *testing.T) {
		var zeroValue float32

		t.Run("nil pointer ID gets omitted", func(t *testing.T) {
			client := ClientWithExtension{
				Id: nil,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.False(t, jsonContainsKey(b, "id"))
		})

		t.Run("pointer to zero value ID does not get omitted", func(t *testing.T) {
			client := ClientWithExtension{
				Id: &zeroValue,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "id"))
		})

		t.Run("pointer to value ID does not get omitted", func(t *testing.T) {
			client := ClientWithExtension{
				Id: &zeroValue,
			}

			b, err := json.Marshal(client)
			require.NoError(t, err)

			assert.True(t, jsonContainsKey(b, "id"))
		})
	})
}

func TestContainerTypeWithRequired(t *testing.T) {
	t.Run("zero value on HasIsZero does not get omitted", func(t *testing.T) {
		container := ContainerTypeWithRequired{
			HasIsZero: FieldWithCustomIsZeroMethod{},
		}

		b, err := json.Marshal(container)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "has_is_zero"))
	})

	t.Run("value defined as zero value by IsZero on HasIsZero gets omitted", func(t *testing.T) {
		magicIDValue := "this is a zero value, for some weird reason!"

		container := ContainerTypeWithRequired{
			HasIsZero: FieldWithCustomIsZeroMethod{
				Id: &magicIDValue,
			},
		}

		b, err := json.Marshal(container)
		require.NoError(t, err)

		assert.False(t, jsonContainsKey(b, "has_is_zero"))
	})
}

func TestContainerTypeWithOptional(t *testing.T) {
	t.Run("zero value (nil pointer) on HasIsZero gets omitted", func(t *testing.T) {
		container := ContainerTypeWithOptional{
			HasIsZero: nil,
		}

		b, err := json.Marshal(container)
		require.NoError(t, err)

		assert.False(t, jsonContainsKey(b, "has_is_zero"))
	})

	t.Run("value (pointer to zero value of FieldWithCustomIsZeroMethod) on HasIsZero does not get omitted", func(t *testing.T) {
		container := ContainerTypeWithOptional{
			HasIsZero: &FieldWithCustomIsZeroMethod{},
		}

		b, err := json.Marshal(container)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "has_is_zero"))
	})

	t.Run("value defined as zero value by IsZero on HasIsZero gets omitted", func(t *testing.T) {
		magicIDValue := "this is a zero value, for some weird reason!"

		container := ContainerTypeWithOptional{
			HasIsZero: &FieldWithCustomIsZeroMethod{
				Id: &magicIDValue,
			},
		}

		b, err := json.Marshal(container)
		require.NoError(t, err)

		assert.False(t, jsonContainsKey(b, "has_is_zero"))
	})
}

func TestFieldWithOmitZeroOnRequiredField(t *testing.T) {
	t.Run("zero value (empty string) on Id gets omitted", func(t *testing.T) {
		field := FieldWithOmitZeroOnRequiredField{
			Id: "",
		}

		b, err := json.Marshal(field)
		require.NoError(t, err)

		assert.False(t, jsonContainsKey(b, "id"))
	})

	t.Run("value for Id not get omitted", func(t *testing.T) {
		field := FieldWithOmitZeroOnRequiredField{
			Id: "value",
		}

		b, err := json.Marshal(field)
		require.NoError(t, err)

		assert.True(t, jsonContainsKey(b, "id"))
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
