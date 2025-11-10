package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/go-chi/chi/v5"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

var (
	port            = shared.EnvString("PORT", ":8081")
	shutdownTimeout = shared.EnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	logger          = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
		Level:       slog.LevelInfo,
		TimeFormat:  time.DateTime,
		SrcFileMode: slogcolor.ShortFile,
	}))
)

func main() {
	logger.Info("Starting WebSocket server", "port", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := NewServer(logger)
	closer, err := s.Start(ctx, port, listener)
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	// Start HTTP server for profiling
	go func() {
		r := chi.NewRouter()
		r.Mount("/debug", http.DefaultServeMux)
		http.ListenAndServe(":18081", r)
	}()

	if err := shared.WaitForGracefulExit(ctx, shutdownTimeout, closer); err != nil {
		logger.Error("graceful exit error", "error", err)
		os.Exit(1)
	}

	logger.Info("WebSocket server stopped gracefully")
}
