package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstantsValues(t *testing.T) {
	assert.Equal(t, ">", string(ExpressionOperatorLargerThan))
	assert.Equal(t, ">=", string(ExpressionOperatorLargerThanOrEqualTo))
	assert.Equal(t, "<", string(ExpressionOperatorLowerThan))
	assert.Equal(t, "<=", string(ExpressionOperatorLowerThanOrEqualTo))
	assert.Equal(t, "=", string(ExpressionOperatorEqual))
	assert.Equal(t, "!=", string(ExpressionOperatorNotEqual))
}
