package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	HTTPPort       string
	GRPCPort       string
	PostgresURL    string
	RedisURL       string
	KafkaBrokers   []string
	JWTIssuer      string
	JWTPrivateKey  string // PEM (for dev); in prod use KMS or file path
	JWTPublicKey   string // PEM
	MagicTTLMin    int
	AccessTTLMin   int
	RefreshTTLHour int
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env %s", key)
	}
	return v
}

func Load() Config {
	return Config{
		HTTPPort:       getenv("PORT", "8080"),
		GRPCPort:       getenv("GRPC_PORT", "9090"),
		PostgresURL:    getenv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"),
		RedisURL:       getenv("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers:   strings.Split(getenv("KAFKA_BROKERS", "localhost:9092"), ","),
		JWTIssuer:      getenv("JWT_ISSUER", "templespace"),
		JWTPrivateKey:  getenv("JWT_PRIVATE_KEY_PEM", ""),
		JWTPublicKey:   getenv("JWT_PUBLIC_KEY_PEM", ""),
		MagicTTLMin:    10,
		AccessTTLMin:   60,
		RefreshTTLHour: 720,
	}
}
