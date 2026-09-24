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

// TestFromReplacesAdoptedProperty: a value the union adopted from an earlier
// From* is replaced by the next one, rather than overwriting it with the old
// value when marshaling.
func TestFromReplacesAdoptedProperty(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromCat(Cat{Id: "old", Meow: "purr"}))
	require.NoError(t, s.FromCat(Cat{Id: "new", Meow: "hiss"}))
	assert.Equal(t, "new", s.Id)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "new", "meow": "hiss"}`, string(b))
}

// TestFromReplacesUnmarshaledProperty: the same after UnmarshalJSON, which
// fills the union and its own fields from the same document.
func TestFromReplacesUnmarshaledProperty(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"id": "old", "bark": "woof"}`), &s))
	require.NoError(t, s.FromCat(Cat{Id: "new", Meow: "purr"}))

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "new", "meow": "purr"}`, string(b))
}

// TestFromKeepsPropertyChangedByCaller: once the caller has changed the
// union's own field, a later From* does not overwrite it.
func TestFromKeepsPropertyChangedByCaller(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromCat(Cat{Id: "old", Meow: "purr"}))
	s.Id = "custom"
	require.NoError(t, s.FromCat(Cat{Id: "new", Meow: "hiss"}))
	assert.Equal(t, "custom", s.Id)
}
