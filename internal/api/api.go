package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/config"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/dakotalillie/rota/internal/presentation"
)

type API struct {
	conf *config.Config
	db   *sql.DB
}

func New(conf *config.Config, db *sql.DB) *API {
	return &API{conf: conf, db: db}
}

func (a *API) Start() error {
	return a.runServer(a.makeServer())
}

func (a *API) makeServer() *http.Server {
	var (
		mux                = http.NewServeMux()
		rotationRepo       = sqlite.NewRotationRepository(a.db)
		getRotationUseCase = application.NewGetRotationUseCase(rotationRepo)
		getRotationHandler = presentation.NewGetRotationHandler(a.conf.Hostname, getRotationUseCase.Execute)
	)

	mux.HandleFunc("GET /api/rotations/{rotationID}", getRotationHandler.Handle)

	return &http.Server{Addr: ":8080", Handler: mux}
}

func (a *API) runServer(server *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	select {
	case <-quit:
		return server.Shutdown(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
