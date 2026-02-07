package output

import (
	"testing"
)

// TestAllOfWithFormatCompiles verifies that using allOf to add format
// specifications doesn't cause generation errors.
// https://github.com/oapi-codegen/oapi-codegen/issues/775
//
// Note: The current implementation generates empty struct types for these
// properties instead of the ideal Go types (uuid.UUID for format:uuid,
// time.Time for format:date). This is a known limitation.
func TestAllOfWithFormatCompiles(t *testing.T) {
	// The fact that this compiles proves the original issue is fixed
	// (generation no longer errors on allOf + format)
	obj := TestObject{
		UUIDProperty: &TestObjectUUIDProperty{},
		DateProperty: &TestObjectDateProperty{},
	}

	// Access the fields to ensure they exist
	_ = obj.UUIDProperty
	_ = obj.DateProperty
}

// TestIdealBehavior documents the expected ideal behavior.
// Currently this would require changes to handle format-only allOf schemas.
func TestIdealBehavior(t *testing.T) {
	t.Skip("TODO: allOf with format-only schemas should produce proper Go types (uuid.UUID, time.Time)")

	// Ideal behavior would be:
	// type TestObject struct {
	//     UUIDProperty *uuid.UUID `json:"uuidProperty,omitempty"`
	//     DateProperty *time.Time `json:"dateProperty,omitempty"`
	// }
}
