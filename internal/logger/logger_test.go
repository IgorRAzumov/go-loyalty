package logger

import "testing"

func TestNewLogger_Default(t *testing.T) {
	log, err := NewLogger("")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	_, err := NewLogger("nope-nope")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewNopLogger(t *testing.T) {
	log := NewNopLogger()
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	log.Info().String("key", "val").Message("nop")
}
