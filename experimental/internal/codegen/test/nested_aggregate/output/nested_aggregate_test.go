package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

// TestArrayOfAnyOf tests marshaling/unmarshaling of arrays with anyOf items
func TestArrayOfAnyOf(t *testing.T) {
	t.Run("unmarshal string item", func(t *testing.T) {
		input := `["hello", "world"]`
		var arr ArrayOfAnyOf
		err := json.Unmarshal([]byte(input), &arr)
		require.NoError(t, err)
		require.Len(t, arr, 2)

		// String items should populate the string field
		assert.NotNil(t, arr[0].String0)
		assert.Equal(t, "hello", *arr[0].String0)
		assert.NotNil(t, arr[1].String0)
		assert.Equal(t, "world", *arr[1].String0)
	})

	t.Run("unmarshal object item", func(t *testing.T) {
		input := `[{"id": 42}]`
		var arr ArrayOfAnyOf
		err := json.Unmarshal([]byte(input), &arr)
		require.NoError(t, err)
		require.Len(t, arr, 1)

		// Object item should populate the object field
		assert.NotNil(t, arr[0].ArrayOfAnyOfAnyOf1)
		assert.NotNil(t, arr[0].ArrayOfAnyOfAnyOf1.ID)
		assert.Equal(t, 42, *arr[0].ArrayOfAnyOfAnyOf1.ID)
	})

	t.Run("unmarshal mixed items", func(t *testing.T) {
		input := `["hello", {"id": 1}, "world", {"id": 2}]`
		var arr ArrayOfAnyOf
		err := json.Unmarshal([]byte(input), &arr)
		require.NoError(t, err)
		require.Len(t, arr, 4)

		assert.NotNil(t, arr[0].String0)
		assert.Equal(t, "hello", *arr[0].String0)

		assert.NotNil(t, arr[1].ArrayOfAnyOfAnyOf1)
		assert.Equal(t, 1, *arr[1].ArrayOfAnyOfAnyOf1.ID)

		assert.NotNil(t, arr[2].String0)
		assert.Equal(t, "world", *arr[2].String0)

		assert.NotNil(t, arr[3].ArrayOfAnyOfAnyOf1)
		assert.Equal(t, 2, *arr[3].ArrayOfAnyOfAnyOf1.ID)
	})

	t.Run("marshal string item", func(t *testing.T) {
		arr := ArrayOfAnyOf{
			{String0: ptr("hello")},
		}
		data, err := json.Marshal(arr)
		require.NoError(t, err)
		assert.JSONEq(t, `["hello"]`, string(data))
	})

	t.Run("marshal object item", func(t *testing.T) {
		arr := ArrayOfAnyOf{
			{ArrayOfAnyOfAnyOf1: &ArrayOfAnyOfAnyOf1{ID: ptr(42)}},
		}
		data, err := json.Marshal(arr)
		require.NoError(t, err)
		assert.JSONEq(t, `[{"id": 42}]`, string(data))
	})

	t.Run("round trip mixed", func(t *testing.T) {
		original := ArrayOfAnyOf{
			{String0: ptr("test")},
			{ArrayOfAnyOfAnyOf1: &ArrayOfAnyOfAnyOf1{ID: ptr(99)}},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded ArrayOfAnyOf
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.Len(t, decoded, 2)
		assert.Equal(t, "test", *decoded[0].String0)
		assert.Equal(t, 99, *decoded[1].ArrayOfAnyOfAnyOf1.ID)
	})
}

// TestObjectWithAnyOfProperty tests marshaling/unmarshaling of objects with anyOf properties
func TestObjectWithAnyOfProperty(t *testing.T) {
	t.Run("unmarshal string value", func(t *testing.T) {
		input := `{"value": "hello"}`
		var obj ObjectWithAnyOfProperty
		err := json.Unmarshal([]byte(input), &obj)
		require.NoError(t, err)

		require.NotNil(t, obj.Value)
		assert.NotNil(t, obj.Value.String0)
		assert.Equal(t, "hello", *obj.Value.String0)
	})

	t.Run("unmarshal integer value", func(t *testing.T) {
		input := `{"value": 42}`
		var obj ObjectWithAnyOfProperty
		err := json.Unmarshal([]byte(input), &obj)
		require.NoError(t, err)

		require.NotNil(t, obj.Value)
		assert.NotNil(t, obj.Value.Int1)
		assert.Equal(t, 42, *obj.Value.Int1)
	})

	t.Run("marshal string value", func(t *testing.T) {
		obj := ObjectWithAnyOfProperty{
			Value: &ObjectWithAnyOfPropertyValue{
				String0: ptr("hello"),
			},
		}
		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"value": "hello"}`, string(data))
	})

	t.Run("marshal integer value", func(t *testing.T) {
		obj := ObjectWithAnyOfProperty{
			Value: &ObjectWithAnyOfPropertyValue{
				Int1: ptr(42),
			},
		}
		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"value": 42}`, string(data))
	})

	t.Run("round trip string", func(t *testing.T) {
		original := ObjectWithAnyOfProperty{
			Value: &ObjectWithAnyOfPropertyValue{String0: ptr("test")},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded ObjectWithAnyOfProperty
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.NotNil(t, decoded.Value)
		assert.Equal(t, "test", *decoded.Value.String0)
	})
}

// TestObjectWithOneOfProperty tests marshaling/unmarshaling of objects with oneOf properties
func TestObjectWithOneOfProperty(t *testing.T) {
	t.Run("unmarshal ambiguous input errors", func(t *testing.T) {
		// Both variants have optional "kind" field, so this JSON matches both
		// oneOf requires exactly one match, so this should error
		input := `{"variant": {"kind": "person", "name": "Alice"}}`
		var obj ObjectWithOneOfProperty
		err := json.Unmarshal([]byte(input), &obj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected exactly one type to match, got 2")
	})

	t.Run("unmarshal unambiguous variant 0", func(t *testing.T) {
		// Only variant 0 has "name" as a field that can be set
		// But since all fields are optional, both variants still match
		// This demonstrates why discriminators are important for oneOf
		input := `{"variant": {"name": "Alice"}}`
		var obj ObjectWithOneOfProperty
		err := json.Unmarshal([]byte(input), &obj)
		// Still ambiguous because both variants can unmarshal (missing fields are just nil)
		require.Error(t, err)
	})

	t.Run("unmarshal unambiguous variant 1", func(t *testing.T) {
		// Only variant 1 has "count" field
		// But since all fields are optional, both variants still match
		input := `{"variant": {"count": 10}}`
		var obj ObjectWithOneOfProperty
		err := json.Unmarshal([]byte(input), &obj)
		// Still ambiguous because both variants can unmarshal (missing fields are just nil)
		require.Error(t, err)
	})

	t.Run("marshal variant 0", func(t *testing.T) {
		obj := ObjectWithOneOfProperty{
			Variant: &ObjectWithOneOfPropertyVariant{
				ObjectWithOneOfPropertyVariantOneOf0: &ObjectWithOneOfPropertyVariantOneOf0{
					Kind: ptr("person"),
					Name: ptr("Alice"),
				},
			},
		}
		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"variant": {"kind": "person", "name": "Alice"}}`, string(data))
	})

	t.Run("marshal variant 1", func(t *testing.T) {
		obj := ObjectWithOneOfProperty{
			Variant: &ObjectWithOneOfPropertyVariant{
				ObjectWithOneOfPropertyVariantOneOf1: &ObjectWithOneOfPropertyVariantOneOf1{
					Kind:  ptr("counter"),
					Count: ptr(10),
				},
			},
		}
		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"variant": {"kind": "counter", "count": 10}}`, string(data))
	})

	t.Run("marshal fails with zero variants set", func(t *testing.T) {
		obj := ObjectWithOneOfProperty{
			Variant: &ObjectWithOneOfPropertyVariant{},
		}
		_, err := json.Marshal(obj)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one member must be set")
	})

	t.Run("marshal fails with two variants set", func(t *testing.T) {
		obj := ObjectWithOneOfProperty{
			Variant: &ObjectWithOneOfPropertyVariant{
				ObjectWithOneOfPropertyVariantOneOf0: &ObjectWithOneOfPropertyVariantOneOf0{
					Kind: ptr("person"),
					Name: ptr("Alice"),
				},
				ObjectWithOneOfPropertyVariantOneOf1: &ObjectWithOneOfPropertyVariantOneOf1{
					Kind:  ptr("counter"),
					Count: ptr(10),
				},
			},
		}
		_, err := json.Marshal(obj)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one member must be set")
	})
}

// TestAllOfWithOneOf tests marshaling/unmarshaling of allOf containing oneOf
func TestAllOfWithOneOf(t *testing.T) {
	t.Run("unmarshal with optionA - ambiguous oneOf errors", func(t *testing.T) {
		// The nested oneOf has same ambiguity issue - both variants match
		input := `{"base": "test", "optionA": true}`
		var obj AllOfWithOneOf
		err := json.Unmarshal([]byte(input), &obj)
		// The nested AllOfWithOneOfAllOf1 (oneOf) will error due to ambiguity
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected exactly one type to match")
	})

	t.Run("unmarshal with optionB - ambiguous oneOf errors", func(t *testing.T) {
		input := `{"base": "test", "optionB": 42}`
		var obj AllOfWithOneOf
		err := json.Unmarshal([]byte(input), &obj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected exactly one type to match")
	})

	t.Run("marshal with optionA", func(t *testing.T) {
		obj := AllOfWithOneOf{
			Base: ptr("test"),
			AllOfWithOneOfAllOf1: &AllOfWithOneOfAllOf1{
				AllOfWithOneOfAllOf1OneOf0: &AllOfWithOneOfAllOf1OneOf0{
					OptionA: ptr(true),
				},
			},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)

		// Should contain both base and optionA merged
		var m map[string]any
		err = json.Unmarshal(data, &m)
		require.NoError(t, err)

		assert.Equal(t, "test", m["base"])
		assert.Equal(t, true, m["optionA"])
	})

	t.Run("marshal with optionB", func(t *testing.T) {
		obj := AllOfWithOneOf{
			Base: ptr("test"),
			AllOfWithOneOfAllOf1: &AllOfWithOneOfAllOf1{
				AllOfWithOneOfAllOf1OneOf1: &AllOfWithOneOfAllOf1OneOf1{
					OptionB: ptr(42),
				},
			},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)

		var m map[string]any
		err = json.Unmarshal(data, &m)
		require.NoError(t, err)

		assert.Equal(t, "test", m["base"])
		assert.Equal(t, float64(42), m["optionB"]) // JSON numbers are float64
	})

	t.Run("marshal with nil union", func(t *testing.T) {
		obj := AllOfWithOneOf{
			Base: ptr("only-base"),
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)

		var m map[string]any
		err = json.Unmarshal(data, &m)
		require.NoError(t, err)

		assert.Equal(t, "only-base", m["base"])
		assert.NotContains(t, m, "optionA")
		assert.NotContains(t, m, "optionB")
	})
}
