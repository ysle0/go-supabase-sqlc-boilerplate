package stats

import "time"

// EventType represents the type of event being tracked
type EventType string

const (
	EventTypeAPICall      EventType = "api_call"
	EventTypeUserAction   EventType = "user_action"
	EventTypeError        EventType = "error"
	EventTypePerformance  EventType = "performance"
)

// Event represents a generic event to be tracked
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// APICallEvent represents an API call event
type APICallEvent struct {
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration"`
	UserID       string        `json:"user_id,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// UserActionEvent represents a user action event
type UserActionEvent struct {
	Action   string                 `json:"action"`
	UserID   string                 `json:"user_id"`
	Resource string                 `json:"resource,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	Context    string `json:"context,omitempty"`
}

// PerformanceEvent represents a performance metric event
type PerformanceEvent struct {
	Operation string        `json:"operation"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsSummary represents aggregated metrics
type MetricsSummary struct {
	TotalEvents      int64                  `json:"total_events"`
	EventsByType     map[EventType]int64    `json:"events_by_type"`
	AverageDuration  float64                `json:"average_duration,omitempty"`
	ErrorRate        float64                `json:"error_rate,omitempty"`
	LastUpdated      time.Time              `json:"last_updated"`
}
