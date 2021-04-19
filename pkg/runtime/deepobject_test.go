package runtime

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InnerObject struct {
	Name string
	ID   int
}

// These are all possible field types, mandatory and optional.
type AllFields struct {
	I   int          `json:"i"`
	Oi  *int         `json:"oi,omitempty"`
	F   float32      `json:"f"`
	Of  *float32     `json:"of,omitempty"`
	B   bool         `json:"b"`
	Ob  *bool        `json:"ob,omitempty"`
	As  []string     `json:"as"`
	Oas *[]string    `json:"oas,omitempty"`
	O   InnerObject  `json:"o"`
	Oo  *InnerObject `json:"oo,omitempty"`
	D   MockBinder   `json:"d"`
	Od  *MockBinder  `json:"od,omitempty"`
}

func TestDeepObject(t *testing.T) {
	oi := int(5)
	of := float32(3.7)
	ob := true
	oas := []string{"foo", "bar"}
	oo := InnerObject{
		Name: "Marcin Romaszewicz",
		ID:   123,
	}
	d := MockBinder{Time: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)}

	srcObj := AllFields{
		I:   12,
		Oi:  &oi,
		F:   4.2,
		Of:  &of,
		B:   true,
		Ob:  &ob,
		As:  []string{"hello", "world"},
		Oas: &oas,
		O: InnerObject{
			Name: "Joe Schmoe",
			ID:   456,
		},
		Oo: &oo,
		D:  d,
		Od: &d,
	}

	marshaled, err := MarshalDeepObject(srcObj, "p")
	require.NoError(t, err)
	t.Log(marshaled)

	params := make(url.Values)
	marshaledParts := strings.Split(marshaled, "&")
	for _, p := range marshaledParts {
		parts := strings.Split(p, "=")
		require.Equal(t, 2, len(parts))
		params.Set(parts[0], parts[1])
	}

	var dstObj AllFields
	err = UnmarshalDeepObject(&dstObj, "p", params)
	require.NoError(t, err)
	assert.EqualValues(t, srcObj, dstObj)
}
