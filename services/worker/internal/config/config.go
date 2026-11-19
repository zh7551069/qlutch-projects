package config

import (
	"os"
	"time"
)

// Config is the worker process configuration. Defaults match .env.example.
type Config struct {
	RedisAddr string
	QueueKey  string
	Heartbeat time.Duration
}

// Load reads environment variables and falls back to the local development values.
func Load() Config {
	return Config{
		RedisAddr: env("QLUTCH_REDIS_ADDR", "127.0.0.1:4714"),
		QueueKey:  env("QLUTCH_WORKER_QUEUE", "qlutch:worker:jobs"),
		Heartbeat: durationEnv("QLUTCH_WORKER_HEARTBEAT", 30*time.Second),
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
