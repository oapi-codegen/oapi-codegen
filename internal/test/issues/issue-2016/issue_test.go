package issue2016

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullableArrayItemsMarshaling tests that nullable array items can handle nil values
func TestNullableArrayItemsMarshaling(t *testing.T) {
	t.Run("Sample with nullable number items", func(t *testing.T) {
		// Create a sample with nil values in the array
		val1 := float32(1.5)
		val3 := float32(3.5)

		sample := Sample{
			Metrics: []*float32{&val1, nil, &val3},
		}

		// Marshal to JSON
		data, err := json.Marshal(sample)
		require.NoError(t, err)

		// Verify JSON contains null
		assert.JSONEq(t, `{"metrics":[1.5,null,3.5]}`, string(data))

		// Unmarshal back
		var decoded Sample
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		// Verify values
		require.Len(t, decoded.Metrics, 3)
		assert.Equal(t, float32(1.5), *decoded.Metrics[0])
		assert.Nil(t, decoded.Metrics[1])
		assert.Equal(t, float32(3.5), *decoded.Metrics[2])
	})

	t.Run("StringArrayNullable with nil values", func(t *testing.T) {
		str1 := "hello"
		str3 := "world"

		obj := StringArrayNullable{
			Values: &[]*string{&str1, nil, &str3},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"values":["hello",null,"world"]}`, string(data))

		var decoded StringArrayNullable
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.NotNil(t, decoded.Values)
		require.Len(t, *decoded.Values, 3)
		assert.Equal(t, "hello", *(*decoded.Values)[0])
		assert.Nil(t, (*decoded.Values)[1])
		assert.Equal(t, "world", *(*decoded.Values)[2])
	})

	t.Run("BooleanArrayNullable with nil values", func(t *testing.T) {
		valTrue := true
		valFalse := false

		obj := BooleanArrayNullable{
			Flags: &[]*bool{&valTrue, nil, &valFalse},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"flags":[true,null,false]}`, string(data))
	})

	t.Run("IntegerArrayNullable with nil values", func(t *testing.T) {
		val1 := int64(100)
		val3 := int64(300)

		obj := IntegerArrayNullable{
			Counts: &[]*int64{&val1, nil, &val3},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"counts":[100,null,300]}`, string(data))
	})

	t.Run("NestedArrayNullable with nil values", func(t *testing.T) {
		val1 := float32(1.0)
		val3 := float32(3.0)

		obj := NestedArrayNullable{
			Matrix: &[][]*float32{
				{&val1, nil, &val3},
				{nil, nil, nil},
			},
		}

		data, err := json.Marshal(obj)
		require.NoError(t, err)
		assert.JSONEq(t, `{"matrix":[[1.0,null,3.0],[null,null,null]]}`, string(data))
	})

	t.Run("NullableArrayWithNullableItems both null", func(t *testing.T) {
		str1 := "test"

		// Test with nil array
		obj1 := NullableArrayWithNullableItems{
			Data: nil,
		}
		data1, err := json.Marshal(obj1)
		require.NoError(t, err)
		assert.JSONEq(t, `{"data":null}`, string(data1))

		// Test with array containing nil items
		obj2 := NullableArrayWithNullableItems{
			Data: &[]*string{&str1, nil},
		}
		data2, err := json.Marshal(obj2)
		require.NoError(t, err)
		assert.JSONEq(t, `{"data":["test",null]}`, string(data2))
	})
}

// TestTypeCorrectness verifies that the types are generated correctly
func TestTypeCorrectness(t *testing.T) {
	t.Run("Sample.Metrics should be []*float32", func(t *testing.T) {
		var sample Sample
		// This will fail to compile if the type is wrong
		var _ []*float32 = sample.Metrics
	})

	t.Run("StringArrayNullable.Values should be *[]*string", func(t *testing.T) {
		var obj StringArrayNullable
		// This will fail to compile if the type is wrong
		var _ *[]*string = obj.Values
	})

	t.Run("NestedArrayNullable.Matrix should be *[][]*float32", func(t *testing.T) {
		var obj NestedArrayNullable
		// This will fail to compile if the type is wrong
		var _ *[][]*float32 = obj.Matrix
	})
}
