package issue579

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAliasedDate(t *testing.T) {
	pet := Pet{}
	err := json.Unmarshal([]byte(`{"born": "2022-05-19", "born_at": "2022-05-20"}`), &pet)
	require.NoError(t, err)
}
