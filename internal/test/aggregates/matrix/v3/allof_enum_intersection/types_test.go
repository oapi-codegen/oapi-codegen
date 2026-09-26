package matrixv3allofenumintersection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The composition allows only the values both enums allow.
func TestEnumIntersection(t *testing.T) {
	assert.True(t, SubjectGreen.Valid())
	assert.True(t, SubjectBlue.Valid())
	assert.False(t, Subject("red").Valid())
	assert.False(t, Subject("violet").Valid())
}
