package matrixv3allofofunionscloseddiscriminator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The discriminator is Subject's, not Payment's: AsPayment leaves it out, so
// Payment's closed variants read what it returns, also after From*.
func TestPlacedDiscriminatorIsLeftOutOfTheUnion(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"method": "card", "card": "4111", "address": "1 Main St"}`), &s))
	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, Card{Card: "4111"}, v)
	payment, err := s.AsPayment()
	require.NoError(t, err)
	card, err := payment.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "4111", card.Card)

	var built Subject
	require.NoError(t, built.FromCard(Card{Card: "1"}))
	require.NoError(t, built.FromCourier(Courier{Address: "a"}))
	payment, err = built.AsPayment()
	require.NoError(t, err)
	card, err = payment.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "1", card.Card)
	delivery, err := built.AsDelivery()
	require.NoError(t, err)
	courier, err := delivery.AsCourier()
	require.NoError(t, err)
	assert.Equal(t, "a", courier.Address)
}
