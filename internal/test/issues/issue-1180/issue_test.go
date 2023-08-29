package issue1180

import (
	"testing"
)

// TestIssue1180 validates that the parameter `param` is a string type, rather than an int, as we should prioritise the `param` definition closest to the path
func TestIssue1180(t *testing.T) {
	_, _ = NewGetSimplePrimitiveRequest("http://example.com/", "test-string")
}
