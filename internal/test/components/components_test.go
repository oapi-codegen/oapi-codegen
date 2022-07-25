package components

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertJsonEqual(t *testing.T, j1 []byte, j2 []byte) {
	var v1, v2 interface{}

	err := json.Unmarshal(j1, &v1)
	assert.NoError(t, err)

	err = json.Unmarshal(j2, &v2)
	assert.NoError(t, err)

	assert.EqualValues(t, v1, v2)
}

func TestRawJSON(t *testing.T) {
	// Check raw json unmarshaling
	const buf = `{"name":"bob","value1":{"present":true}}`
	var dst ObjectWithJsonField
	err := json.Unmarshal([]byte(buf), &dst)
	assert.NoError(t, err)

	buf2, err := json.Marshal(dst)
	assert.NoError(t, err)

	assertJsonEqual(t, []byte(buf), buf2)

}

func TestAdditionalProperties(t *testing.T) {
	buf := `{"name": "bob", "id": 5, "optional":"yes", "additional": 42}`
	var dst AdditionalPropertiesObject1
	err := json.Unmarshal([]byte(buf), &dst)
	assert.NoError(t, err)
	assert.Equal(t, "bob", dst.Name)
	assert.Equal(t, 5, dst.Id)
	assert.Equal(t, "yes", *dst.Optional)
	additional, found := dst.Get("additional")
	assert.True(t, found)
	assert.Equal(t, 42, additional)

	obj4 := AdditionalPropertiesObject4{
		Name: "bob",
	}
	obj4.Set("add1", "hi")
	obj4.Set("add2", 7)

	foo, found := obj4.Get("add1")
	assert.True(t, found)
	assert.EqualValues(t, "hi", foo)
	foo, found = obj4.Get("add2")
	assert.True(t, found)
	assert.EqualValues(t, 7, foo)

	// test that additionalProperties that reference a schema work when unmarshalling
	bossSchema := SchemaObject{
		FirstName: "bob",
		Role:      "warehouse manager",
	}

	buf2 := `{"boss": { "firstName": "bob", "role": "warehouse manager" }, "employee": { "firstName": "kevin", "role": "warehouse"}}`
	var obj5 AdditionalPropertiesObject5
	err = json.Unmarshal([]byte(buf2), &obj5)
	assert.NoError(t, err)
	assert.Equal(t, bossSchema, obj5["boss"])
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
	assert.Equal(t, OneOfVariant2([]int{1, 2, 3}), v2)

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

func TestMarshalWhenNoUnionValueSet(t *testing.T) {
	const expected = `{"one":null,"three":null,"two":null}`

	var dst OneOfObject10

	bytes, err := dst.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, expected, string(bytes))
}
