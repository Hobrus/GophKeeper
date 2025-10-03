package config

import (
	"os"
	"testing"
)

func TestLoadDefaultsAndEnv(t *testing.T) {
	// defaults
	os.Unsetenv("GOPHKEEPER_HTTP_ADDR")
	os.Unsetenv("GOPHKEEPER_DB_DSN")
	os.Unsetenv("GOPHKEEPER_JWT_SECRET")
	cfg := Load()
	if cfg.HTTPAddr == "" || cfg.DatabaseDSN == "" || cfg.JWTSecret == "" {
		t.Fatalf("empty config fields")
	}

	// env override
	os.Setenv("GOPHKEEPER_HTTP_ADDR", ":9999")
	os.Setenv("GOPHKEEPER_DB_DSN", "file::memory:")
	os.Setenv("GOPHKEEPER_JWT_SECRET", "secret")
	cfg = Load()
	if cfg.HTTPAddr != ":9999" || cfg.DatabaseDSN != "file::memory:" || cfg.JWTSecret != "secret" {
		t.Fatalf("env not applied: %+v", cfg)
	}
}
