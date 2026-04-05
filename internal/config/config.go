package config

type Config struct {
	Hostname string
}

func Load() (*Config, error) {
	return &Config{
		Hostname: "http://localhost:5173",
	}, nil
}
