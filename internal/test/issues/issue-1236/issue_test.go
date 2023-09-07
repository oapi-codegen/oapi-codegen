package issue1236_test

import (
	"encoding/json"
	"testing"

	issue1236 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1236"
	"github.com/stretchr/testify/assert"
)

// Test treatment additionalProperties in mergeOpenapiSchemas()
func TestIssue1236(t *testing.T) {
	var without issue1236.WithoutAdditionalProperties
	var i int
	assert.IsType(t, i, without.Required)
	assert.IsType(t, &i, without.Optional)
	assert.IsType(t, &i, without.ReadOnly)
	assert.IsType(t, &i, without.WriteOnly)
	assert.IsType(t, &i, without.Nullable)
	buf, err := json.Marshal(without)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Required": 0, "Nullable": null}`, string(buf))

	var with issue1236.WithAdditionalProperties
	assert.IsType(t, i, with.Required)
	assert.IsType(t, &i, with.Optional)
	assert.IsType(t, &i, with.ReadOnly)
	assert.IsType(t, &i, with.WriteOnly)
	assert.IsType(t, &i, with.Nullable)
	with.Set("Extra", 0)
	buf, err = json.Marshal(with)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Required": 0, "Nullable": null, "Extra": 0}`, string(buf))
}
