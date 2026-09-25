package matrixv3oneofdiscriminatorint64

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The discriminator is read as an int64, so 2^53+1 is not 2^53, as it would
// be as a float64, and From* stamps it exactly.
func TestInt64Discriminator(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"id": 9007199254740993, "size": "b"}`), &s))
	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, Big{Id: 9007199254740993, Size: "b"}, v)

	require.NoError(t, json.Unmarshal([]byte(`{"id": 9007199254740992, "size": "b"}`), &s))
	_, err = s.ValueByDiscriminator()
	assert.EqualError(t, err, "unknown discriminator value: 9007199254740992")

	require.NoError(t, s.FromBig(Big{Size: "b"}))
	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"id":9007199254740993`)
}

// An integer discriminator written with a fraction isn't read as one.
func TestInt64DiscriminatorFraction(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"id": 1.0, "name": "a"}`), &s))
	_, err := s.ValueByDiscriminator()
	assert.Error(t, err)
}
