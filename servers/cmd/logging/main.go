package main

import (
	"context"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/logging"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

var (
	port            = shared.EnvString("PORT", ":8082")
	shutdownTimeout = shared.EnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	logger          = slog.New(slogcolor.NewHandler(os.Stdout, slogcolor.DefaultOptions))
)

func main() {
	logger.Info("Starting Log Server", slog.String("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := logging.NewServer(logger)

	closer, err := s.Start(ctx, port)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	if err := shared.WaitForGracefulExit(ctx, shutdownTimeout, closer); err != nil {
		logger.Error("graceful exit error", "error", err)
		os.Exit(1)
	}

	logger.Info("Logging server stopped gracefully")
}
