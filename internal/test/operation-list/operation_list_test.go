package operationlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrictOperationIDs(t *testing.T) {
	expected := []string{"ListUsers", "CreateUser", "GetUser", "DeleteUser"}
	assert.ElementsMatch(t, expected, StrictOperationIDs)
}
