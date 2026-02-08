package output

import (
	"encoding/json"
	"testing"
	"time"
)

// TestAliasedDateType verifies that date format types work correctly.
// https://github.com/oapi-codegen/oapi-codegen/issues/579
func TestDateType(t *testing.T) {
	// Direct date type should use Date
	date := Date{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}

	data, err := json.Marshal(date)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"2024-01-15"` {
		t.Errorf("Marshal result = %s, want %q", string(data), "2024-01-15")
	}

	var decoded Date
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !decoded.Equal(date.Time) {
		t.Errorf("Unmarshal result = %v, want %v", decoded.Time, date.Time)
	}
}

func TestPetWithDateFields(t *testing.T) {
	// Pet has born_at as *Date (direct format: date)
	date := Date{Time: time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)}
	pet := Pet{
		BornAt: &date,
	}

	if pet.BornAt == nil {
		t.Fatal("BornAt should not be nil")
	}
	if pet.BornAt.String() != "2020-06-15" {
		t.Errorf("BornAt = %q, want %q", pet.BornAt.String(), "2020-06-15")
	}
}

// Born field is now properly typed as *AliasedDate (= Date alias).
func TestPetBornFieldExists(t *testing.T) {
	born := AliasedDate{Time: time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)}
	pet := Pet{
		Born: &born,
	}

	if pet.Born == nil {
		t.Fatal("Born should not be nil")
	}
	if pet.Born.String() != "2020-06-15" {
		t.Errorf("Born = %q, want %q", pet.Born.String(), "2020-06-15")
	}
}
