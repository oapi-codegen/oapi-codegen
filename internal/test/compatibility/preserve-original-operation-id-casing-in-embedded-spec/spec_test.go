package preserveoriginaloperationidcasinginembeddedspec

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecReturnsOperationIdAsOriginallySpecified(t *testing.T) {
	spec, err := GetSwagger()
	require.NoError(t, err)

	path := spec.Paths.Find("/pet")
	require.NotNil(t, path, "The path /pet could not be found")

	operation := path.GetOperation(http.MethodGet)
	require.NotNil(t, operation, "The GET operation on the path /pet could not be found")

	// this should be the raw operationId from the spec
	assert.Equal(t, "getPet", operation.OperationID)

	operation = path.GetOperation(http.MethodDelete)
	require.NotNil(t, operation, "The DELETE operation on the path /pet could not be found")

	// this should be the raw operationId from the spec
	assert.Equal(t, "this-is-a-kebabAndCamel_SNAKE", operation.OperationID)
}
