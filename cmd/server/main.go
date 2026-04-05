package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/config"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
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
	defer db.Close() //nolint:errcheck

	var (
		rotationRepo       = sqlite.NewRotationRepository(db)
		getRotationUseCase = application.NewGetRotationUseCase(rotationRepo)
		getRotationHandler = httpapi.NewGetRotationHandler(conf.Hostname, getRotationUseCase.Execute)
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/rotations/{rotationID}", getRotationHandler.Handle)

	server := &http.Server{Addr: ":8080", Handler: mux}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	select {
	case <-quit:
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down server: %v\n", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "Error running server: %v\n", err)
			os.Exit(1)
		}
	}
}
