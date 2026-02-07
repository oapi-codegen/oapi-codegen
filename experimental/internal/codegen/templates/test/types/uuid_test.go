package types

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUID_MarshalJSON_Zero(t *testing.T) {
	var testUUID UUID
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{
		UUIDField: testUUID,
	}
	marshaled, err := json.Marshal(b)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"uuid":"00000000-0000-0000-0000-000000000000"}`, string(marshaled))
}

func TestUUID_MarshalJSON_Pass(t *testing.T) {
	testUUID := uuid.MustParse("9cb14230-b640-11ec-b909-0242ac120002")
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{
		UUIDField: testUUID,
	}
	jsonBytes, err := json.Marshal(b)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"uuid":"9cb14230-b640-11ec-b909-0242ac120002"}`, string(jsonBytes))
}

func TestUUID_UnmarshalJSON_Fail(t *testing.T) {
	jsonStr := `{"uuid":"this-is-not-a-uuid"}`
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.Error(t, err)
}

func TestUUID_UnmarshalJSON_Pass(t *testing.T) {
	testUUID := uuid.MustParse("9cb14230-b640-11ec-b909-0242ac120002")
	jsonStr := `{"uuid":"9cb14230-b640-11ec-b909-0242ac120002"}`
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.NoError(t, err)
	assert.Equal(t, testUUID, b.UUIDField)
}
