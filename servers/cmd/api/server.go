package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/feature/user_profile"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
	sharedMiddleware "github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/middleware"
)

type Server struct {
	ctx                context.Context
	router             *chi.Mux
	logger             *slog.Logger
	httpRequestTimeout time.Duration
	httpServer         *http.Server
}

func NewServer(
	ctx context.Context,
	router *chi.Mux,
	logger *slog.Logger,
	httpRequestTimeout time.Duration,
) *Server {
	s := &Server{
		ctx:                ctx,
		router:             router,
		logger:             logger,
		httpRequestTimeout: httpRequestTimeout,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(s.httpRequestTimeout))
	s.router.Use(sharedMiddleware.MeasureExecTime(s.logger))
	s.router.Use(sharedMiddleware.ApiVersionWith("v1"))
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)

	// API v1 routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Example ping endpoint
		r.Get("/ping", s.handlePing)
	})

	// Vertical slice architecture routes
	user_profile.MapRoutes(s.router, "v1")
}

type apiServerCloser struct {
	server     *Server
	httpServer *http.Server
}

func (asc *apiServerCloser) Close(ctx context.Context) error {
	asc.server.logger.Info("Shutting down API server...")
	return asc.httpServer.Shutdown(ctx)
}

func (s *Server) Start(ctx context.Context, port string) (shared.Closer, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	s.httpServer = &http.Server{
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		s.logger.Info("API server listening", slog.String("addr", listener.Addr().String()))
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", "error", err)
		}
	}()

	return &apiServerCloser{
		server:     s,
		httpServer: s.httpServer,
	}, nil
}

// Health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// Readiness check endpoint
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// Add any readiness checks here (database, redis, etc.)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

// Example ping endpoint
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"pong"}`))
}
