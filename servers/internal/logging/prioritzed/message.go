package prioritized

import "time"

// LogMessage represents a log entry in JSON format (replaces protobuf LogRequest)
type LogMessage struct {
	// Common fields (1-15)
	LogID          string    `json:"log_id"`
	RequestID      string    `json:"request_id"`
	SessionID      string    `json:"session_id"`
	ServerID       string    `json:"server_id"`
	EventType      string    `json:"event_type"`
	Severity       string    `json:"severity"` // "DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"
	ServerTime     time.Time `json:"server_time"`
	ServiceType    string    `json:"service_type"` // "INGAME", "QUESTION", "RANKING", "OUTGAME", "STATS", "LOGGING"
	ServiceName    string    `json:"service_name"`
	ServiceVersion string    `json:"service_version"`
	Environment    string    `json:"environment"`
	ServerIP       string    `json:"server_ip"`
	DurationMs     int64     `json:"duration_ms"`
	StatusCode     int32     `json:"status_code"`
	Method         string    `json:"method"`

	// Server fields (16-99)
	Endpoint          string  `json:"endpoint"`
	MemoryUsageMb     int64   `json:"memory_usage_mb"`
	CPUUsagePercent   float32 `json:"cpu_usage_percent"`
	GoroutineCount    int32   `json:"goroutine_count"`
	ActiveConnections int32   `json:"active_connections"`
	GameMode          string  `json:"game_mode"`
	RoomID            string  `json:"room_id"`
	Data              string  `json:"data"`

	// Client fields (100-199)
	ClientIP            string  `json:"client_ip"`
	UserID              string  `json:"user_id"`
	DeviceID            string  `json:"device_id"`
	GameVersion         string  `json:"game_version"`
	Platform            string  `json:"platform"`
	OSVersion           string  `json:"os_version"`
	DeviceModel         string  `json:"device_model"`
	CountryCode         string  `json:"country_code"`
	Language            string  `json:"language"`
	Timezone            string  `json:"timezone"`
	Region              string  `json:"region"`
	NetworkType         string  `json:"network_type"`
	PingMs              string  `json:"ping_ms"`
	FPS                 int32   `json:"fps"`
	ClientMemoryUsageMb int64   `json:"client_memory_usage_mb"`
	ClientCPUUsageCount float32 `json:"client_cpu_usage_count"`
	ClientData          string  `json:"client_data"`
}
