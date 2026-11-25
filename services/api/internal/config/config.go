package config

import "os"

// Config is process configuration. Defaults match .env.example and are local
// development values, not production credentials.
type Config struct {
	EnvName      string
	HTTPAddr     string
	DatabaseURL  string
	RedisAddr    string
	S3Endpoint   string
	S3Region     string
	S3Bucket     string
	S3AccessKey  string
	S3SecretKey  string
	WorkerQueue  string
	OTLPEndpoint string
	CORSOrigins  []string
}

// Load reads environment variables. Empty values fall back to the local ports
// documented in the README.
func Load() Config {
	return Config{
		EnvName:      env("QLUTCH_ENV", "local"),
		HTTPAddr:     env("QLUTCH_HTTP_ADDR", ":4710"),
		DatabaseURL:  env("QLUTCH_DATABASE_URL", "postgres://qlutch:qlutch-local-dev@127.0.0.1:4713/qlutch?sslmode=disable"),
		RedisAddr:    env("QLUTCH_REDIS_ADDR", "127.0.0.1:4714"),
		S3Endpoint:   env("QLUTCH_S3_ENDPOINT", "http://127.0.0.1:4715"),
		S3Region:     env("QLUTCH_S3_REGION", "us-east-1"),
		S3Bucket:     env("QLUTCH_S3_BUCKET", "qlutch-local"),
		S3AccessKey:  env("QLUTCH_S3_ACCESS_KEY", "qlutch"),
		S3SecretKey:  env("QLUTCH_S3_SECRET_KEY", "qlutch-local-dev"),
		WorkerQueue:  env("QLUTCH_WORKER_QUEUE", "qlutch:worker:jobs"),
		OTLPEndpoint: env("QLUTCH_OTLP_ENDPOINT", ""),
		CORSOrigins:  splitCSV(env("QLUTCH_CORS_ORIGINS", "http://127.0.0.1:4712,http://localhost:4712")),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range splitComma(value) {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func splitComma(value string) []string {
	fields := make([]string, 0, 2)
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ',' {
			part := trimSpace(value[start:i])
			fields = append(fields, part)
			start = i + 1
		}
	}
	return fields
}

func trimSpace(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
}
