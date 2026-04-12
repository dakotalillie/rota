package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger(t *testing.T) {
	t.Run("logs request at debug level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})

		handler := httpapi.RequestLogger(logger, inner)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/rotations", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, "request completed", entry["msg"])
		assert.Equal(t, "POST", entry["method"])
		assert.Equal(t, "/api/rotations", entry["path"])
		assert.Equal(t, float64(http.StatusCreated), entry["status"])
		assert.Contains(t, entry, "duration")
	})

	t.Run("captures default 200 status", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})

		handler := httpapi.RequestLogger(logger, inner)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/rotations", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, float64(http.StatusOK), entry["status"])
	})

	t.Run("does not log at info level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

		inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := httpapi.RequestLogger(logger, inner)
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Empty(t, buf.String())
	})
}
