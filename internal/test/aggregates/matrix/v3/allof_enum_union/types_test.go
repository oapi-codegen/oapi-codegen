package matrixv3allofenumunion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// x-oapi-codegen-enum-merge: union allows the values of either enum.
func TestEnumUnion(t *testing.T) {
	for _, v := range []Subject{SubjectActive, SubjectInactive, SubjectArchived} {
		assert.True(t, v.Valid(), v)
	}
}
