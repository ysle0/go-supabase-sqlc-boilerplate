package main

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/go-chi/chi/v5"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

var (
	port               = shared.EnvString("PORT", ":8080")
	httpRequestTimeout = shared.EnvDuration("HTTP_REQUEST_TIMEOUT", 30*time.Second)
	shutdownTimeout    = shared.EnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	logger             = slog.New(slogcolor.NewHandler(os.Stdout, slogcolor.DefaultOptions))
)

func main() {
	logger.Info("Starting API server", slog.String("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := chi.NewRouter()
	s := NewServer(ctx, r, logger, httpRequestTimeout)

	// Mount pprof for profiling
	r.Mount("/debug", http.DefaultServeMux)

	closer, err := s.Start(ctx, port)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	if err := shared.WaitForGracefulExit(ctx, shutdownTimeout, closer); err != nil {
		logger.Error("graceful exit error", "error", err)
		os.Exit(1)
	}

	logger.Info("API server stopped gracefully")
}
