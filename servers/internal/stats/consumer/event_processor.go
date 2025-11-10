package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/stats"
)

// EventProcessor processes and aggregates events
type EventProcessor struct {
	logger   *slog.Logger
	metrics  *stats.MetricsSummary
	mu       sync.RWMutex

	// In a real implementation, you would store these in a database
	// For this example, we'll use in-memory storage
	apiCallMetrics map[string]int64
	userActions    map[string]int64
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(logger *slog.Logger) *EventProcessor {
	return &EventProcessor{
		logger: logger,
		metrics: &stats.MetricsSummary{
			EventsByType: make(map[stats.EventType]int64),
			LastUpdated:  time.Now(),
		},
		apiCallMetrics: make(map[string]int64),
		userActions:    make(map[string]int64),
	}
}

// ProcessEvent processes a single event
func (ep *EventProcessor) ProcessEvent(ctx context.Context, event stats.Event) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.logger.Info("processing event",
		"event_id", event.ID,
		"event_type", event.Type,
		"timestamp", event.Timestamp)

	// Update total events
	ep.metrics.TotalEvents++
	ep.metrics.EventsByType[event.Type]++
	ep.metrics.LastUpdated = time.Now()

	// Process based on event type
	switch event.Type {
	case stats.EventTypeAPICall:
		return ep.processAPICallEvent(ctx, event)
	case stats.EventTypeUserAction:
		return ep.processUserActionEvent(ctx, event)
	case stats.EventTypeError:
		return ep.processErrorEvent(ctx, event)
	case stats.EventTypePerformance:
		return ep.processPerformanceEvent(ctx, event)
	default:
		ep.logger.Warn("unknown event type", "event_type", event.Type)
		return nil
	}
}

func (ep *EventProcessor) processAPICallEvent(ctx context.Context, event stats.Event) error {
	var apiCall stats.APICallEvent

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &apiCall); err != nil {
		return fmt.Errorf("failed to unmarshal API call event: %w", err)
	}

	// Track API call metrics
	endpoint := fmt.Sprintf("%s %s", apiCall.Method, apiCall.Path)
	ep.apiCallMetrics[endpoint]++

	ep.logger.Debug("processed API call event",
		"method", apiCall.Method,
		"path", apiCall.Path,
		"status_code", apiCall.StatusCode,
		"duration", apiCall.Duration)

	// In a real implementation, you would:
	// - Store in database
	// - Update time-series metrics
	// - Trigger alerts if needed

	return nil
}

func (ep *EventProcessor) processUserActionEvent(ctx context.Context, event stats.Event) error {
	var userAction stats.UserActionEvent

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &userAction); err != nil {
		return fmt.Errorf("failed to unmarshal user action event: %w", err)
	}

	// Track user action metrics
	ep.userActions[userAction.Action]++

	ep.logger.Debug("processed user action event",
		"action", userAction.Action,
		"user_id", userAction.UserID,
		"resource", userAction.Resource)

	return nil
}

func (ep *EventProcessor) processErrorEvent(ctx context.Context, event stats.Event) error {
	var errorEvent stats.ErrorEvent

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &errorEvent); err != nil {
		return fmt.Errorf("failed to unmarshal error event: %w", err)
	}

	ep.logger.Error("processed error event",
		"message", errorEvent.Message,
		"context", errorEvent.Context,
		"user_id", errorEvent.UserID)

	// In a real implementation, you would:
	// - Store error details
	// - Update error rate metrics
	// - Send to error tracking service (e.g., Sentry)
	// - Trigger alerts

	return nil
}

func (ep *EventProcessor) processPerformanceEvent(ctx context.Context, event stats.Event) error {
	var perfEvent stats.PerformanceEvent

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &perfEvent); err != nil {
		return fmt.Errorf("failed to unmarshal performance event: %w", err)
	}

	ep.logger.Debug("processed performance event",
		"operation", perfEvent.Operation,
		"duration", perfEvent.Duration,
		"success", perfEvent.Success)

	return nil
}

// GetMetrics returns the current metrics summary
func (ep *EventProcessor) GetMetrics() stats.MetricsSummary {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	// Create a copy to avoid race conditions
	metricsCopy := stats.MetricsSummary{
		TotalEvents:  ep.metrics.TotalEvents,
		EventsByType: make(map[stats.EventType]int64),
		LastUpdated:  ep.metrics.LastUpdated,
	}

	for k, v := range ep.metrics.EventsByType {
		metricsCopy.EventsByType[k] = v
	}

	return metricsCopy
}
