package logger

import "testing"

func TestInitLogger_Default(t *testing.T) {
	t.Setenv("LOG_LEVEL", "")
	if err := InitLogger(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestInitLogger_InvalidLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "nope-nope")
	if err := InitLogger(); err == nil {
		t.Fatalf("expected error")
	}
}
