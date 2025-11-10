package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/stats/consumer"
)

type Server struct {
	ctx           context.Context
	logger        *slog.Logger
	redisClient   *redis.Client
	eventConsumer *consumer.EventConsumer
	processor     *consumer.EventProcessor
}

func NewServer(
	ctx context.Context,
	logger *slog.Logger,
	redisClient *redis.Client,
) *Server {
	processor := consumer.NewEventProcessor(logger)
	eventConsumer := consumer.NewEventConsumer(logger, redisClient, processor)

	return &Server{
		ctx:           ctx,
		logger:        logger,
		redisClient:   redisClient,
		eventConsumer: eventConsumer,
		processor:     processor,
	}
}

func (s *Server) StartConsumer(ctx context.Context) error {
	s.logger.Info("starting event consumer")
	return s.eventConsumer.Start(ctx)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down stats server")

	if err := s.eventConsumer.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown event consumer", "error", err)
		return err
	}

	if err := s.redisClient.Close(); err != nil {
		s.logger.Error("failed to close Redis client", "error", err)
		return err
	}

	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.processor.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		s.logger.Error("failed to encode metrics", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
