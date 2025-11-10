package logging

import (
	"context"
	"log/slog"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/logging/non_prioritized"
	prioritized "github.com/your-org/go-monorepo-boilerplate/servers/internal/logging/prioritzed"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
)

type Server struct {
	logger      *slog.Logger
	nonPrLogger *non_prioritized.LogHandler
	prLogger    *prioritized.Consumer
}

func NewServer(logger *slog.Logger) *Server {
	nonPrLogger := non_prioritized.NewLogHandler(logger)
	prLogger := prioritized.NewConsumer(logger, nil)

	return &Server{
		logger:      logger,
		nonPrLogger: nonPrLogger,
		prLogger:    prLogger,
	}
}

type loggingServerCloser struct {
	server *Server
}

func (lsc *loggingServerCloser) Close(ctx context.Context) error {
	lsc.server.logger.Info("shutting down server")
	return lsc.server.Shutdown(ctx)
}

func (s *Server) Start(ctx context.Context, port string) (shared.Closer, error) {
	if err := s.nonPrLogger.Start(port); err != nil {
		s.logger.Error("failed to start logger", "error", err)
		return nil, err
	}

	return &loggingServerCloser{
		server: s,
	}, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}
