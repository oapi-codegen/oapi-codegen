package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ json.Marshaler = (*File)(nil)
var _ json.Unmarshaler = (*File)(nil)

func TestFileJSON(t *testing.T) {
	type Object struct {
		BinaryField File `json:"binary_field"`
	}

	// Check whether we encode JSON properly.
	var o Object
	o.BinaryField.InitFromBytes([]byte("hello"), "")
	buf, err := json.Marshal(o)
	require.NoError(t, err)
	t.Log(string(buf))

	// Decode the same object back into File, ensure result is correct.
	var o2 Object
	err = json.Unmarshal(buf, &o2)
	require.NoError(t, err)
	o2Bytes, err := o2.BinaryField.Bytes()
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), o2Bytes)

	// Ensure it also works via pointer.
	type Object2 struct {
		BinaryFieldPtr *File `json:"binary_field"`
	}

	var o3 Object2
	var f File
	f.InitFromBytes([]byte("hello"), "")
	o3.BinaryFieldPtr = &f
	buf, err = json.Marshal(o)
	require.NoError(t, err)
	t.Log(string(buf))

	var o4 Object2
	err = json.Unmarshal(buf, &o4)
	require.NoError(t, err)
	o4Bytes, err := o4.BinaryFieldPtr.Bytes()
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), o4Bytes)

}
