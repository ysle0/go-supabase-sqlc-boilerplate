package redisstream

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds configuration for Redis stream consumer
type Config struct {
	StreamKey        string
	ConsumerGroup    string
	ConsumerIDPrefix string
	BatchSize        int
	BlockTime        time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	MinIdle          time.Duration
}

// ParseFunc is a function that parses a Redis message into type T
type ParseFunc[T any] func(redis.XMessage) (T, error)

// Consumer handles Redis stream consumption with proper error handling and graceful shutdown
type Consumer[T any] struct {
	consumerID string
	client     *redis.Client
	logger     *slog.Logger
	transferCh chan<- T
	parseFunc  ParseFunc[T]
	exitCh     chan struct{}
	ErrCh      chan error
	config     Config
}

// NewConsumer creates a new generic Redis stream consumer
func NewConsumer[T any](
	logger *slog.Logger,
	redisClient *redis.Client,
	transferCh chan<- T,
	parseFunc ParseFunc[T],
	config Config,
) *Consumer[T] {
	logger.Info("stream consumer config",
		"consumer_group", config.ConsumerGroup,
		"stream_key", config.StreamKey,
		"batch_size", config.BatchSize,
		"block_time", config.BlockTime,
	)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	consumerID := fmt.Sprintf("%s-%s-%d", config.ConsumerIDPrefix, hostname, os.Getpid())

	consumer := &Consumer[T]{
		logger:     logger,
		client:     redisClient,
		consumerID: consumerID,
		transferCh: transferCh,
		parseFunc:  parseFunc,
		exitCh:     make(chan struct{}, 1),
		ErrCh:      make(chan error, 1),
		config:     config,
	}

	return consumer
}

// ConsumeLoop begins consuming messages from the Redis stream
func (c *Consumer[T]) ConsumeLoop(ctx context.Context, wg *sync.WaitGroup) error {
	// initConsumerGroup creates the consumer group if it doesn't exist
	if err := c.client.XGroupCreateMkStream(
		ctx,
		c.config.StreamKey,
		c.config.ConsumerGroup,
		"0-0",
	).Err(); err != nil {
		// ignore BUSYGROUP error if it already exists
		if strings.Contains(err.Error(), "BUSYGROUP") {
			c.logger.Info("consumer group already exists, proceeding")
		} else {
			c.logger.Error("failed to create consumer group", "error", err)
			return fmt.Errorf("failed to create consumer group '%s': %w", c.config.ConsumerGroup, err)
		}
	}

	c.logger.Info("consumer group ready; stream consumer initialized; start loop",
		"consumer_id", c.consumerID,
		"group", c.config.ConsumerGroup,
		"stream_key", c.config.StreamKey)

	wg.Go(func() {
		c.consumeLoop(ctx)
	})

	return nil
}

// consumeLoop is the main consumption loop with error handling and recovery
func (c *Consumer[T]) consumeLoop(ctx context.Context) {
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consume loop stopping due to context cancellation")
			return
		case <-c.exitCh:
			c.logger.Info("consume loop stopping due to shutdown signal")
			return
		default:
			// Process pending messages first
			c.processPendingMessages(ctx, c.ErrCh)
			c.readMessages(ctx, c.ErrCh)

			select {
			case err := <-c.ErrCh:
				c.logger.Warn("error during message processing",
					"error", err,
					"retry_count", retryCount,
				)
				retryCount++

				if retryCount >= c.config.MaxRetries {
					c.logger.Error("max retries exceeded; stop consumeLoop.", "error", err)
					c.ErrCh <- fmt.Errorf("max retries (%d) exceeded: %w", c.config.MaxRetries, err)
					return
				}

				// Exponential backoff
				backoffDelay := time.Duration(retryCount) * c.config.RetryDelay
				c.logger.Info("waiting before retry",
					"delay", backoffDelay,
					"retry_count", retryCount,
					"backoff_delay", backoffDelay,
				)

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoffDelay):
					continue
				}
			default:
				retryCount = 0
			}
		}
	}
}

// readMessages reads new messages from the stream using XREADGROUP
func (c *Consumer[T]) readMessages(ctx context.Context, errCh chan error) {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.config.ConsumerGroup,
		Consumer: c.consumerID,
		Streams:  []string{c.config.StreamKey, ">"},
		Count:    int64(c.config.BatchSize),
		Block:    c.config.BlockTime,
	}).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			errCh <- fmt.Errorf("XREADGROUP failed: %w", err)
		}
		return
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := c.processMessage(message); err != nil {
				c.logger.Error("failed to process message",
					"message_id", message.ID,
					"error", err)
				continue
			}

			// Acknowledge the message
			if err = c.ackMessage(ctx, message.ID); err != nil {
				c.logger.Error("failed to acknowledge message",
					"message_id", message.ID,
					"error", err)
				continue
			}
		}
	}
}

// processPendingMessages handles messages that were delivered but not acknowledged
func (c *Consumer[T]) processPendingMessages(ctx context.Context, errCh chan error) {
	allPending, err := c.getAllPendingMessages(ctx, errCh)
	if err != nil {
		c.logger.Error("failed to get all pending messages", "error", err)
		return
	}

	for _, pending := range allPending {
		c.logger.Info("Pending message details",
			"message_id", pending.ID,
			"consumer", pending.Consumer,
			"retry_count", pending.RetryCount,
			"idle_time", pending.Idle)

		shouldClaim, minIdle := c.shouldClaimMessage(pending)
		c.logger.Info("Claim condition check",
			"should_claim", shouldClaim,
			"retry_count", pending.RetryCount,
			"max_retries", c.config.MaxRetries,
			"idle", pending.Idle,
			"own_consumer", pending.Consumer == c.consumerID)

		if shouldClaim {
			messages, err := c.claimMessage(ctx, pending, minIdle)
			if err != nil {
				continue
			}

			c.logger.Info("Successfully claimed message",
				"message_id", pending.ID,
				"claimed_count", len(messages))

			c.processClaimedMessages(ctx, messages)
		}
	}
}

func (c *Consumer[T]) claimMessage(ctx context.Context, e redis.XPendingExt, minIdle time.Duration) ([]redis.XMessage, error) {
	c.logger.Info("claiming pending message",
		"message_id", e.ID,
		"consumer", e.Consumer,
		"retry_count", e.RetryCount,
		"idle_time", e.Idle)

	args := &redis.XClaimArgs{
		Stream:   c.config.StreamKey,
		Group:    c.config.ConsumerGroup,
		Consumer: c.consumerID,
		MinIdle:  minIdle,
		Messages: []string{e.ID},
	}

	msgs, err := c.client.XClaim(ctx, args).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.logger.Error("failed to claim message",
			"message_id", e.ID,
			slog.Any("args", args),
			"error", err)
		return nil, err
	}
	return msgs, nil
}

// getAllPendingMessages retrieves all pending messages across all consumers
func (c *Consumer[T]) getAllPendingMessages(ctx context.Context, errCh chan error) ([]redis.XPendingExt, error) {
	totalPending, err := c.client.XPending(ctx, c.config.StreamKey, c.config.ConsumerGroup).Result()
	if err != nil {
		c.logger.Error("Failed to get total pending count", "error", err)
		errCh <- fmt.Errorf("failed to get total pending count: %w", err)
		return nil, err
	}

	var allPending []redis.XPendingExt
	for consumer, count := range totalPending.Consumers {
		if count == 0 {
			continue
		}

		args := &redis.XPendingExtArgs{
			Stream:   c.config.StreamKey,
			Group:    c.config.ConsumerGroup,
			Start:    "-",
			End:      "+",
			Consumer: consumer,
		}
		detailedPending, err := c.client.XPendingExt(ctx, args).Result()
		if err != nil {
			c.logger.Error("XPENDINGEXT Failed",
				slog.Any("args", args),
				"error", err)
			errCh <- fmt.Errorf("failed to get detailed pending for consumer %s: %w", consumer, err)
			continue
		}

		allPending = append(allPending, detailedPending...)
	}

	return allPending, nil
}

// shouldClaimMessage determines if a pending message should be claimed
func (c *Consumer[T]) shouldClaimMessage(e redis.XPendingExt) (bool, time.Duration) {
	if e.Consumer == c.consumerID {
		return true, 0
	}

	hasReachedAtMaxRetries := e.RetryCount > int64(c.config.MaxRetries)
	isIdleTimeout := e.Idle > c.config.MinIdle
	shouldClaim := hasReachedAtMaxRetries || isIdleTimeout

	return shouldClaim, c.config.MinIdle
}

// processClaimedMessages processes and acknowledges claimed messages
func (c *Consumer[T]) processClaimedMessages(ctx context.Context, messages []redis.XMessage) {
	for _, msg := range messages {
		if err := c.processMessage(msg); err != nil {
			c.logger.Error("failed to process claimed message",
				slog.String("message_id", msg.ID),
				slog.String("error", err.Error()))
			continue
		}

		if err := c.ackMessage(ctx, msg.ID); err != nil {
			c.logger.Error("failed to acknowledge claimed message",
				slog.String("message_id", msg.ID),
				slog.String("error", err.Error()))
			continue
		}
	}
}

// processMessage handles a single message
func (c *Consumer[T]) processMessage(msg redis.XMessage) error {
	c.logger.Debug("processing message",
		"message_id", msg.ID,
		slog.Any("values", msg.Values))

	// Parse the message using the injected parse function
	parsedMsg, err := c.parseFunc(msg)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Send to transfer channel
	c.transferCh <- parsedMsg
	c.logger.Info("message processed successfully",
		"message_id", msg.ID)
	return nil
}

// ackMessage acknowledges a processed message
func (c *Consumer[T]) ackMessage(ctx context.Context, messageID string) error {
	if ackCount, err := c.client.XAck(ctx, c.config.StreamKey, c.config.ConsumerGroup, messageID).Result(); err != nil {
		return fmt.Errorf("XACK failed for message %s: %w", messageID, err)
	} else {
		c.logger.Debug("message acknowledged",
			"message_id", messageID,
			"ack_count", ackCount)
		return nil
	}
}

// Shutdown gracefully stops the consumer
func (c *Consumer[T]) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	c.logger.Info("initiating consumer shutdown")

	// Signal the consumption loop to stop
	close(c.exitCh)

	// Wait for the consumption loop to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("consumer stopped gracefully")
		return nil
	case <-ctx.Done():
		c.logger.Warn("consumer shutdown timed out")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
