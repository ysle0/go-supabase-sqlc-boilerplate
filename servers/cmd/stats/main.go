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
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/inmem"
)

var (
	port            = shared.EnvString("PORT", ":8084")
	shutdownTimeout = shared.EnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	redisAddr       = shared.EnvString("REDIS_ADDR", "localhost:6379")
	redisPassword   = shared.EnvString("REDIS_PASSWORD", "")
	redisDB         = shared.EnvInt("REDIS_DB", 0)
	logger          = slog.New(slogcolor.NewHandler(os.Stdout, &slogcolor.Options{
		Level:       slog.LevelInfo,
		TimeFormat:  time.DateTime,
		SrcFileMode: slogcolor.ShortFile,
	}))
)

func main() {
	logger.Info("Starting Stats server", "port", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	// Note: Using GetClient with CacheKey for stats consumer
	redisClient := inmem.GetClient(ctx, inmem.CacheKey)
	if redisClient == nil {
		logger.Error("failed to initialize Redis client")
		os.Exit(1)
	}
	logger.Info("connected to Redis")

	s := NewServer(ctx, logger, redisClient)

	// Start HTTP server for metrics and health checks
	go func() {
		r := chi.NewRouter()
		r.Get("/health", s.handleHealth)
		r.Get("/metrics", s.handleMetrics)
		r.Mount("/debug", http.DefaultServeMux)

		logger.Info("HTTP server listening", "port", port)
		if err := http.ListenAndServe(port, r); err != nil {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	// Start consuming events
	if err := s.StartConsumer(ctx); err != nil {
		logger.Error("failed to start consumer", "error", err)
		os.Exit(1)
	}

	if err := shared.WaitForGracefulExit(ctx, shutdownTimeout, s); err != nil {
		logger.Error("graceful exit error", "error", err)
		os.Exit(1)
	}

	logger.Info("Stats server stopped gracefully")
}
