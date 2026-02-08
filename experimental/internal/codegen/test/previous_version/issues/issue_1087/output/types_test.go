package output

import (
	"encoding/json"
	"testing"
)

// TestThingConstruction verifies that Thing and ThingList types can be
// constructed and their fields accessed correctly.
// https://github.com/oapi-codegen/oapi-codegen/issues/1087
func TestThingConstruction(t *testing.T) {
	thing := Thing{Name: "widget"}
	if thing.Name != "widget" {
		t.Errorf("Thing.Name = %q, want %q", thing.Name, "widget")
	}

	list := ThingList{
		Keys: []Thing{
			{Name: "a"},
			{Name: "b"},
		},
	}
	if len(list.Keys) != 2 {
		t.Errorf("ThingList.Keys length = %d, want 2", len(list.Keys))
	}
	if list.Keys[0].Name != "a" {
		t.Errorf("ThingList.Keys[0].Name = %q, want %q", list.Keys[0].Name, "a")
	}

	// ThingListKeys is a type alias for []Thing
	var keys ThingListKeys = []Thing{{Name: "c"}}
	if len(keys) != 1 {
		t.Errorf("ThingListKeys length = %d, want 1", len(keys))
	}
}

func TestThingJSONRoundTrip(t *testing.T) {
	original := ThingList{
		Keys: []Thing{
			{Name: "first"},
			{Name: "second"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ThingList
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Keys) != 2 {
		t.Fatalf("decoded Keys length = %d, want 2", len(decoded.Keys))
	}
	if decoded.Keys[0].Name != "first" {
		t.Errorf("decoded Keys[0].Name = %q, want %q", decoded.Keys[0].Name, "first")
	}
	if decoded.Keys[1].Name != "second" {
		t.Errorf("decoded Keys[1].Name = %q, want %q", decoded.Keys[1].Name, "second")
	}
}

func TestApplyDefaults(t *testing.T) {
	thing := &Thing{Name: "test"}
	thing.ApplyDefaults() // should not panic

	list := &ThingList{Keys: []Thing{{Name: "x"}}}
	list.ApplyDefaults() // should not panic
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("GetOpenAPISpecJSON() returned empty data")
	}
}
