package output

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestAlarmInstantiation verifies the Alarm type with UUID fields.
func TestAlarmInstantiation(t *testing.T) {
	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	alarm := Alarm{
		UnderscoreID: &id1,
		ID:           &id2,
	}

	if alarm.UnderscoreID == nil {
		t.Fatal("UnderscoreID should not be nil")
	}
	if *alarm.UnderscoreID != id1 {
		t.Errorf("UnderscoreID = %v, want %v", *alarm.UnderscoreID, id1)
	}
	if alarm.ID == nil {
		t.Fatal("ID should not be nil")
	}
	if *alarm.ID != id2 {
		t.Errorf("ID = %v, want %v", *alarm.ID, id2)
	}
}

// TestAlarmJSONRoundTrip verifies JSON marshal/unmarshal for Alarm with UUID fields.
func TestAlarmJSONRoundTrip(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	original := Alarm{
		UnderscoreID: &id,
		ID:           &id,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Alarm
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.UnderscoreID == nil || *decoded.UnderscoreID != id {
		t.Errorf("UnderscoreID = %v, want %v", decoded.UnderscoreID, id)
	}
	if decoded.ID == nil || *decoded.ID != id {
		t.Errorf("ID = %v, want %v", decoded.ID, id)
	}
}

// TestUUIDTypeAlias verifies that the UUID type alias resolves to uuid.UUID.
func TestUUIDTypeAlias(t *testing.T) {
	var u UUID = uuid.New()
	// Verify it behaves like uuid.UUID
	str := u.String()
	if len(str) != 36 {
		t.Errorf("UUID string length = %d, want 36", len(str))
	}
}

// TestAlarmNilFields verifies Alarm works with nil optional fields.
func TestAlarmNilFields(t *testing.T) {
	alarm := Alarm{}

	data, err := json.Marshal(alarm)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Both fields are optional, so empty JSON object is valid
	var decoded Alarm
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.UnderscoreID != nil {
		t.Error("UnderscoreID should be nil")
	}
	if decoded.ID != nil {
		t.Error("ID should be nil")
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic.
func TestApplyDefaults(t *testing.T) {
	alarm := &Alarm{}
	alarm.ApplyDefaults()
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
