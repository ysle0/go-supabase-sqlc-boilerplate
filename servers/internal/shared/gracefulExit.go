package shared

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatusOllah/slogcolor"
)

var logger = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
	Level:       slog.LevelInfo,
	TimeFormat:  time.DateTime,
	SrcFileMode: slogcolor.ShortFile,
}))

func WaitForGracefulExit(ctx context.Context, shutdownTimeout time.Duration, closer Closer) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var sig os.Signal
	select {
	case sig = <-sigCh:
		logger.Info("received shutdown signal", "signal", sig.String())
	case <-ctx.Done():
		logger.Info("context done")
	}

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("shutting down server... interrupted by signal: ", "signal", sig)
	return closer.Close(shutdownCtx)
}

func WaitForGracefulExitExt(ctx context.Context, shutdownTimeout time.Duration, errCh <-chan error, closer func() error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var sig os.Signal
	select {
	case sig = <-sigCh:
		logger.Info("received shutdown signal", "signal", sig.String())
	case <-ctx.Done():
		logger.Info("context done")
	case err := <-errCh:
		logger.Error("error occurred", "error", err)
	}

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("shutting down server... interrupted by signal: ", "signal", sig)
	return closer()
}
