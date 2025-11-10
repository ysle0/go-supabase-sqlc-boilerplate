package non_prioritized

import (
	"context"
	"log/slog"

	protobufext "github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/pb"
	pb "github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/pb/logger"
	"google.golang.org/grpc"
)

type LoggerClient struct {
	Address        string
	InternalLogger *slog.Logger
	ConnPool       *protobufext.ConnectionPool
}

func NewLoggerClient(
	address string,
	logger *slog.Logger,
) *LoggerClient {
	cp := protobufext.NewConnectionPool()
	return &LoggerClient{
		Address:        address,
		InternalLogger: logger,
		ConnPool:       cp,
	}
}

func (c *LoggerClient) Close(ctx context.Context) error {
	if err := c.ConnPool.Close(ctx); err != nil {
		c.InternalLogger.Error("failed to close connection pool!", "error", err)
		return err
	}
	return nil
}

func (c *LoggerClient) SendLog(
	ctx context.Context,
	in *pb.LogRequest,
	callOpts ...grpc.CallOption,
) (*pb.LogResponse, error) {
	dialOpts := make([]grpc.DialOption, 0, len(protobufext.DefaultDialOpts))
	dialOpts = append(dialOpts, protobufext.DefaultDialOpts...)

	conn, err := c.ConnPool.GetConn(c.Address, dialOpts...)
	if err != nil {
		c.InternalLogger.Error("failed to get connection out of the connection pool!", "error", err)
		return nil, err
	}
	client := pb.NewLoggerClient(conn)

	resp, err := client.SendLog(ctx, in, callOpts...)
	if err != nil {
		c.InternalLogger.Error("failed to send the log to the server!", "error", err)
		return nil, err
	}

	c.InternalLogger.Info("SendLog succeed.", "response", resp)
	return resp, nil
}
