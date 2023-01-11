package issue1180

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1180(t *testing.T) {
	req, err := NewGetSimplePrimitiveRequest("http://example.com/", "test-string")
	require.NoError(t, err)
	require.Equal(t, "http://example.com/simplePrimitive/test-string", req.URL.String())
}
