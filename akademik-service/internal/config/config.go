package config

import "os"

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	Port       string
}

func Load() Config {
	return Config{
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBUser:     env("DB_USER", "postgres"),
		DBPassword: env("DB_PASSWORD", "postgres"),
		DBName:     env("DB_NAME", "siakad"),
		DBSSLMode:  env("DB_SSLMODE", "disable"),
		Port:       env("PORT", "8080"),
	}
}

func (c Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + c.DBPort +
		"/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
