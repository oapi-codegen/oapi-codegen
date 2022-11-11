package typemappings

import (
	"testing"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	var c Client

	// it is the UUID type, and it is not defined with the pointer
	assert.Equal(t, types.UUID{}, c.Uuid)
}
