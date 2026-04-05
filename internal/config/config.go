package config

type Config struct {
	Hostname     string
	DatabasePath string
}

func Load() (*Config, error) {
	return &Config{
		Hostname:     "http://localhost:5173",
		DatabasePath: "rota.db",
	}, nil
}
