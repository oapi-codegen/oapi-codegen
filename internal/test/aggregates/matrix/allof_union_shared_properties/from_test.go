package matrixallofunionsharedproperties

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFromAdoptsSharedProperty: the union's own Id must not overwrite
// the variant's id with an empty string.
func TestFromAdoptsSharedProperty(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromCat(Cat{Id: "cat-1", Meow: "purr"}))
	assert.Equal(t, "cat-1", s.Id)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "cat-1", "meow": "purr"}`, string(b))
}

// TestFromKeepsPropertySetByCaller: a value the caller already put in the
// union's own field wins over the variant's, as it always has.
func TestFromKeepsPropertySetByCaller(t *testing.T) {
	s := Subject{Id: "explicit"}
	require.NoError(t, s.FromCat(Cat{Meow: "purr"}))
	assert.Equal(t, "explicit", s.Id)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "explicit", "meow": "purr"}`, string(b))
}

// TestMergeAdoptsSharedProperty covers Merge*, which has the same problem.
func TestMergeAdoptsSharedProperty(t *testing.T) {
	var s Subject
	require.NoError(t, s.MergeDog(Dog{Id: "dog-1", Bark: "woof"}))
	assert.Equal(t, "dog-1", s.Id)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "dog-1", "bark": "woof"}`, string(b))
}
