package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      string
	JWTIssuer      string
	TokenTTL       time.Duration
	RedisAddr      string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
}

func Load() Config {
	return Config{
		HTTPAddr:       env("NETCORD_HTTP_ADDR", ":8080"),
		DatabaseURL:    env("NETCORD_DATABASE_URL", "postgres://netcord:CHANGE_ME@127.0.0.1:5432/netcord?sslmode=disable"),
		JWTSecret:      env("NETCORD_JWT_SECRET", "change-me-local-dev-only"),
		JWTIssuer:      env("NETCORD_JWT_ISSUER", "netcord"),
		TokenTTL:       time.Duration(envInt("NETCORD_TOKEN_TTL_HOURS", 168)) * time.Hour,
		RedisAddr:      env("NETCORD_REDIS_ADDR", "127.0.0.1:6379"),
		MinIOEndpoint:  env("NETCORD_MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinIOAccessKey: env("NETCORD_MINIO_ACCESS_KEY", "netcord"),
		MinIOSecretKey: env("NETCORD_MINIO_SECRET_KEY", "CHANGE_ME"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
