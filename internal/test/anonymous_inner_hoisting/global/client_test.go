package global

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHoistedTypesExist verifies that with output-options.generate-types-for-
// anonymous-schemas enabled, the inline schemas in spec.yaml become named
// Go types we can reference directly. The spec is the canonical issue #1139
// shape: a response body using `allOf` to merge a $ref with sibling
// `properties:` containing an inline `data` object.
func TestHoistedTypesExist(t *testing.T) {
	// Both the response root and the nested inline `data` schema should be
	// emitted as named types — assigning a typed zero value would not
	// compile if either were still anonymous structs.
	var responseBody GetRolesId200JSONResponseBody
	var dataField GetRolesId200JSONResponseBody_Data

	// Field-level type identity: GetRolesId200JSONResponseBody.Data must be
	// of the hoisted GetRolesId200JSONResponseBody_Data type. This
	// assignment fails to compile if Data is still an anonymous struct.
	responseBody.Data = dataField
	_ = responseBody
}

func TestHoistedTypesRoundTrip(t *testing.T) {
	body := GetRolesId200JSONResponseBody{
		Data: GetRolesId200JSONResponseBody_Data{
			Role: Role{Id: 7, Name: "admin"},
		},
		Ok: true,
	}

	encoded, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded GetRolesId200JSONResponseBody
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, body, decoded)
}
