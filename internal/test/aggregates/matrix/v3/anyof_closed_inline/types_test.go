package matrixv3anyofclosedinline

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An anyOf's closed variants are checked as a oneOf's are.
func TestClosedAnyOfVariants(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"email": "a@example.com"}`), &s))
	email, err := s.AsSubject0()
	require.NoError(t, err)
	assert.Equal(t, "a@example.com", email.Email)
	_, err = s.AsSubject1()
	assert.EqualError(t, err, `Subject1 doesn't allow the property "email"`)
}
