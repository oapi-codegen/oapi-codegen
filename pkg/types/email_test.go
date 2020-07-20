package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmail_MarshalJSON(t *testing.T) {
	testEmail := "gaben@valvesoftware.com"
	b := struct {
		EmailField Email `json:"email"`
	}{
		EmailField: Email(testEmail),
	}
	jsonBytes, err := json.Marshal(b)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"email":"gaben@valvesoftware.com"}`, string(jsonBytes))
}

func TestEmail_UnmarshalJSON(t *testing.T) {
	testEmail := Email("gaben@valvesoftware.com")
	jsonStr := `{"email":"gaben@valvesoftware.com"}`
	b := struct {
		EmailField Email `json:"email"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.NoError(t, err)
	assert.Equal(t, testEmail, b.EmailField)
}
