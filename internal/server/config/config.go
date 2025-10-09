package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr              string
	DatabaseDSN           string
	JWTSecret             string
	MaxRequestBytes       int64
	MaxRecordPayloadBytes int64
}

func Load() Config {
	cfg := Config{
		HTTPAddr:              getEnv("GOPHKEEPER_HTTP_ADDR", ":8080"),
		DatabaseDSN:           getEnv("GOPHKEEPER_DB_DSN", "file:gophkeeper.db?cache=shared&mode=rwc"),
		JWTSecret:             getEnv("GOPHKEEPER_JWT_SECRET", "dev-secret-change"),
		MaxRequestBytes:       getEnvInt64("GOPHKEEPER_MAX_REQUEST_BYTES", 1<<20),
		MaxRecordPayloadBytes: getEnvInt64("GOPHKEEPER_MAX_RECORD_PAYLOAD_BYTES", 1<<20),
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

func getEnvInt64(key string, def int64) int64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
		log.Printf("WARNING: invalid %s='%s', using default %d", key, v, def)
	}
	return def
}
