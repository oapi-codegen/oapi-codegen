package matrixoneofdiscriminatorinteger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegerDiscriminator covers an integer discriminator that the union
// declares itself: From* assigns the union's own field and stamps the JSON.
func TestIntegerDiscriminator(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromV2(V2{FullName: "a b"}))
	assert.Equal(t, 2, s.Version)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"version": 2, "fullName": "a b"}`, string(b))

	d, err := s.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "2", d)

	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, V2{Version: 2, FullName: "a b"}, v)
}
