package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type GooseLogger struct {
	logger *slog.Logger
}

func NewGooseLogger(logger *slog.Logger) *GooseLogger {
	return &GooseLogger{logger: logger.With("component", "goose")}
}

func (g *GooseLogger) Printf(format string, v ...any) {
	g.logger.Info(strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (g *GooseLogger) Fatalf(format string, v ...any) {
	g.logger.Error(strings.TrimSpace(fmt.Sprintf(format, v...)))
	os.Exit(1)
}
