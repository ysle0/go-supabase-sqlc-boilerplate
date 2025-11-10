package packet_handler

import "encoding/json"

// Packet types
const (
	PacketTypePing      = "ping"
	PacketTypePong      = "pong"
	PacketTypeEcho      = "echo"
	PacketTypeBroadcast = "broadcast"
)

// Base packet structure
type Packet struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Ping packet
type PingPacket struct {
	Timestamp int64 `json:"timestamp"`
}

// Pong packet
type PongPacket struct {
	Timestamp int64 `json:"timestamp"`
}

// Echo packet
type EchoPacket struct {
	Message string `json:"message"`
}

// Broadcast packet
type BroadcastPacket struct {
	Message string `json:"message"`
}
