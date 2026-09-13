package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration values.
type Config struct {
	Env         string
	Port        int
	DatabaseURL string
	LogLevel    string
	CORSOrigin  string
	PublicURL   string
}

// Load reads configuration from environment variables with production-ready defaults.
func Load() *Config {
	env := getEnv("ENV", "development")
	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 8080
	}

	return &Config{
		Env:         env,
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", ""),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "*"),
		PublicURL:   getEnv("PUBLIC_URL", "http://localhost:8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
