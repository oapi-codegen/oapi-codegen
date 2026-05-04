package issue2031

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	value := ArrayContainer{}
	content, err := json.Marshal(value)
	require.NoError(t, err)
	// the _optional array_ should be _omitted_ when null, not marshaled as null
	// (which is not valid per the schema)
	assert.Equal(t, "{}", string(content))
}
