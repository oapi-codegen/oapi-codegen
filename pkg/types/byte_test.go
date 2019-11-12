package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByte_MarshalJSON(t *testing.T) {
	b := struct {
		ByteField Byte `json:"bytes"`
	}{
		ByteField: Byte{bytes:[]byte{1,2,3}},
	}
	jsonBytes, err := json.Marshal(b)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"bytes":"AQID"}`, string(jsonBytes))
}

func TestByte_UnmarshalJSON(t *testing.T) {
	jsonStr := `{"bytes":"AQID"}`
	b := struct {
		ByteField Byte `json:"bytes"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.NoError(t, err)
	assert.Equal(t, []byte{1,2,3}, b.ByteField.bytes)
}
