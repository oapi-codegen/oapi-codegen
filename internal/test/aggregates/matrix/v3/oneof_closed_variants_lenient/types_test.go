package matrixv3oneofclosedvariantslenient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With lenient-union-accessors, a closed variant's As* ignores the keys it
// doesn't declare.
func TestLenientAccessorsIgnoreOtherKeys(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"kind": "dog", "bark": "woof", "extra": 1}`), &s))
	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, Cat{}, cat)
}
