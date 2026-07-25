package config

import "os"

type Config struct {
	Port            string
	AkademikBaseURL string
}

func Load() Config {
	return Config{
		Port:            env("PORT", "8080"),
		AkademikBaseURL: env("AKADEMIK_BASE_URL", "http://akademik-service:8080"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
