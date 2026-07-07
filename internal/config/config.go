package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTTTLHours int
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://blog:blog@localhost:5432/blog?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "dev-secret-change-me-please-32bytes!"),
		JWTTTLHours: envInt("JWT_TTL_HOURS", 72),
	}
	if len(c.JWTSecret) < 16 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 16 bytes")
	}
	return c, nil
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v, ok := os.LookupEnv(k); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}