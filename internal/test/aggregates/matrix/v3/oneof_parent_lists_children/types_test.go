package matrixv3oneofparentlistschildren

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cat is one branch of Subject, not a union of all of them: a plain struct
// with its two fields. This fails to compile otherwise.
var _ = Cat{nil, "Cat"}

// Subject, the parent, is still the union of its children.
func TestParentIsTheUnion(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"petType": "Cat", "meow": "m"}`), &s))
	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	cat, ok := v.(Cat)
	require.True(t, ok, "%T", v)
	assert.Equal(t, "m", *cat.Meow)
}
