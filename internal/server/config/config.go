package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseDSN string
	JWTSecret   string
}

func Load() Config {
	cfg := Config{
		HTTPAddr:    getEnv("GOPHKEEPER_HTTP_ADDR", ":8080"),
		DatabaseDSN: getEnv("GOPHKEEPER_DB_DSN", "file:gophkeeper.db?cache=shared&mode=rwc"),
		JWTSecret:   getEnv("GOPHKEEPER_JWT_SECRET", "dev-secret-change"),
	}
	if cfg.JWTSecret == "dev-secret-change" {
		log.Println("WARNING: using development JWT secret; set GOPHKEEPER_JWT_SECRET")
	}
	return cfg
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
