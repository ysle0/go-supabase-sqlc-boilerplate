package logging

import (
	"context"
	"log/slog"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/logging/non_prioritized"
	prioritized "github.com/your-org/go-monorepo-boilerplate/servers/internal/logging/prioritzed"
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

func (s *Server) Start(ctx context.Context, port string) (func() error, error) {
	if err := s.nonPrLogger.Start(port); err != nil {
		s.logger.Error("failed to start logger", "error", err)
		return nil, err
	}

	closer := func() error {
		s.logger.Info("shutting down server")
		if err := s.Shutdown(ctx); err != nil {
			s.logger.Error("failed to shutdown server", "error", err)
			return err
		}
		return nil
	}

	return closer, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}
