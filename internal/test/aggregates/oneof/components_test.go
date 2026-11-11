package aggregatesoneof

// Comprehensive oneOf + anyOf marshaling tests folded from components/components.yaml.

import (
	"encoding/json"
	"testing"

	"github.com/oapi-codegen/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingCodec counts the calls that reach it.
type countingCodec struct {
	runtime.StdJSONCodec
	marshals, unmarshals int
}

func (c *countingCodec) Marshal(v any) ([]byte, error) {
	c.marshals++
	return c.StdJSONCodec.Marshal(v)
}

func (c *countingCodec) Unmarshal(data []byte, v any) error {
	c.unmarshals++
	return c.StdJSONCodec.Unmarshal(data, v)
}

// TestUnionJSONCodec checks that a union's From*, As*, Merge*, MarshalJSON and
// UnmarshalJSON use the codec set with runtime.SetJSONCodec, with and without
// additionalProperties. MarshalJSON and UnmarshalJSON run inside the codec's
// own Marshal and Unmarshal of the value, so this is also the re-entrant case.
func TestUnionJSONCodec(t *testing.T) {
	var codec countingCodec
	runtime.SetJSONCodec(&codec)
	t.Cleanup(func() { runtime.SetJSONCodec(nil) })

	// calls runs f and returns how many Marshal and Unmarshal calls it made.
	calls := func(f func()) (marshals, unmarshals int) {
		m, u := codec.marshals, codec.unmarshals
		f()
		return codec.marshals - m, codec.unmarshals - u
	}

	t.Run("union", func(t *testing.T) {
		var dst OneOfObject1
		m, u := calls(func() {
			require.NoError(t, dst.FromOneOfVariant1(OneOfVariant1{Name: "123"}))
		})
		assert.Equal(t, 1, m, "From")
		assert.Zero(t, u, "From")

		m, u = calls(func() {
			v1, err := dst.AsOneOfVariant1()
			require.NoError(t, err)
			assert.Equal(t, "123", v1.Name)
		})
		assert.Zero(t, m, "As")
		assert.Equal(t, 1, u, "As")

		m, _ = calls(func() {
			require.NoError(t, dst.MergeOneOfVariant1(OneOfVariant1{Name: "456"}))
		})
		assert.Positive(t, m, "Merge")
	})

	t.Run("union and additionalProperties", func(t *testing.T) {
		x := OneOfObject13{
			AdditionalProperties: map[string]any{"x": "y"},
		}
		require.NoError(t, x.MergeOneOfVariant1(OneOfVariant1{Name: "test-name"}))

		var b []byte
		m, _ := calls(func() {
			var err error
			b, err = runtime.JSONMarshal(x)
			require.NoError(t, err)
		})
		assert.JSONEq(t, `{"x":"y", "name":"test-name", "type":"v1"}`, string(b))
		// One call from here, the rest from MarshalJSON.
		assert.Greater(t, m, 1, "MarshalJSON")

		var y OneOfObject13
		_, u := calls(func() {
			require.NoError(t, runtime.JSONUnmarshal(b, &y))
		})
		yAdditional, ok := y.Get("x")
		assert.True(t, ok)
		assert.Equal(t, "y", yAdditional)
		// One call from here, the rest from UnmarshalJSON.
		assert.Greater(t, u, 1, "UnmarshalJSON")
	})
}

func assertJsonEqual(t *testing.T, j1 []byte, j2 []byte) {
	t.Helper()
	assert.JSONEq(t, string(j1), string(j2))
}

func TestOneOf(t *testing.T) {
	const variant1 = `{"name": "123"}`
	const variant2 = `[1, 2, 3]`
	const variant3 = `true`
	var dst OneOfObject1

	err := json.Unmarshal([]byte(variant1), &dst)
	assert.NoError(t, err)
	v1, err := dst.AsOneOfVariant1()
	assert.NoError(t, err)
	assert.Equal(t, "123", v1.Name)

	err = json.Unmarshal([]byte(variant2), &dst)
	assert.NoError(t, err)
	v2, err := dst.AsOneOfVariant2()
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, v2)

	err = json.Unmarshal([]byte(variant3), &dst)
	assert.NoError(t, err)
	v3, err := dst.AsOneOfVariant3()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant3(true), v3)

	err = dst.FromOneOfVariant1(OneOfVariant1{Name: "123"})
	assert.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant1), marshaled)

	err = dst.FromOneOfVariant2([]int{1, 2, 3})
	assert.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant2), marshaled)

	err = dst.FromOneOfVariant3(true)
	assert.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant3), marshaled)
}

func TestOneOfWithDiscriminator(t *testing.T) {
	const variant4 = `{"discriminator": "v4", "name": "123"}`
	const variant5 = `{"discriminator": "v5", "id": 123}`
	var dst OneOfObject6

	err := json.Unmarshal([]byte(variant4), &dst)
	assert.NoError(t, err)
	discriminator, err := dst.Discriminator()
	assert.NoError(t, err)
	assert.Equal(t, "v4", discriminator)
	v4, err := dst.ValueByDiscriminator()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant4{Discriminator: "v4", Name: "123"}, v4)

	err = json.Unmarshal([]byte(variant5), &dst)
	assert.NoError(t, err)
	discriminator, err = dst.Discriminator()
	assert.NoError(t, err)
	assert.Equal(t, "v5", discriminator)
	v5, err := dst.ValueByDiscriminator()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant5{Discriminator: "v5", Id: 123}, v5)

	// discriminator value will be filled by the generated code
	err = dst.FromOneOfVariant4(OneOfVariant4{Name: "123"})
	assert.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant4), marshaled)

	err = dst.FromOneOfVariant5(OneOfVariant5{Id: 123})
	assert.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant5), marshaled)
}

func TestOneOfWithDiscriminator_PartialMapping(t *testing.T) {
	const variant4 = `{"discriminator": "v4", "name": "123"}`
	const variant5 = `{"discriminator": "OneOfVariant5", "id": 321}`
	var dst OneOfObject61

	err := json.Unmarshal([]byte(variant4), &dst)
	assert.NoError(t, err)
	discriminator, err := dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "v4", discriminator)
	v4, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant4{Discriminator: "v4", Name: "123"}, v4)

	err = json.Unmarshal([]byte(variant5), &dst)
	require.NoError(t, err)
	discriminator, err = dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "OneOfVariant5", discriminator)
	v5, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant5{Discriminator: "OneOfVariant5", Id: 321}, v5)

	// discriminator value will be filled by the generated code
	err = dst.FromOneOfVariant4(OneOfVariant4{Name: "123"})
	require.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant4), marshaled)

	err = dst.FromOneOfVariant5(OneOfVariant5{Id: 321})
	require.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant5), marshaled)
}

func TestOneOfWithDiscriminator_SchemaNameUsed(t *testing.T) {
	const variant4 = `{"discriminator": "variant_four", "name": "789"}`
	const variant51 = `{"discriminator": "one_of_variant51", "id": 987}`
	var dst OneOfObject62

	err := json.Unmarshal([]byte(variant4), &dst)
	assert.NoError(t, err)
	discriminator, err := dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "variant_four", discriminator)
	v4, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant4{Discriminator: "variant_four", Name: "789"}, v4)

	err = json.Unmarshal([]byte(variant51), &dst)
	require.NoError(t, err)
	discriminator, err = dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "one_of_variant51", discriminator)
	v5, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant51{Discriminator: "one_of_variant51", Id: 987}, v5)

	// discriminator value will be filled by the generated code
	err = dst.FromOneOfVariant4(OneOfVariant4{Name: "789"})
	require.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant4), marshaled)

	err = dst.FromOneOfVariant51(OneOfVariant51{Id: 987})
	require.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant51), marshaled)
}

func TestOneOfWithDiscriminator_FullImplicitMapping(t *testing.T) {
	const variant4 = `{"discriminator": "OneOfVariant4", "name": "456"}`
	const variant5 = `{"discriminator": "OneOfVariant5", "id": 654}`
	var dst OneOfObject5

	err := json.Unmarshal([]byte(variant4), &dst)
	assert.NoError(t, err)
	discriminator, err := dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "OneOfVariant4", discriminator)
	v4, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant4{Discriminator: "OneOfVariant4", Name: "456"}, v4)

	err = json.Unmarshal([]byte(variant5), &dst)
	require.NoError(t, err)
	discriminator, err = dst.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "OneOfVariant5", discriminator)
	v5, err := dst.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, OneOfVariant5{Discriminator: "OneOfVariant5", Id: 654}, v5)

	// discriminator value will be filled by the generated code
	err = dst.FromOneOfVariant4(OneOfVariant4{Name: "456"})
	require.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant4), marshaled)

	err = dst.FromOneOfVariant5(OneOfVariant5{Id: 654})
	require.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(variant5), marshaled)
}

func TestOneOfWithFixedProperties(t *testing.T) {
	const variant1 = "{\"type\": \"v1\", \"name\": \"123\"}"
	const variant6 = "{\"type\": \"v6\", \"values\": [1, 2, 3]}"
	var dst OneOfObject9

	err := json.Unmarshal([]byte(variant1), &dst)
	assert.NoError(t, err)
	discriminator, err := dst.Discriminator()
	assert.NoError(t, err)
	assert.Equal(t, "v1", discriminator)
	v1, err := dst.ValueByDiscriminator()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant1{Name: "123"}, v1)

	err = json.Unmarshal([]byte(variant6), &dst)
	assert.NoError(t, err)
	discriminator, err = dst.Discriminator()
	assert.NoError(t, err)
	assert.Equal(t, "v6", discriminator)
	v2, err := dst.AsOneOfVariant6()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant6{[]int{1, 2, 3}}, v2)

	err = dst.FromOneOfVariant1(OneOfVariant1{Name: "123"})
	assert.NoError(t, err)
	marshaled, err := json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant1), marshaled)

	err = dst.FromOneOfVariant6(OneOfVariant6{[]int{1, 2, 3}})
	assert.NoError(t, err)
	marshaled, err = json.Marshal(dst)
	assert.NoError(t, err)
	assertJsonEqual(t, []byte(variant6), marshaled)
}

func TestAnyOf(t *testing.T) {
	const anyOfStr = `{"discriminator": "all", "name": "123", "id": 456}`

	var dst AnyOfObject1
	err := json.Unmarshal([]byte(anyOfStr), &dst)
	assert.NoError(t, err)

	v4, err := dst.AsOneOfVariant4()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant4{Discriminator: "all", Name: "123"}, v4)

	v5, err := dst.AsOneOfVariant5()
	assert.NoError(t, err)
	assert.Equal(t, OneOfVariant5{Discriminator: "all", Id: 456}, v5)
}

func TestOneOfWithAdditional(t *testing.T) {
	x := OneOfObject13{
		AdditionalProperties: map[string]any{"x": "y"},
	}
	err := x.MergeOneOfVariant1(OneOfVariant1{Name: "test-name"})
	require.NoError(t, err)
	b, err := json.Marshal(x)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":"y", "name":"test-name", "type":"v1"}`, string(b))
	var y OneOfObject13
	err = json.Unmarshal(b, &y)
	require.NoError(t, err)
	assert.Equal(t, x.Type, y.Type)
	xVariant, err := x.AsOneOfVariant1()
	require.NoError(t, err)
	yVariant, err := y.AsOneOfVariant1()
	require.NoError(t, err)
	assert.Equal(t, xVariant, yVariant)
	xAdditional, ok := x.Get("x")
	assert.True(t, ok)
	yAdditional, ok := y.Get("x")
	assert.True(t, ok)
	assert.Equal(t, xAdditional, yAdditional)
	b, err = json.Marshal(y)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":"y", "name":"test-name", "type":"v1"}`, string(b))
}

func TestMarshalWhenNoUnionValueSet(t *testing.T) {
	const expected = `{}`

	var dst OneOfObject10

	bytes, err := dst.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, expected, string(bytes))
}
