package issue1825

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverlayApply(t *testing.T) {
	spec, err := GetSwagger()
	require.NoError(t, err)

	require.Equal(t, spec.Info.Extensions["x-overlay-applied"], "structured-overlay")
}
