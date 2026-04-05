package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/config"
	"github.com/dakotalillie/rota/internal/presentation"
)

type API struct {
	conf *config.Config
}

func (a *API) Start() error {
	return a.runServer(a.makeServer())
}

func (a *API) makeServer() *http.Server {
	var (
		mux                = http.NewServeMux()
		getRotationUseCase = application.NewGetRotationUseCase()
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

func New(conf *config.Config) *API {
	return &API{conf: conf}
}
