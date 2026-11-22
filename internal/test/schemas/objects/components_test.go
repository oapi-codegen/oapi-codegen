package schemasobjects

// additionalProperties + raw-JSON object tests folded from components/components.yaml.

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

// TestAdditionalPropertiesJSONCodec checks that the MarshalJSON and
// UnmarshalJSON of an additionalProperties type use the codec set with
// runtime.SetJSONCodec. They run inside the codec's own Marshal and Unmarshal
// of the value, so this is also the re-entrant case.
func TestAdditionalPropertiesJSONCodec(t *testing.T) {
	var codec countingCodec
	runtime.SetJSONCodec(&codec)
	t.Cleanup(func() { runtime.SetJSONCodec(nil) })

	const buf = `{"name": "bob", "id": 5, "optional":"yes", "additional": 42}`
	var dst AdditionalPropertiesObject1
	require.NoError(t, runtime.JSONUnmarshal([]byte(buf), &dst))
	additional, found := dst.Get("additional")
	assert.True(t, found)
	assert.Equal(t, 42, additional)
	// One call from here, the rest from UnmarshalJSON.
	assert.Greater(t, codec.unmarshals, 1)

	marshaled, err := runtime.JSONMarshal(dst)
	require.NoError(t, err)
	assertJsonEqual(t, []byte(buf), marshaled)
	// One call from here, the rest from MarshalJSON.
	assert.Greater(t, codec.marshals, 1)
}

func assertJsonEqual(t *testing.T, j1 []byte, j2 []byte) {
	t.Helper()
	assert.JSONEq(t, string(j1), string(j2))
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

	// test that additionalProperties that reference a schema work when unmarshaling
	bossSchema := SchemaObject{
		FirstName: "bob",
		Role:      "warehouse manager",
	}

	buf2 := `{"boss": { "firstName": "bob", "role": "warehouse manager" }, "employee": { "firstName": "kevin", "role": "warehouse"}}`
	var obj5 AdditionalPropertiesObject5
	err = json.Unmarshal([]byte(buf2), &obj5)
	assert.NoError(t, err)
	assert.Equal(t, bossSchema, obj5["boss"])

	bossSchemaNullable := &SchemaObjectNullable{
		FirstName: "bob",
		Role:      "warehouse manager",
	}

	buf3 := `{"boss": { "firstName": "bob", "role": "warehouse manager" }, "employee": null}`
	var obj7 AdditionalPropertiesObject7
	err = json.Unmarshal([]byte(buf3), &obj7)
	assert.NoError(t, err)
	employee, ok := obj7["employee"]
	assert.True(t, ok)
	assert.Equal(t, bossSchemaNullable, obj7["boss"])
	assert.Nil(t, employee)
}
