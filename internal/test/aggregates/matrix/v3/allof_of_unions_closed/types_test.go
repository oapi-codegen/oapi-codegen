package matrixv3allofofunionsclosed

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A closed variant of one union allows the other union's keys and the
// object's own property, and rejects its own union's other variant's keys.
func TestClosedVariantInAnAllOfOfUnions(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"card": "4111", "address": "1 Main St", "orderId": "o1"}`), &s))
	card, err := s.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "4111", card.Card)
	courier, err := s.AsCourier()
	require.NoError(t, err)
	assert.Equal(t, "1 Main St", courier.Address)

	_, err = s.AsTransfer()
	assert.EqualError(t, err, `Transfer doesn't allow the property "card"`)
	_, err = s.AsPickup()
	assert.EqualError(t, err, `Pickup doesn't allow the property "address"`)

	require.NoError(t, json.Unmarshal([]byte(`{"card": "4111", "iban": "DE00", "address": "1 Main St"}`), &s))
	_, err = s.AsCard()
	assert.EqualError(t, err, `Card doesn't allow the property "iban"`)
}

// AsPayment leaves out the other union's keys and the object's own property,
// so Payment's closed variants read it.
func TestUnionOfAnAllOfOfUnionsReadsItsOwnData(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"card": "4111", "address": "1 Main St", "orderId": "o1"}`), &s))
	payment, err := s.AsPayment()
	require.NoError(t, err)
	card, err := payment.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "4111", card.Card)
	delivery, err := s.AsDelivery()
	require.NoError(t, err)
	courier, err := delivery.AsCourier()
	require.NoError(t, err)
	assert.Equal(t, "1 Main St", courier.Address)
}
