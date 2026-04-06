package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		transactor            = sqlite.NewTransactor(db)
		rotationRepo          = sqlite.NewRotationRepository(db)
		userRepo              = sqlite.NewUserRepository(db)
		memberRepo            = sqlite.NewMemberRepository(db)
		overrideRepo          = sqlite.NewOverrideRepository(db)
		createRotationUseCase = application.NewCreateRotationUseCase(transactor, rotationRepo)
		getRotationUseCase    = application.NewGetRotationUseCase(rotationRepo, overrideRepo)
		listRotationsUseCase  = application.NewListRotationsUseCase(rotationRepo)
		createMemberUseCase   = application.NewCreateMemberUseCase(transactor, rotationRepo, userRepo, memberRepo)
		getScheduleUseCase    = application.NewGetScheduleUseCase(rotationRepo)
		createOverrideUseCase = application.NewCreateOverrideUseCase(transactor, rotationRepo, overrideRepo)
		worker                = application.NewAdvanceRotationWorker(rotationRepo, memberRepo, 5*time.Second, slog.Default().With("component", "advance_rotation_worker"))
		createRotationHandler = httpapi.NewCreateRotationHandler(conf.Hostname, createRotationUseCase.Execute)
		getRotationHandler    = httpapi.NewGetRotationHandler(conf.Hostname, getRotationUseCase.Execute)
		listRotationsHandler  = httpapi.NewListRotationsHandler(conf.Hostname, listRotationsUseCase.Execute)
		createMemberHandler   = httpapi.NewCreateMemberHandler(conf.Hostname, createMemberUseCase.Execute)
		getScheduleHandler    = httpapi.NewGetScheduleHandler(conf.Hostname, getScheduleUseCase.Execute)
		createOverrideHandler = httpapi.NewCreateOverrideHandler(conf.Hostname, createOverrideUseCase.Execute)
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/rotations", createRotationHandler.Handle)
	mux.HandleFunc("GET /api/rotations", listRotationsHandler.Handle)
	mux.HandleFunc("GET /api/rotations/{rotationID}", getRotationHandler.Handle)
	mux.HandleFunc("POST /api/rotations/{rotationID}/members", createMemberHandler.Handle)
	mux.HandleFunc("GET /api/rotations/{rotationID}/schedule", getScheduleHandler.Handle)
	mux.HandleFunc("POST /api/rotations/{rotationID}/overrides", createOverrideHandler.Handle)

	server := &http.Server{Addr: ":8080", Handler: mux}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	go worker.Run(workerCtx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	select {
	case <-quit:
		cancelWorker()
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
