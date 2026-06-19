package config

import (
	"log/slog"
	"os"
)

// Config bündelt die per Umgebungsvariablen konfigurierbaren Werte.
type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    slog.Level
}

// Load liest die Configuration aus Umgebungsvariablen mit sinnvollen Defaults.
func Load() Config {
	return Config{
		Port: getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL",
			"postgres://parkhaus:parkhaus@localhost:5432/parkhaus?sslmode=disable&search_path=parkhaus"),
		LogLevel: parseLevel(getEnv("LOG_LEVEL", "info")),
	}
}

// getEnv liest eine Umgebungsvariable oder gibt einen Standardwert zurück.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// parseLevel wandelt einen Log-Level-String in slog.Level um.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
