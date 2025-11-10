package prioritized

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type LoggerClient struct {
	redisClient    *redis.Client
	internalLogger *slog.Logger
}

func NewLoggerClient(
	redisClient *redis.Client,
	logger *slog.Logger,
) *LoggerClient {
	return &LoggerClient{
		redisClient:    redisClient,
		internalLogger: logger,
	}
}

func (c *LoggerClient) SendLog(ctx context.Context, message LogMessage) error {
	streamKey, err := c.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "log",
		ID:     "*",
		Values: message,
	}).Result()
	if err != nil {
		c.internalLogger.Error("failed to add to log stream", "error", err)
		return err
	}
	c.internalLogger.Info("successfully added to log stream", "created_stream_key", streamKey)

	return nil
}
