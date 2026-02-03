package app

import (
	"context"
	"loyalty/internal/config"
	"testing"
)

func TestInitDb_ReturnsErrorOnEmptyDatabaseURI(t *testing.T) {
	ctx := context.Background()
	_, err := initDb(ctx, config.Config{DatabaseURI: ""})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
