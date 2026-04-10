package config

import "os"

type Config struct {
	Hostname         string
	DatabasePath     string
	Port             string
	TimeOverrideFile string
}

func Load() (*Config, error) {
	return &Config{
		Hostname:         envOr("HOSTNAME", "http://localhost:5173"),
		DatabasePath:     envOr("DATABASE_PATH", "rota.db"),
		Port:             envOr("PORT", "8080"),
		TimeOverrideFile: os.Getenv("TIME_OVERRIDE_FILE"),
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
