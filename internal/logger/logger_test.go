package logger

import "testing"

func TestInitLogger_Default(t *testing.T) {
	if err := InitLogger(""); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	if err := InitLogger("nope-nope"); err == nil {
		t.Fatalf("expected error")
	}
}
