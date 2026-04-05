package main

import (
	"fmt"
	"os"

	"github.com/dakotalillie/rota/internal/api"
	"github.com/dakotalillie/rota/internal/config"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	api := api.New(conf)

	if err := api.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting API: %v\n", err)
		os.Exit(1)
	}
}
