package main

import (
	"fmt"

	"github.com/dakotalillie/rota/internal/api"
)

func main() {
	api := api.New()

	if err := api.Start(); err != nil {
		fmt.Println("Error starting API:", err)
	}
}
