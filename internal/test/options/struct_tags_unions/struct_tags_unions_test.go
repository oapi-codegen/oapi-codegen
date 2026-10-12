package optionsstructtagsunions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keys(t *testing.T, v any) []string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	var object map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(b, &object))
	var names []string
	for name := range object {
		names = append(names, name)
	}
	return names
}

// A closed variant's As* reads what its From* wrote under the keys the json
// template gives, and another variant's As* rejects them.
func TestClosedVariantAllowsTemplatedKeys(t *testing.T) {
	holder := "Ann"
	var p Payment
	require.NoError(t, p.FromCard(Card{Card: "4111", Holder: &holder}))
	assert.ElementsMatch(t, []string{"card_wire", "opt_holder_wire"}, keys(t, p))
	card, err := p.AsCard()
	require.NoError(t, err)
	assert.Equal(t, "4111", card.Card)
	assert.Equal(t, "Ann", *card.Holder)
	_, err = p.AsTransfer()
	assert.EqualError(t, err, `Transfer doesn't allow the property "card_wire"`)
}

// From* of one union in an allOf of unions replaces that union's keys as the
// json template gives them, and keeps the other union's.
func TestUnionComponentsReplaceTemplatedKeys(t *testing.T) {
	holder := "Ann"
	var o Order
	require.NoError(t, o.FromCard(Card{Card: "4111", Holder: &holder}))
	require.NoError(t, o.FromCourier(Courier{Address: "1 Main St", Holder: "Bob"}))
	assert.ElementsMatch(t, []string{"card_wire", "opt_holder_wire", "address_wire", "holder_wire"}, keys(t, o))
	_, err := o.AsCard()
	require.NoError(t, err)

	// Card's optional holder is opt_holder_wire and Courier's required one
	// holder_wire, so opt_holder_wire is Payment's alone and goes with Card.
	require.NoError(t, o.FromTransfer(Transfer{Iban: "DE00"}))
	assert.ElementsMatch(t, []string{"iban_wire", "address_wire", "holder_wire"}, keys(t, o))
	transfer, err := o.AsTransfer()
	require.NoError(t, err)
	assert.Equal(t, "DE00", transfer.Iban)
	courier, err := o.AsCourier()
	require.NoError(t, err)
	assert.Equal(t, "1 Main St", courier.Address)
}
