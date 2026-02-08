package child

import (
	"testing"
)

// TestGetOpenAPISpecJSON verifies the embedded spec can be decoded.
// https://github.com/oapi-codegen/oapi-codegen/issues/1093
func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("GetOpenAPISpecJSON() returned empty data")
	}
}
