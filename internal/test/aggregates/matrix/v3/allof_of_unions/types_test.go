package matrixv3allofofunions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Subject is a Payment and a Delivery at once. Setting a variant of one
// union replaces that union's keys and keeps the other's.
func TestFromKeepsTheOtherUnion(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromCard(Card{Card: "4111"}))
	require.NoError(t, s.FromCourier(Courier{Address: "1 Main St"}))
	assert.JSONEq(t, `{"card": "4111", "address": "1 Main St"}`, marshal(t, s))

	require.NoError(t, s.FromTransfer(Transfer{Iban: "DE00"}))
	assert.JSONEq(t, `{"iban": "DE00", "address": "1 Main St"}`, marshal(t, s))

	require.NoError(t, s.FromPickup(Pickup{Store: "Downtown"}))
	assert.JSONEq(t, `{"iban": "DE00", "store": "Downtown"}`, marshal(t, s))

	transfer, err := s.AsTransfer()
	require.NoError(t, err)
	assert.Equal(t, "DE00", transfer.Iban)
	pickup, err := s.AsPickup()
	require.NoError(t, err)
	assert.Equal(t, "Downtown", pickup.Store)
}

// A member that is a $ref to a union gets As* and From* for that union type,
// which also keep the other union's data.
func TestUnionMemberAccessors(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"card": "4111", "address": "1 Main St"}`), &s))

	payment, err := s.AsPayment()
	require.NoError(t, err)
	card, err := payment.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "4111", card.Card)

	var transfer Payment
	require.NoError(t, transfer.FromTransfer(Transfer{Iban: "DE00"}))
	require.NoError(t, s.FromPayment(transfer))
	assert.JSONEq(t, `{"iban": "DE00", "address": "1 Main St"}`, marshal(t, s))

	var pickup Delivery
	require.NoError(t, pickup.FromPickup(Pickup{Store: "Downtown"}))
	require.NoError(t, s.FromDelivery(pickup))
	assert.JSONEq(t, `{"iban": "DE00", "store": "Downtown"}`, marshal(t, s))
}

// As* for a union type leaves out the other union's data, so setting it back
// later doesn't bring back a delivery that has since changed; a zero union
// clears only its own data.
func TestUnionMemberAccessorsKeepToThemselves(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromCard(Card{Card: "4111"}))
	require.NoError(t, s.FromCourier(Courier{Address: "1 Main St"}))
	payment, err := s.AsPayment()
	require.NoError(t, err)
	assert.JSONEq(t, `{"card": "4111"}`, marshal(t, payment))

	require.NoError(t, s.FromPickup(Pickup{Store: "Downtown"}))
	require.NoError(t, s.FromPayment(payment))
	assert.JSONEq(t, `{"card": "4111", "store": "Downtown"}`, marshal(t, s))

	require.NoError(t, s.FromPayment(Payment{}))
	assert.JSONEq(t, `{"store": "Downtown"}`, marshal(t, s))
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}
