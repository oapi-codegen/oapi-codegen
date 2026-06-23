package optionscompatibility

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compatibility/preserve-original-operation-id-casing-in-embedded-spec:
// The embedded spec blob must preserve the raw operationId values from the
// OpenAPI spec — both simple camelCase and mixed kebab+camel+SNAKE forms.
func TestSpecReturnsOperationIdAsOriginallySpecified(t *testing.T) {
	spec, err := GetSpec()
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
