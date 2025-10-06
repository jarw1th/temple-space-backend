package config

import (
	"os"
)

type Config struct {
	HTTPAddr     string
	GRPCAddr     string
	PostgresURL  string
	RedisURL     string
	KafkaBrokers string
	AuthGRPCAddr string
}

func FromEnv() *Config {
	return &Config{
		HTTPAddr:     getenv("BOOKING_HTTP_ADDR", ":8081"),
		GRPCAddr:     getenv("BOOKING_GRPC_ADDR", ":9091"),
		PostgresURL:  getenv("BOOKING_POSTGRES_URL", "postgres://localhost:5432/templespace?sslmode=disable"),
		RedisURL:     getenv("BOOKING_REDIS_URL", "redis://localhost:6379"),
		KafkaBrokers: getenv("BOOKING_KAFKA_BROKERS", "localhost:9092"),
		AuthGRPCAddr: getenv("AUTH_GRPC_ADDR", ":9090"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
