package components

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawJSON(t *testing.T) {
	// Check raw json unmarshaling
	const buf = `{"name":"bob","value1":{"present":true}}`
	var dst ObjectWithJsonField
	err := json.Unmarshal([]byte(buf), &dst)
	assert.NoError(t, err)

	buf2, err := json.Marshal(dst)
	assert.NoError(t, err)
	assert.EqualValues(t, string(buf), string(buf2))

}
