package prioritized

import (
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// parse converts a Redis stream message to LogMessage
// Parses all fields explicitly from the Redis XMessage.Values map
func parse(msg redis.XMessage) (LogMessage, error) {
	var parsed LogMessage

	// Helper function to get string value
	getString := func(key string) string {
		if val, exists := msg.Values[key]; exists {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return ""
	}

	// Helper function to parse int64
	parseInt64 := func(key string) int64 {
		if val, exists := msg.Values[key]; exists {
			if str, ok := val.(string); ok {
				if parsed, err := strconv.ParseInt(str, 10, 64); err == nil {
					return parsed
				}
			}
		}
		return 0
	}

	// Helper function to parse int32
	parseInt32 := func(key string) int32 {
		if val, exists := msg.Values[key]; exists {
			if str, ok := val.(string); ok {
				if parsed, err := strconv.ParseInt(str, 10, 32); err == nil {
					return int32(parsed)
				}
			}
		}
		return 0
	}

	// Helper function to parse float32
	parseFloat32 := func(key string) float32 {
		if val, exists := msg.Values[key]; exists {
			if str, ok := val.(string); ok {
				if parsed, err := strconv.ParseFloat(str, 32); err == nil {
					return float32(parsed)
				}
			}
		}
		return 0
	}

	// Helper function to parse time
	parseTime := func(key string) time.Time {
		if val, exists := msg.Values[key]; exists {
			if str, ok := val.(string); ok && str != "" {
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					return parsed
				}
			}
		}
		return time.Time{}
	}

	// Common fields (1-15)
	parsed.LogID = getString("log_id")
	parsed.RequestID = getString("request_id")
	parsed.SessionID = getString("session_id")
	parsed.ServerID = getString("server_id")
	parsed.EventType = getString("event_type")
	parsed.Severity = getString("severity")
	parsed.ServerTime = parseTime("server_time")
	parsed.ServiceType = getString("service_type")
	parsed.ServiceName = getString("service_name")
	parsed.ServiceVersion = getString("service_version")
	parsed.Environment = getString("environment")
	parsed.ServerIP = getString("server_ip")
	parsed.DurationMs = parseInt64("duration_ms")
	parsed.StatusCode = parseInt32("status_code")
	parsed.Method = getString("method")

	// Server fields (16-99)
	parsed.Endpoint = getString("endpoint")
	parsed.MemoryUsageMb = parseInt64("memory_usage_mb")
	parsed.CPUUsagePercent = parseFloat32("cpu_usage_percent")
	parsed.GoroutineCount = parseInt32("goroutine_count")
	parsed.ActiveConnections = parseInt32("active_connections")
	parsed.GameMode = getString("game_mode")
	parsed.RoomID = getString("room_id")
	parsed.Data = getString("data")

	// Client fields (100-199)
	parsed.ClientIP = getString("client_ip")
	parsed.UserID = getString("user_id")
	parsed.DeviceID = getString("device_id")
	parsed.GameVersion = getString("game_version")
	parsed.Platform = getString("platform")
	parsed.OSVersion = getString("os_version")
	parsed.DeviceModel = getString("device_model")
	parsed.CountryCode = getString("country_code")
	parsed.Language = getString("language")
	parsed.Timezone = getString("timezone")
	parsed.Region = getString("region")
	parsed.NetworkType = getString("network_type")
	parsed.PingMs = getString("ping_ms")
	parsed.FPS = parseInt32("fps")
	parsed.ClientMemoryUsageMb = parseInt64("client_memory_usage_mb")
	parsed.ClientCPUUsageCount = parseFloat32("client_cpu_usage_count")
	parsed.ClientData = getString("client_data")

	return parsed, nil
}
