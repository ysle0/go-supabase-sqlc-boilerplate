package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared/redisstream"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/stats"
	"golang.org/x/sync/errgroup"
)

// EventConsumer consumes events from Redis Streams and processes them
type EventConsumer struct {
	logger      *slog.Logger
	redisClient *redis.Client
	consumer    *redisstream.Consumer[stats.Event]
	eventCh     chan stats.Event
	processor   *EventProcessor
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(
	logger *slog.Logger,
	redisClient *redis.Client,
	processor *EventProcessor,
) *EventConsumer {
	eventCh := make(chan stats.Event, 100)

	config := redisstream.Config{
		StreamKey:        "stats:events",
		ConsumerGroup:    "stats-service",
		ConsumerIDPrefix: "stats-consumer",
		BatchSize:        10,
		BlockTime:        5 * time.Second,
		MaxRetries:       5,
		RetryDelay:       2 * time.Second,
		MinIdle:          10 * time.Second,
	}

	consumer := redisstream.NewConsumer(
		logger,
		redisClient,
		eventCh,
		parseEvent,
		config,
	)

	return &EventConsumer{
		logger:      logger,
		redisClient: redisClient,
		consumer:    consumer,
		eventCh:     eventCh,
		processor:   processor,
	}
}

// parseEvent parses a Redis message into an Event
func parseEvent(msg redis.XMessage) (stats.Event, error) {
	var event stats.Event

	// Extract event data from message
	eventJSON, ok := msg.Values["data"].(string)
	if !ok {
		return event, fmt.Errorf("missing 'data' field in message")
	}

	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return event, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return event, nil
}

// Start begins consuming events
func (ec *EventConsumer) Start(ctx context.Context) error {
	eg := new(errgroup.Group)

	// Start the Redis stream consumer
	if err := ec.consumer.ConsumeLoop(ctx, eg); err != nil {
		return fmt.Errorf("failed to start consumer loop: %w", err)
	}

	// Start the event processor
	eg.Go(func() error {
		return ec.processEvents(ctx)
	})

	return nil
}

// processEvents processes events from the channel
func (ec *EventConsumer) processEvents(ctx context.Context) error {
	ec.logger.Info("starting event processor")

	for {
		select {
		case <-ctx.Done():
			ec.logger.Info("stopping event processor")
			return nil
		case event, ok := <-ec.eventCh:
			if !ok {
				ec.logger.Info("event channel closed")
				return nil
			}

			if err := ec.processor.ProcessEvent(ctx, event); err != nil {
				ec.logger.Error("failed to process event",
					"event_id", event.ID,
					"event_type", event.Type,
					"error", err)
				// Continue processing other events
				continue
			}

			ec.logger.Debug("event processed successfully",
				"event_id", event.ID,
				"event_type", event.Type)
		}
	}
}

// Shutdown gracefully stops the consumer
func (ec *EventConsumer) Shutdown(ctx context.Context) error {
	ec.logger.Info("shutting down event consumer")

	eg := new(errgroup.Group)
	if err := ec.consumer.Shutdown(ctx, eg); err != nil {
		return fmt.Errorf("failed to shutdown consumer: %w", err)
	}

	// Close the event channel
	close(ec.eventCh)

	return nil
}
