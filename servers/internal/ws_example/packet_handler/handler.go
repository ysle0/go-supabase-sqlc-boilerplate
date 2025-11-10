package packet_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/your-org/go-monorepo-boilerplate/servers/internal/ws_example/session"
)

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

// HandlePacket routes packets to appropriate handlers
func HandlePacket(ctx context.Context, sess *session.Session, data []byte, logger *slog.Logger) error {
	var packet Packet
	if err := json.Unmarshal(data, &packet); err != nil {
		logger.Error("failed to unmarshal packet", "error", err)
		return fmt.Errorf("invalid packet format: %w", err)
	}

	logger.Info("received packet", "type", packet.Type, "remoteAddr", sess.GetRemoteAddr())

	switch packet.Type {
	case PacketTypePing:
		return handlePing(ctx, sess, packet.Data, logger)
	case PacketTypeEcho:
		return handleEcho(ctx, sess, packet.Data, logger)
	default:
		logger.Warn("unknown packet type", "type", packet.Type)
		return fmt.Errorf("unknown packet type: %s", packet.Type)
	}
}

func handlePing(ctx context.Context, sess *session.Session, data json.RawMessage, logger *slog.Logger) error {
	var ping PingPacket
	if err := json.Unmarshal(data, &ping); err != nil {
		return fmt.Errorf("invalid ping packet: %w", err)
	}

	logger.Info("handling ping", "timestamp", ping.Timestamp)

	// Send pong response
	response := map[string]interface{}{
		"type": PacketTypePong,
		"data": PongPacket{
			Timestamp: ping.Timestamp,
		},
	}

	select {
	case sess.SendPacketCh <- response:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func handleEcho(ctx context.Context, sess *session.Session, data json.RawMessage, logger *slog.Logger) error {
	var echo EchoPacket
	if err := json.Unmarshal(data, &echo); err != nil {
		return fmt.Errorf("invalid echo packet: %w", err)
	}

	logger.Info("handling echo", "message", echo.Message)

	// Echo the message back
	response := map[string]interface{}{
		"type": PacketTypeEcho,
		"data": EchoPacket{
			Message: echo.Message,
		},
	}

	select {
	case sess.SendPacketCh <- response:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
