package logger

import (
	"github.com/medods/auth-service/internal/config"
	"io"
	"log/slog"
	"os"
)

func SetupLogger(cfg *config.Config) error {
	const op = "internal.logger.SetupLogger"

	if err := os.MkdirAll(cfg.Log.GetLogDir(), 0755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(
		cfg.Log.FilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	var logger *slog.Logger

	var handler slog.Handler
	if cfg.Env == "development" {
		handler = slog.NewJSONHandler(io.MultiWriter(os.Stdout, logFile), &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	logger = slog.New(handler)

	slog.SetDefault(logger)
	return nil
}
