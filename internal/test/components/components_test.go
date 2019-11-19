package components

import (
	"encoding/json"
	"encoding/xml"
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

func assertXmlEqual(t *testing.T, j1 []byte, j2 []byte) {
	var v1, v2 interface{}

	err := xml.Unmarshal(j1, &v1)
	assert.NoError(t, err)

	err = xml.Unmarshal(j2, &v2)
	assert.NoError(t, err)

	assert.EqualValues(t, v1, v2)
}

func TestRawJSONwithXML(t *testing.T) {
	// Check raw json unmarshaling
	const buf = `<name>bob</name><value1>{"present":true}</value1>`
	var dst ObjectWithJsonField
	err := xml.Unmarshal([]byte(buf), &dst)
	assert.NoError(t, err)

	buf2, err := xml.Marshal(dst)
	assert.NoError(t, err)

	assertXmlEqual(t, []byte(buf), buf2)

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
	assert.Equal(t, bossSchema, obj5.AdditionalProperties["boss"])
}

func TestAdditionalPropertiesXMLRaw(t *testing.T) {
	t.SkipNow()
	// Check raw json unmarshaling
	const buf = `<AdditionalPropertiesObject1><name>bob</name><id>5</id><optional>yes</optional><additional>42</additional></AdditionalPropertiesObject1>`
	var dst AdditionalPropertiesObject1
	err := xml.Unmarshal([]byte(buf), &dst)
	assert.NoError(t, err)

	buf2, err := xml.Marshal(dst)
	assert.NoError(t, err)

	assertXmlEqual(t, []byte(buf), buf2)

}

func TestAdditionalPropertiesXML(t *testing.T) {
	t.SkipNow()
	buf := `<AdditionalPropertiesObject1><name>bob</name><id>5</id><optional>yes</optional><additional>42</additional></AdditionalPropertiesObject1>`
	var dst AdditionalPropertiesObject1
	err := xml.Unmarshal([]byte(buf), &dst)
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

	buf2 := `<AdditionalPropertiesObject5><boss><firstName>bob</firstName><role>warehouse manager</role></boss><employee><firstName>kevin</firstName><role>warehouse</role></employee></AdditionalPropertiesObject5>`
	var obj5 AdditionalPropertiesObject5
	err = xml.Unmarshal([]byte(buf2), &obj5)
	assert.NoError(t, err)
	assert.Equal(t, bossSchema, obj5.AdditionalProperties["boss"])
}
