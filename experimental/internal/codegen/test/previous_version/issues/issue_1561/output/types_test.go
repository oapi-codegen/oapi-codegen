package output

import (
	"encoding/json"
	"testing"
)

// TestPongInstantiation verifies the Pong type has a required Ping field.
func TestPongInstantiation(t *testing.T) {
	p := Pong{Ping: "pong"}
	if p.Ping != "pong" {
		t.Errorf("Ping = %q, want %q", p.Ping, "pong")
	}
}

// TestPongJSONRoundTrip verifies JSON round-trip for Pong.
func TestPongJSONRoundTrip(t *testing.T) {
	original := Pong{Ping: "test"}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Pong
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Ping != "test" {
		t.Errorf("Ping = %q, want %q", decoded.Ping, "test")
	}
}

// TestResponseBodyInstantiation verifies the ResponseBody type with various field types.
func TestResponseBodyInstantiation(t *testing.T) {
	rb := ResponseBody{
		RequiredSlice: []Pong{{Ping: "a"}},
		ASlice:        []Pong{{Ping: "b"}},
		UnknownObject: map[string]any{"key": "val"},
		AdditionalProps: map[string]any{
			"extra": float64(42),
		},
		ASliceWithAdditionalProps: []ResponseBodyASliceWithAdditionalPropsItem{"item"},
		Bytes:                     []byte("binary"),
		BytesWithOverride:         []byte("override"),
	}

	if len(rb.RequiredSlice) != 1 || rb.RequiredSlice[0].Ping != "a" {
		t.Errorf("RequiredSlice unexpected: %v", rb.RequiredSlice)
	}
	if len(rb.ASlice) != 1 || rb.ASlice[0].Ping != "b" {
		t.Errorf("ASlice unexpected: %v", rb.ASlice)
	}
	if rb.UnknownObject["key"] != "val" {
		t.Errorf("UnknownObject[key] = %v, want %q", rb.UnknownObject["key"], "val")
	}
	if rb.AdditionalProps["extra"] != float64(42) {
		t.Errorf("AdditionalProps[extra] = %v, want 42", rb.AdditionalProps["extra"])
	}
	if string(rb.Bytes) != "binary" {
		t.Errorf("Bytes = %q, want %q", rb.Bytes, "binary")
	}
}

// TestResponseBodyJSONRoundTrip verifies JSON round-trip for ResponseBody.
func TestResponseBodyJSONRoundTrip(t *testing.T) {
	original := ResponseBody{
		RequiredSlice: []Pong{{Ping: "hello"}},
		Bytes:         []byte{0x01, 0x02},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ResponseBody
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.RequiredSlice) != 1 || decoded.RequiredSlice[0].Ping != "hello" {
		t.Errorf("RequiredSlice = %v, want [{Ping: hello}]", decoded.RequiredSlice)
	}
}

// TestTypeAliases verifies that type aliases resolve correctly.
func TestTypeAliases(t *testing.T) {
	// ResponseBodyRequiredSlice is an alias for []Pong
	var rs ResponseBodyRequiredSlice = []Pong{{Ping: "alias"}}
	if len(rs) != 1 {
		t.Errorf("ResponseBodyRequiredSlice len = %d, want 1", len(rs))
	}

	// ResponseBodyASlice is an alias for []Pong
	var as ResponseBodyASlice = []Pong{{Ping: "slice"}}
	if len(as) != 1 {
		t.Errorf("ResponseBodyASlice len = %d, want 1", len(as))
	}

	// ResponseBodyAdditionalProps is an alias for map[string]any
	var ap ResponseBodyAdditionalProps = map[string]any{"k": "v"}
	if ap["k"] != "v" {
		t.Errorf("ResponseBodyAdditionalProps[k] = %v, want %q", ap["k"], "v")
	}

	// ResponseBodyASliceWithAdditionalPropsItem is an alias for any
	var item ResponseBodyASliceWithAdditionalPropsItem = "anything"
	if item != "anything" {
		t.Errorf("item = %v, want %q", item, "anything")
	}
}

// TestResponseBodyUnknownObjectEmpty verifies the empty struct type.
func TestResponseBodyUnknownObjectEmpty(t *testing.T) {
	obj := ResponseBodyUnknownObject{}
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("Marshal result = %s, want {}", string(data))
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic on all types.
func TestApplyDefaults(t *testing.T) {
	(&ResponseBody{}).ApplyDefaults()
	(&ResponseBodyUnknownObject{}).ApplyDefaults()
	(&Pong{}).ApplyDefaults()
}

// TestGetOpenAPISpecJSON verifies the embedded spec can be decoded.
func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}
