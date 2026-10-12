package matrixv3oneofdiscriminatornumber

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The discriminator's number can be written several ways; ValueByDiscriminator
// compares it as a number, while Discriminator returns the text as written.
func TestNumberSpellings(t *testing.T) {
	for text, want := range map[string]any{
		`1`: Subject0{}, `1.0`: Subject0{}, `1e0`: Subject0{},
		`2.5`: Subject1{}, `2.50`: Subject1{}, `25e-1`: Subject1{},
	} {
		var s Subject
		require.NoError(t, json.Unmarshal([]byte(`{"version": `+text+`, "name": "a", "fullName": "a b"}`), &s), text)
		d, err := s.Discriminator()
		require.NoError(t, err)
		assert.Equal(t, text, d)
		v, err := s.ValueByDiscriminator()
		require.NoError(t, err, text)
		assert.IsType(t, want, v, text)
	}
}

// From* stamps each variant's number.
func TestNumberStamps(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromSubject1(Subject1{FullName: "a b"}))
	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"version": 2.5, "fullName": "a b"}`, string(b))
}
