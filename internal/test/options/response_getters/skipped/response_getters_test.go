package optionsresponsegettersskipped

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// outputoptions/response-body-getters/skipped: when skip-response-body-getters is true,
// the response type must NOT expose GetBody() or any of the typed getters the default
// template emits — but the response struct itself (and its Body field) is still generated.
func TestResponseBodyGettersSkipped(t *testing.T) {
	respType := reflect.TypeOf(GetThingResponse{})

	for _, name := range []string{"GetBody", "GetJSON200", "GetJSONDefault"} {
		_, ok := respType.MethodByName(name)
		assert.Falsef(t, ok, "method %s should not be generated when skip-response-body-getters is true", name)
	}

	// Sanity check that the response type itself is still generated — only the
	// getters are skipped, not the struct.
	_, ok := respType.FieldByName("Body")
	assert.True(t, ok, "Body field should still be present on the response struct")
}
