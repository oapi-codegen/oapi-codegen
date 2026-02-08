package helpers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalForm_SimpleStruct(t *testing.T) {
	type simple struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Admin bool   `json:"admin"`
	}

	vals, err := marshalForm(simple{Name: "alice", Age: 30, Admin: true})
	require.NoError(t, err)
	assert.Equal(t, "alice", vals.Get("name"))
	assert.Equal(t, "30", vals.Get("age"))
	assert.Equal(t, "true", vals.Get("admin"))
}

func TestMarshalForm_PointerFields(t *testing.T) {
	type withPtrs struct {
		Required string  `json:"required"`
		Optional *string `json:"optional"`
	}

	t.Run("nil pointer omitted", func(t *testing.T) {
		vals, err := marshalForm(withPtrs{Required: "yes", Optional: nil})
		require.NoError(t, err)
		assert.Equal(t, "yes", vals.Get("required"))
		_, exists := vals["optional"]
		assert.False(t, exists)
	})

	t.Run("non-nil pointer included", func(t *testing.T) {
		opt := "hello"
		vals, err := marshalForm(withPtrs{Required: "yes", Optional: &opt})
		require.NoError(t, err)
		assert.Equal(t, "yes", vals.Get("required"))
		assert.Equal(t, "hello", vals.Get("optional"))
	})
}

func TestMarshalForm_OmitEmpty(t *testing.T) {
	type withOmitEmpty struct {
		Name  string `json:"name"`
		Value string `json:"value,omitempty"`
	}

	t.Run("omitempty with zero value", func(t *testing.T) {
		vals, err := marshalForm(withOmitEmpty{Name: "test"})
		require.NoError(t, err)
		assert.Equal(t, "test", vals.Get("name"))
		_, exists := vals["value"]
		assert.False(t, exists)
	})

	t.Run("omitempty with non-zero value", func(t *testing.T) {
		vals, err := marshalForm(withOmitEmpty{Name: "test", Value: "present"})
		require.NoError(t, err)
		assert.Equal(t, "present", vals.Get("value"))
	})
}

func TestMarshalForm_SliceField(t *testing.T) {
	type withSlice struct {
		Tags []string `json:"tags"`
	}

	vals, err := marshalForm(withSlice{Tags: []string{"a", "b", "c"}})
	require.NoError(t, err)
	assert.Equal(t, "a", vals.Get("tags[0]"))
	assert.Equal(t, "b", vals.Get("tags[1]"))
	assert.Equal(t, "c", vals.Get("tags[2]"))
}

func TestMarshalForm_NestedStruct(t *testing.T) {
	type inner struct {
		City string `json:"city"`
	}
	type outer struct {
		Address inner `json:"address"`
	}

	vals, err := marshalForm(outer{Address: inner{City: "NYC"}})
	require.NoError(t, err)
	assert.Equal(t, "NYC", vals.Get("address[city]"))
}

func TestMarshalForm_AdditionalProperties(t *testing.T) {
	type withAP struct {
		Name                 string            `json:"name"`
		AdditionalProperties map[string]string `json:"-"`
	}

	vals, err := marshalForm(withAP{
		Name:                 "test",
		AdditionalProperties: map[string]string{"extra1": "val1", "extra2": "val2"},
	})
	require.NoError(t, err)
	assert.Equal(t, "test", vals.Get("name"))
	// AdditionalProperties from the top-level struct are at the top level
	// But since marshalForm is called on the struct, the top-level fields
	// go through the outer loop, not the recursive one. The inner struct
	// handling sees AdditionalProperties with json:"-" and expands the map.
	// However, the top-level loop skips json:"-" fields.
	// Let me verify: the top-level loop in marshalForm skips tag == "-".
	// So AdditionalProperties at top level would be skipped.
	// This matches the real codegen where AdditionalProperties is in a nested struct.

	// Test nested AdditionalProperties
	type innerWithAP struct {
		Base                 string            `json:"base"`
		AdditionalProperties map[string]string `json:"-"`
	}
	type outerWithAP struct {
		Data innerWithAP `json:"data"`
	}

	vals2, err := marshalForm(outerWithAP{
		Data: innerWithAP{
			Base:                 "hello",
			AdditionalProperties: map[string]string{"x": "1", "y": "2"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello", vals2.Get("data[base]"))
	assert.Equal(t, "1", vals2.Get("data[x]"))
	assert.Equal(t, "2", vals2.Get("data[y]"))
}

func TestMarshalForm_NonStructReturnsError(t *testing.T) {
	_, err := marshalForm("not a struct")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "struct")
}

func TestMarshalForm_SkipDashTag(t *testing.T) {
	type withDash struct {
		Visible  string `json:"visible"`
		Excluded string `json:"-"`
	}

	vals, err := marshalForm(withDash{Visible: "yes", Excluded: "no"})
	require.NoError(t, err)
	assert.Equal(t, url.Values{"visible": {"yes"}}, vals)
}
