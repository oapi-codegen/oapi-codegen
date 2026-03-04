package allof

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests allOf composition: PersonProperties → Person → PersonWithID.
func TestAllOfCompositionHierarchy(t *testing.T) {
	t.Run("PersonProperties has all fields optional", func(t *testing.T) {
		pp := PersonProperties{}
		assert.Nil(t, pp.FirstName)
		assert.Nil(t, pp.LastName)
		assert.Nil(t, pp.GovernmentIDNumber)
	})

	t.Run("Person has required FirstName and LastName", func(t *testing.T) {
		p := Person{
			FirstName: "Alex",
			LastName:  "Smith",
		}
		assert.Equal(t, "Alex", p.FirstName)
		assert.Equal(t, "Smith", p.LastName)
		assert.Nil(t, p.GovernmentIDNumber)
	})

	t.Run("PersonWithID adds required ID", func(t *testing.T) {
		p := PersonWithID{
			FirstName: "Alex",
			LastName:  "Smith",
			ID:        12345,
		}
		assert.Equal(t, "Alex", p.FirstName)
		assert.Equal(t, int64(12345), p.ID)
	})
}

func TestAllOfCompositionJSONRoundTrip(t *testing.T) {
	govID := int64(999)
	p := PersonWithID{
		FirstName:          "Alex",
		LastName:           "Smith",
		ID:                 42,
		GovernmentIDNumber: &govID,
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var roundTripped PersonWithID
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, p.FirstName, roundTripped.FirstName)
	assert.Equal(t, p.LastName, roundTripped.LastName)
	assert.Equal(t, p.ID, roundTripped.ID)
	assert.Equal(t, *p.GovernmentIDNumber, *roundTripped.GovernmentIDNumber)
}

// From issue-1219: exhaustive additionalProperties merge matrix.
func TestAdditionalPropertiesMerge(t *testing.T) {
	t.Run("both any: result has any additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]interface{}{}, MergeWithAnyWithAny{}.AdditionalProperties)
	})

	t.Run("any+string: result has typed additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]string{}, MergeWithAnyWithString{}.AdditionalProperties)
	})

	t.Run("string+any: result has typed additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]string{}, MergeWithStringWithAny{}.AdditionalProperties)
	})

	t.Run("any+default: result has any additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]interface{}{}, MergeWithAnyDefault{}.AdditionalProperties)
	})

	t.Run("default+any: result has any additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]interface{}{}, MergeDefaultWithAny{}.AdditionalProperties)
	})

	t.Run("any+false: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithAnyWithout{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("false+any: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithoutWithAny{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("string+default: result has typed additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]string{}, MergeWithStringDefault{}.AdditionalProperties)
	})

	t.Run("default+string: result has typed additionalProperties", func(t *testing.T) {
		assert.IsType(t, map[string]string{}, MergeDefaultWithString{}.AdditionalProperties)
	})

	t.Run("string+false: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithStringWithout{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("false+string: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithoutWithString{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("default+default: no additionalProperties field (compat)", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeDefaultDefault{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("default+false: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeDefaultWithout{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("false+default: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithoutDefault{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})

	t.Run("false+false: no additionalProperties field", func(t *testing.T) {
		_, exist := reflect.TypeOf(MergeWithoutWithout{}).FieldByName("AdditionalProperties")
		assert.False(t, exist)
	})
}

// Verify merged structs contain fields from both source schemas.
func TestAllOfMergedFields(t *testing.T) {
	m := MergeWithAnyWithAny{}
	typ := reflect.TypeOf(m)

	_, hasField1 := typ.FieldByName("Field1")
	assert.True(t, hasField1, "merged struct should have Field1 from first schema")

	_, hasFieldA := typ.FieldByName("FieldA")
	assert.True(t, hasFieldA, "merged struct should have FieldA from second schema")
}
