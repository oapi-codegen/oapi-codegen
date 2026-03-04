package issue1561

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseBody_DoesNotHaveOptionalPointerToContainerTypes(t *testing.T) {
	pong0 := Pong{
		Ping: "0th",
	}

	pong1 := Pong{
		Ping: "1th",
	}

	slice := []Pong{
		pong0,
		pong1,
	}

	m := map[string]Pong{
		"0": pong0,
		"1": pong1,
	}

	byteData := []byte("some bytes")

	body := ResponseBody{
		RequiredSlice:             slice,
		ASlice:                    slice,
		AMap:                      m,
		UnknownObject:             map[string]any{},
		AdditionalProps:           m,
		ASliceWithAdditionalProps: []map[string]Pong{m},
		Bytes:                     byteData,
		BytesWithOverride:         &byteData,
	}

	assert.NotNil(t, body.RequiredSlice)
	assert.NotZero(t, body.RequiredSlice)

	assert.NotNil(t, body.ASlice)
	assert.NotZero(t, body.ASlice)

	assert.NotNil(t, body.AMap)
	assert.NotZero(t, body.AMap)

	assert.NotNil(t, body.UnknownObject)
	assert.Empty(t, body.UnknownObject)

	assert.NotNil(t, body.AdditionalProps)
	assert.NotZero(t, body.AdditionalProps)

	assert.NotNil(t, body.ASliceWithAdditionalProps)
	assert.NotZero(t, body.ASliceWithAdditionalProps)

	assert.NotNil(t, body.Bytes)
	assert.NotZero(t, body.Bytes)

	assert.NotNil(t, body.BytesWithOverride)
	assert.NotZero(t, body.BytesWithOverride)
}
