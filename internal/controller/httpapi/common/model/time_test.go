package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRFC3339Time_MarshalJSON(t *testing.T) {
	tm := time.Date(2020, 12, 10, 15, 15, 45, 0, time.UTC)
	rfc := RFC3339Time{Time: tm}

	data, err := json.Marshal(rfc)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	expected := `"2020-12-10T15:15:45Z"`
	if string(data) != expected {
		t.Errorf("MarshalJSON() = %v, want %v", string(data), expected)
	}
}

func TestRFC3339Time_UnmarshalJSON(t *testing.T) {
	data := []byte(`"2020-12-10T15:15:45Z"`)
	var rfc RFC3339Time

	err := json.Unmarshal(data, &rfc)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	expected := time.Date(2020, 12, 10, 15, 15, 45, 0, time.UTC)
	if !rfc.Time.Equal(expected) {
		t.Errorf("UnmarshalJSON() time = %v, want %v", rfc.Time, expected)
	}
}

func TestRFC3339Time_InvalidJSON(t *testing.T) {
	data := []byte(`"invalid-date"`)
	var rfc RFC3339Time

	err := json.Unmarshal(data, &rfc)
	if err == nil {
		t.Error("UnmarshalJSON() expected error for invalid date")
	}
}
