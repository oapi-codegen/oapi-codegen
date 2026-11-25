package matrixv3nullableunionnestedclosed

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNestedNullableUnionClosedVariants: the inline oneOf the null branch
// wraps keeps the keys its closed variants allow, so As* of a variant rejects
// the other variant's keys (issue #2565).
func TestNestedNullableUnionClosedVariants(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"meow": "purr"}`), &s))
	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, Cat{Meow: "purr"}, cat)
	_, err = s.AsDog()
	assert.EqualError(t, err, `Dog doesn't allow the property "meow"`)
}
