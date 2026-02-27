package objects

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseObjectRoundTrip(t *testing.T) {
	obj := BaseObject{Role: "admin", FirstName: "Alex"}
	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var roundTripped BaseObject
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, obj, roundTripped)
}

func TestTypedAdditionalProperties(t *testing.T) {
	input := `{"name":"test","id":1,"extra1":10,"extra2":20}`

	var obj TypedAdditionalProperties
	err := json.Unmarshal([]byte(input), &obj)
	require.NoError(t, err)

	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, 1, obj.Id)

	val, found := obj.Get("extra1")
	assert.True(t, found)
	assert.Equal(t, 10, val)

	val, found = obj.Get("extra2")
	assert.True(t, found)
	assert.Equal(t, 20, val)

	// Round-trip
	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var roundTripped TypedAdditionalProperties
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, obj.Name, roundTripped.Name)
	assert.Equal(t, obj.Id, roundTripped.Id)
	v, _ := roundTripped.Get("extra1")
	assert.Equal(t, 10, v)
}

func TestTypedAdditionalPropertiesSetGet(t *testing.T) {
	obj := TypedAdditionalProperties{Name: "test", Id: 1}
	obj.Set("count", 42)

	val, found := obj.Get("count")
	assert.True(t, found)
	assert.Equal(t, 42, val)

	_, found = obj.Get("nonexistent")
	assert.False(t, found)
}

func TestNoAdditionalProperties(t *testing.T) {
	obj := NoAdditionalProperties{Name: "test", Id: 1}
	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var roundTripped NoAdditionalProperties
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, obj, roundTripped)
}

func TestAnyAdditionalProperties(t *testing.T) {
	input := `{"name":"test","extra":"value","count":42}`

	var obj AnyAdditionalProperties
	err := json.Unmarshal([]byte(input), &obj)
	require.NoError(t, err)

	assert.Equal(t, "test", obj.Name)

	val, found := obj.Get("extra")
	assert.True(t, found)
	assert.Equal(t, "value", val)

	val, found = obj.Get("count")
	assert.True(t, found)
	assert.Equal(t, float64(42), val) // JSON numbers unmarshal as float64

	// Set and round-trip
	obj.Set("new_key", "new_value")
	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var roundTripped AnyAdditionalProperties
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	v, _ := roundTripped.Get("new_key")
	assert.Equal(t, "new_value", v)
}

func TestNestedAdditionalProperties(t *testing.T) {
	input := `{"name":"outer","inner":{"name":"inner_val","extra_inner":"hi"},"extra_outer":"bye"}`

	var obj NestedAdditionalProperties
	err := json.Unmarshal([]byte(input), &obj)
	require.NoError(t, err)

	assert.Equal(t, "outer", obj.Name)
	assert.Equal(t, "inner_val", obj.Inner.Name)

	innerVal, found := obj.Inner.Get("extra_inner")
	assert.True(t, found)
	assert.Equal(t, "hi", innerVal)

	outerVal, found := obj.Get("extra_outer")
	assert.True(t, found)
	assert.Equal(t, "bye", outerVal)
}

func TestRefAdditionalProperties(t *testing.T) {
	refMap := RefAdditionalProperties{
		"user1": {Role: "admin", FirstName: "Alex"},
		"user2": {Role: "viewer", FirstName: "Sam"},
	}

	data, err := json.Marshal(refMap)
	require.NoError(t, err)

	var roundTripped RefAdditionalProperties
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.Equal(t, "admin", roundTripped["user1"].Role)
	assert.Equal(t, "Sam", roundTripped["user2"].FirstName)
}

func TestArrayOfMaps(t *testing.T) {
	arr := ArrayOfMaps{
		{"key1": {Role: "admin", FirstName: "Alex"}},
		{"key2": {Role: "viewer", FirstName: "Sam"}},
	}

	data, err := json.Marshal(arr)
	require.NoError(t, err)

	var roundTripped ArrayOfMaps
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.Len(t, roundTripped, 2)
	assert.Equal(t, "admin", roundTripped[0]["key1"].Role)
}

func TestReadOnlyWriteOnly(t *testing.T) {
	readOnly := "read_value"
	writeOnly := 42
	obj := ReadOnlyWriteOnly{
		NormalProp:    "normal",
		ReadOnlyProp:  &readOnly,
		WriteOnlyProp: &writeOnly,
	}

	data, err := json.Marshal(obj)
	require.NoError(t, err)

	var roundTripped ReadOnlyWriteOnly
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, "normal", roundTripped.NormalProp)
	assert.Equal(t, "read_value", *roundTripped.ReadOnlyProp)
	assert.Equal(t, 42, *roundTripped.WriteOnlyProp)
}
