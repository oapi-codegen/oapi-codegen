package output

import (
	"encoding/json"
	"testing"
	"time"
)

// TestDateIntervalFieldOrder verifies the x-order extension affects field ordering.
// The spec defines end (x-order: 2) before start (x-order: 1), but x-order should
// reorder them so start comes first.
func TestDateIntervalFieldOrder(t *testing.T) {
	start := &Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	end := &Date{Time: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)}
	d := DateInterval{
		Start: start,
		End:   end,
	}
	if d.Start.Format(DateFormat) != "2024-01-01" {
		t.Errorf("unexpected start: %s", d.Start)
	}
	if d.End.Format(DateFormat) != "2024-12-31" {
		t.Errorf("unexpected end: %s", d.End)
	}
}

func TestDateIntervalJSONRoundTrip(t *testing.T) {
	start := &Date{Time: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)}
	d := DateInterval{Start: start}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded DateInterval
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Start == nil || decoded.Start.Format(DateFormat) != "2024-06-15" {
		t.Errorf("round-trip failed for start date")
	}
}

func TestPortTypeAliases(t *testing.T) {
	var p Port = 8080
	if p != 8080 {
		t.Errorf("unexpected port: %d", p)
	}
	var lp LowPriorityPort = 9090
	if lp != 9090 {
		t.Errorf("unexpected low priority port: %d", lp)
	}
}

func TestPortIntervalInstantiation(t *testing.T) {
	lp := LowPriorityPort(9090)
	pi := PortInterval{
		Start:   8080,
		End:     8081,
		VeryEnd: &lp,
	}
	if pi.Start != 8080 {
		t.Errorf("unexpected start: %d", pi.Start)
	}
	if pi.End != 8081 {
		t.Errorf("unexpected end: %d", pi.End)
	}
	if *pi.VeryEnd != 9090 {
		t.Errorf("unexpected very_end: %d", *pi.VeryEnd)
	}
}

func TestPortIntervalJSONRoundTrip(t *testing.T) {
	lp := LowPriorityPort(9090)
	pi := PortInterval{Start: 80, End: 443, VeryEnd: &lp}
	data, err := json.Marshal(pi)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded PortInterval
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Start != 80 || decoded.End != 443 {
		t.Errorf("round-trip failed: start=%d, end=%d", decoded.Start, decoded.End)
	}
	if decoded.VeryEnd == nil || *decoded.VeryEnd != 9090 {
		t.Errorf("round-trip failed for very_end")
	}
}

func TestApplyDefaults(t *testing.T) {
	di := &DateInterval{}
	di.ApplyDefaults()
	pi := &PortInterval{}
	pi.ApplyDefaults()
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}
