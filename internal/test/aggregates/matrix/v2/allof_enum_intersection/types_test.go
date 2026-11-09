package matrixv2allofenumintersection

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// v2 merges two enums into their union.
func TestEnumUnion(t *testing.T) {
	for _, v := range []Subject{SubjectRed, SubjectGreen, SubjectBlue, SubjectViolet} {
		assert.True(t, v.Valid(), v)
	}
}
