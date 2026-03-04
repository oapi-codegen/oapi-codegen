package xoapicodegenonlyhonourgoname

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeWithUnexportedField(t *testing.T) {
	var v TypeWithUnexportedField

	err := json.Unmarshal([]byte(`{"id": "some-id"}`), &v)
	require.NoError(t, err)

	// this field will never be unmarshaled
	require.Nil(t, v.accountIdentifier)
}
