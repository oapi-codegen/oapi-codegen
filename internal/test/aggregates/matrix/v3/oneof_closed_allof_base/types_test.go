package matrixv3oneofclosedallofbase

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #668: AsBase fails on Extended's data, so the two can be told apart.
// Extended is closed by Base's additionalProperties: false, and allows the
// keys it adds.
func TestBaseRejectsExtendedData(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"id": "2", "extra": "more"}`), &s))
	_, err := s.AsBase()
	assert.EqualError(t, err, `Base doesn't allow the property "extra"`)
	extended, err := s.AsExtended()
	require.NoError(t, err)
	assert.Equal(t, "more", extended.Extra)

	require.NoError(t, json.Unmarshal([]byte(`{"id": "2", "extra": "more", "other": 1}`), &s))
	_, err = s.AsExtended()
	assert.EqualError(t, err, `Extended doesn't allow the property "other"`)
}
