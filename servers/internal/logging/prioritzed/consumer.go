package prioritized

import (
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/redisstream"
)

var (
	consumerGroup     = shared.EnvString("LOGGING_CONSUMER_GROUP", "logging-group")
	maxRetries        = shared.EnvInt("LOGGING_MAX_RETRIES", 3)
	retryDelay        = shared.EnvDuration("LOGGING_RETRY_DELAY", 5*time.Second)
	consumerBlockTime = shared.EnvDuration("LOGGING_CONSUMER_BLOCK_TIME", 3*time.Second)
	batchSize         = shared.EnvInt("LOGGING_BATCH_SIZE", 100)
	minIdle           = 5 * time.Minute
)

// Consumer is an alias for the generic Redis stream consumer
type Consumer = redisstream.Consumer[LogMessage]

// NewConsumer creates a new logging consumer using the shared redisstream consumer
// TODO: This consumer currently uses a temporary channel. It should be updated to accept
// a proper transfer channel when the logging server is fully implemented.
func NewConsumer(logger *slog.Logger, redisClient *redis.Client) *Consumer {
	streamKey := internal.CreateOrUpdateQuestionStatsKey

	// TODO: Replace this temporary channel with a proper transfer channel
	// when the logging server is fully implemented
	tempTransferCh := make(chan LogMessage, 100)
	go func() {
		// Drain the channel to prevent blocking
		for range tempTransferCh {
			// Messages are discarded for now
		}
	}()

	config := redisstream.Config{
		StreamKey:        streamKey,
		ConsumerGroup:    consumerGroup,
		ConsumerIDPrefix: "logging-consumer",
		BatchSize:        batchSize,
		BlockTime:        consumerBlockTime,
		MaxRetries:       maxRetries,
		RetryDelay:       retryDelay,
		MinIdle:          minIdle,
	}

	return redisstream.NewConsumer(
		logger,
		redisClient,
		tempTransferCh,
		parse, // Use the parse function from parser.go
		config,
	)
}
