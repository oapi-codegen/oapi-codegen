package matrixv3nullableunionnestedadditionalproperties

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNestedNullableUnionFromDropsPreviousVariantKeys: the inline oneOf the
// null branch wraps keeps the property names its variants declare. After
// UnmarshalJSON the previous variant's keys are in AdditionalProperties too;
// switching variants drops them, and keeps genuine additional properties
// (issue #2565).
func TestNestedNullableUnionFromDropsPreviousVariantKeys(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"bark": "woof", "extra": 7}`), &s))
	require.NoError(t, s.FromCat(Cat{Meow: "purr"}))

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"meow": "purr", "extra": 7}`, string(b))
}
