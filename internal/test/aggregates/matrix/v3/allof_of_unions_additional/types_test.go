package matrixv3allofofunionsadditional

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Setting one union's variant keeps the other union's data and the caller's
// additional properties, including ones changed since parsing.
func TestFromKeepsAdditionalProperties(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"card": "4111", "address": "1 Main St", "note": "leave at door"}`), &s))
	s.AdditionalProperties["note"] = "ring twice"

	require.NoError(t, s.FromPickup(Pickup{Store: "Downtown"}))
	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"card": "4111", "store": "Downtown", "note": "ring twice"}`, string(b))

	require.NoError(t, s.FromPayment(Payment{}))
	b, err = json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"store": "Downtown", "note": "ring twice"}`, string(b))
}
