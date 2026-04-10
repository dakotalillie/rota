package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/clock"
	"github.com/dakotalillie/rota/internal/config"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/dakotalillie/rota/internal/ui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}

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

	var clk domain.Clock
	if conf.TimeOverrideFile != "" {
		clk = clock.NewFS(conf.TimeOverrideFile)
	} else {
		clk = clock.New()
	}

	var (
		transactor            = sqlite.NewTransactor(db)
		rotationRepo          = sqlite.NewRotationRepository(db)
		userRepo              = sqlite.NewUserRepository(db)
		memberRepo            = sqlite.NewMemberRepository(db)
		overrideRepo          = sqlite.NewOverrideRepository(db)
		createRotationUseCase = application.NewCreateRotationUseCase(transactor, rotationRepo)
		getRotationUseCase    = application.NewGetRotationUseCase(rotationRepo, overrideRepo, clk)
		listRotationsUseCase  = application.NewListRotationsUseCase(rotationRepo, overrideRepo, clk)
		createMemberUseCase   = application.NewCreateMemberUseCase(transactor, rotationRepo, userRepo, memberRepo)
		reorderMembersUseCase = application.NewReorderMembersUseCase(transactor, rotationRepo, memberRepo)
		deleteMemberUseCase   = application.NewDeleteMemberUseCase(transactor, rotationRepo, memberRepo, overrideRepo, userRepo)
		getScheduleUseCase    = application.NewGetScheduleUseCase(rotationRepo, overrideRepo)
		createOverrideUseCase = application.NewCreateOverrideUseCase(transactor, rotationRepo, overrideRepo)
		deleteOverrideUseCase = application.NewDeleteOverrideUseCase(transactor, rotationRepo, overrideRepo)
		worker                = application.NewAdvanceRotationWorker(rotationRepo, memberRepo, clk, 5*time.Second, slog.Default().With("component", "advance_rotation_worker"))
		createRotationHandler = httpapi.NewCreateRotationHandler(conf.Hostname, createRotationUseCase.Execute)
		getRotationHandler    = httpapi.NewGetRotationHandler(conf.Hostname, getRotationUseCase.Execute, clk)
		listRotationsHandler  = httpapi.NewListRotationsHandler(conf.Hostname, listRotationsUseCase.Execute, clk)
		createMemberHandler   = httpapi.NewCreateMemberHandler(conf.Hostname, createMemberUseCase.Execute, clk)
		reorderMembersHandler = httpapi.NewReorderMembersHandler(conf.Hostname, reorderMembersUseCase.Execute)
		deleteMemberHandler   = httpapi.NewDeleteMemberHandler(conf.Hostname, deleteMemberUseCase.Execute, clk)
		getScheduleHandler    = httpapi.NewGetScheduleHandler(conf.Hostname, getScheduleUseCase.Execute, clk)
		createOverrideHandler = httpapi.NewCreateOverrideHandler(conf.Hostname, createOverrideUseCase.Execute)
		deleteOverrideHandler = httpapi.NewDeleteOverrideHandler(conf.Hostname, deleteOverrideUseCase.Execute)
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/rotations", createRotationHandler.Handle)
	mux.HandleFunc("GET /api/rotations", listRotationsHandler.Handle)
	mux.HandleFunc("GET /api/rotations/{rotationID}", getRotationHandler.Handle)
	mux.HandleFunc("POST /api/rotations/{rotationID}/members", createMemberHandler.Handle)
	mux.HandleFunc("PUT /api/rotations/{rotationID}/members", reorderMembersHandler.Handle)
	mux.HandleFunc("DELETE /api/rotations/{rotationID}/members/{memberID}", deleteMemberHandler.Handle)
	mux.HandleFunc("GET /api/rotations/{rotationID}/schedule", getScheduleHandler.Handle)
	mux.HandleFunc("POST /api/rotations/{rotationID}/overrides", createOverrideHandler.Handle)
	mux.HandleFunc("DELETE /api/rotations/{rotationID}/overrides/{overrideID}", deleteOverrideHandler.Handle)

	uiFS, err := fs.Sub(ui.FS, "dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading embedded UI: %v\n", err)
		os.Exit(1)
	}
	fileServer := http.FileServerFS(uiFS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(uiFS, p); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{Addr: ":" + conf.Port, Handler: mux}

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
