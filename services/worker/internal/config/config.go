package config

import (
	"os"
	"time"
)

// Config is the worker process configuration. Defaults match .env.example.
// Values are local development defaults, not production credentials.
type Config struct {
	RedisAddr   string
	QueueKey    string
	Heartbeat   time.Duration
	DatabaseURL string
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
}

// Load reads environment variables and falls back to the local development values.
func Load() Config {
	return Config{
		RedisAddr:   env("QLUTCH_REDIS_ADDR", "127.0.0.1:4714"),
		QueueKey:    env("QLUTCH_WORKER_QUEUE", "qlutch:worker:jobs"),
		Heartbeat:   durationEnv("QLUTCH_WORKER_HEARTBEAT", 30*time.Second),
		DatabaseURL: env("QLUTCH_DATABASE_URL", "postgres://qlutch:qlutch-local-dev@127.0.0.1:4713/qlutch?sslmode=disable"),
		S3Endpoint:  env("QLUTCH_S3_ENDPOINT", "http://127.0.0.1:4715"),
		S3Region:    env("QLUTCH_S3_REGION", "us-east-1"),
		S3Bucket:    env("QLUTCH_S3_BUCKET", "qlutch-local"),
		S3AccessKey: env("QLUTCH_S3_ACCESS_KEY", "qlutch"),
		S3SecretKey: env("QLUTCH_S3_SECRET_KEY", "qlutch-local-dev"),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
