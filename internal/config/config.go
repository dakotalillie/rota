package config

import "os"

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

type Config struct {
	Hostname         string
	DatabasePath     string
	Port             string
	TimeOverrideFile string
	LogLevel         string
	LogFormat        string
}

func Load() (*Config, error) {
	port := envOr("PORT", "8080")

	return &Config{
		Hostname:         envOr("HOSTNAME", "http://localhost:"+port),
		DatabasePath:     envOr("DATABASE_PATH", "rota.db"),
		Port:             port,
		TimeOverrideFile: os.Getenv("TIME_OVERRIDE_FILE"),
		LogLevel:         envOr("LOG_LEVEL", LogLevelInfo),
		LogFormat:        envOr("LOG_FORMAT", LogFormatText),
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
