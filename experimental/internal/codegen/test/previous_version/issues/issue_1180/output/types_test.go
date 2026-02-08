package output

import (
	"strings"
	"testing"
)

// TestEmbeddedSpecAvailable verifies that the embedded OpenAPI spec is generated
// and can be decoded, even when the spec has no component schemas (only paths
// with parameter style handling).
// https://github.com/oapi-codegen/oapi-codegen/issues/1180
func TestEmbeddedSpecAvailable(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}

func TestEmbeddedSpecContainsOpenAPIVersion(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}

	// The spec should contain an openapi version string
	if !strings.Contains(string(data), "openapi") {
		t.Error("Embedded spec should contain 'openapi' field")
	}
}

func TestEmbeddedSpecContainsSimplePrimitivePath(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}

	if !strings.Contains(string(data), "/simplePrimitive/{param}") {
		t.Error("Embedded spec should contain /simplePrimitive/{param} path")
	}
}

func TestEmbeddedSpecCaching(t *testing.T) {
	data1, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	data2, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	if len(data1) != len(data2) {
		t.Errorf("Cached results differ: len(first) = %d, len(second) = %d", len(data1), len(data2))
	}
}
