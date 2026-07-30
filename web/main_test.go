package main

import (
	"testing"

	infraShared "shopigo/internal/infra/shared"
)

func TestRuntimePostgresConfigUsesAppDatabase(t *testing.T) {
	t.Setenv("POSTGRES_DB", "shopigo")
	t.Setenv("POSTGRES_DB_TEST", "shopigo_test")

	cfg := infraShared.NewPostgresConfig()
	if cfg.DB != "shopigo" {
		t.Fatalf("expected runtime db shopigo, got %q", cfg.DB)
	}
}
