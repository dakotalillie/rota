package main

import (
	"fmt"
	"os"

	"github.com/dakotalillie/rota/internal/api"
	"github.com/dakotalillie/rota/internal/config"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	db, err := sqlite.Open(conf.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	api := api.New(conf, db)

	if err := api.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting API: %v\n", err)
		os.Exit(1)
	}
}
