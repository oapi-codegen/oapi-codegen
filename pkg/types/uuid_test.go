package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID_MarshalJSON_Fail(t *testing.T) {
	testUUID := "this-is-not-a-uuid"
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{
		UUIDField: UUID(testUUID),
	}
	_, err := json.Marshal(b)
	assert.Error(t, err)
}

func TestUUID_MarshalJSON_Pass(t *testing.T) {
	testUUID := "9cb14230-b640-11ec-b909-0242ac120002"
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{
		UUIDField: UUID(testUUID),
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
	testUUID := UUID("9cb14230-b640-11ec-b909-0242ac120002")
	jsonStr := `{"uuid":"9cb14230-b640-11ec-b909-0242ac120002"}`
	b := struct {
		UUIDField UUID `json:"uuid"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.NoError(t, err)
	assert.Equal(t, testUUID, b.UUIDField)
}
