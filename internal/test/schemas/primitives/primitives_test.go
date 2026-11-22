package schemasprimitives

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// issue #579: a Pet whose Born field is typed via a $ref to AliasedDate and
// whose BornAt field is an inline date must both unmarshal without error.
func TestAliasedDate(t *testing.T) {
	pet := Pet{}
	err := json.Unmarshal([]byte(`{"born": "2022-05-19", "born_at": "2022-05-20"}`), &pet)
	require.NoError(t, err)
}
