package non_prioritized

import (
	"context"
	"log/slog"
	"net"

	pb "github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/pb/logger"
	"google.golang.org/grpc"
)

type LogHandler struct {
	pb.UnimplementedLoggerServer
	logger *slog.Logger
}

func NewLogHandler(logger *slog.Logger) *LogHandler {
	return &LogHandler{logger: logger}
}

func (l *LogHandler) Start(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		l.logger.Error("failed to listen",
			"error", err,
			"port", port,
		)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterLoggerServer(s, l)

	l.logger.Info("listening grpc server", "address", listener.Addr())
	if err := s.Serve(listener); err != nil {
		l.logger.Error("failed to serve",
			"error", err,
			"address", listener.Addr())
		return err
	}
	return nil
}

func (l *LogHandler) SendLog(ctx context.Context, in *pb.LogRequest) (*pb.LogResponse, error) {
	resp, err := l.SendLog(ctx, in)
	if err != nil {
		l.logger.Error("error> nonPrLogger.SendLog failed", "error", err)
		return nil, err
	}

	l.logger.Info("response> nonPrLogger.SendLog done",
		"is_success", resp.Success,
		"response", resp.GetMessage(),
	)
	return resp, nil
}
