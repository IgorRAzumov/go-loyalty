package app

import (
	"context"
	"loyalty/internal/config"
	"loyalty/internal/logger"
	"testing"
)

func TestInitDb_ReturnsErrorOnEmptyDatabaseURI(t *testing.T) {
	ctx := context.Background()
	_, err := initDb(ctx, config.Config{DatabaseURI: ""}, logger.NewNopLogger())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
