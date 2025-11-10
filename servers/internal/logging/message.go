package logging

import (
	"time"
)

type LogMessage struct {
	LogID          string    `json:"log_id"`
	EventType      string    `json:"event_type"`
	ReceivedTime   time.Time `json:"received_time"`
	ServerTime     time.Time `json:"server_time"`
	UserID         string    `json:"user_id"`
	DeviceID       string    `json:"device_id"`
	GameVersion    string    `json:"game_version"`
	CountryCode    string    `json:"country_code"`
	Language       string    `json:"language"`
	Timezone       string    `json:"timezone"`
	IpAddress      string    `json:"ip_address"`
	ConnectionType string    `json:"connection_type"`
	Data           string    `bson:"data"`
}
