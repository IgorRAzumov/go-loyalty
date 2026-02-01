package authctx

import (
	"context"
	"testing"
)

func TestUserID_Empty(t *testing.T) {
	id, ok := UserID(context.Background())
	if ok || id != 0 {
		t.Fatalf("expected (0,false), got (%d,%v)", id, ok)
	}
}

func TestWithUserID_SetsAndUserID_Reads(t *testing.T) {
	ctx := WithUserID(context.Background(), 42)
	id, ok := UserID(ctx)
	if !ok || id != 42 {
		t.Fatalf("expected (42,true), got (%d,%v)", id, ok)
	}
}
