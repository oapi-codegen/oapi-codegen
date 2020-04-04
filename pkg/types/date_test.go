package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate_MarshalJSON(t *testing.T) {
	testDate := time.Date(2019, 4, 1, 0, 0, 0, 0, time.UTC)
	b := struct {
		DateField Date `json:"date"`
	}{
		DateField: Date{testDate},
	}
	jsonBytes, err := json.Marshal(b)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"date":"2019-04-01"}`, string(jsonBytes))
}

func TestDate_UnmarshalJSON(t *testing.T) {
	testDate := time.Date(2019, 4, 1, 0, 0, 0, 0, time.UTC)
	jsonStr := `{"date":"2019-04-01"}`
	b := struct {
		DateField Date `json:"date"`
	}{}
	err := json.Unmarshal([]byte(jsonStr), &b)
	assert.NoError(t, err)
	assert.Equal(t, testDate, b.DateField.Time)
}
