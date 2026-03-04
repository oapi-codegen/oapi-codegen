package precedence

import (
	"testing"
)

// TestParameterPrecedence validates that the operation-level parameter definition
// takes priority over the path-level parameter with the same name.
// In this spec, the path-level "param" is integer but the operation-level "param" is string.
// The generated client should accept a string, proving the operation-level definition won.
func TestParameterPrecedence(t *testing.T) {
	// This compiles only if the generated client uses string (operation-level)
	// rather than int32 (path-level) for the param parameter.
	_, _ = NewGetSimplePrimitiveRequest("http://example.com/", "test-string")
}
