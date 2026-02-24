package issue2185

import (
	"testing"

	"github.com/oapi-codegen/nullable"
	"github.com/stretchr/testify/require"
)

func TestContainer_UsesNullableType(t *testing.T) {
	c := Container{
		MayBeNull: []nullable.Nullable[string]{
			nullable.NewNullNullable[string](),
		},
	}

	require.Len(t, c.MayBeNull, 1)
	require.True(t, c.MayBeNull[0].IsNull())
}
