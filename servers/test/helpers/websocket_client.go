package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Packet tags for WebSocket communication
const (
	C2S_Ping                  = "C2S_Ping"
	S2C_Pong                  = "S2C_Pong"
	C2S_ReadyGame             = "C2S_ReadyGame"
	S2C_ReadyGame             = "S2C_ReadyGame"
	C2S_ReadyGameChallenge    = "C2S_ReadyGameChallenge"
	S2C_ReadyGameChallenge    = "S2C_ReadyGameChallenge"
	C2S_StartGame             = "C2S_StartGame"
	S2C_StartGame             = "S2C_StartGame"
	C2S_SelectAnswer          = "C2S_SelectAnswer"
	S2C_SelectAnswer          = "S2C_SelectAnswer"
	C2S_SelectAnswerChallenge = "C2S_SelectAnswerChallenge"
	S2C_SelectAnswerChallenge = "S2C_SelectAnswerChallenge"
	S2C_Timeover              = "S2C_Timeover"
	S2C_SyncTime              = "S2C_SyncTime"
	C2S_PauseGame             = "C2S_PauseGame"
	C2S_ResumeGame            = "C2S_ResumeGame"
	C2S_ExitGame              = "C2S_ExitGame"
	C2S_UseItem               = "C2S_UseItem"
	S2C_UseItem               = "S2C_UseItem"
	C2S_UseItemWithGem        = "C2S_UseItemWithGem"
	S2C_UseItemWithGem        = "S2C_UseItemWithGem"
)

// TestWebSocketClient is a WebSocket client for testing
type TestWebSocketClient struct {
	conn *websocket.Conn
	url  string
}

// ReceivedPacket represents a received packet with its tag and data
type ReceivedPacket struct {
	Tag  string
	Data json.RawMessage
}

// Connect establishes a WebSocket connection to the specified URL
func (c *TestWebSocketClient) Connect(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.url = urlStr
	return nil
}

// Disconnect closes the WebSocket connection
func (c *TestWebSocketClient) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

// SendPacket sends a packet with the specified tag and data
func (c *TestWebSocketClient) SendPacket(tag string, data interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Marshal data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal packet data: %w", err)
	}

	// Create packet wrapper
	packetWrapper := map[string]interface{}{
		"tag":  tag,
		"data": json.RawMessage(dataBytes),
	}

	// Marshal entire packet
	packetBytes, err := json.Marshal(packetWrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	// Send packet
	err = c.conn.WriteMessage(websocket.TextMessage, packetBytes)
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}

	return nil
}

// ReceivePacket receives a single packet with timeout
func (c *TestWebSocketClient) ReceivePacket(timeout time.Duration) (*ReceivedPacket, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	if timeout > 0 {
		deadline := time.Now().Add(timeout)
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	// Read message
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		// Don't wrap close errors so they can be checked properly
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	// Parse packet
	var packetWrapper struct {
		Tag  string          `json:"tag"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &packetWrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal packet: %w", err)
	}

	return &ReceivedPacket{
		Tag:  packetWrapper.Tag,
		Data: packetWrapper.Data,
	}, nil
}

// ExpectPacket waits for a packet with the specified tag within the timeout
// Returns the packet data if found, or an error if timeout or wrong tag received
func (c *TestWebSocketClient) ExpectPacket(expectedTag string, timeout time.Duration) (json.RawMessage, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for packet with tag '%s'", expectedTag)
		}

		remainingTime := time.Until(deadline)
		pkt, err := c.ReceivePacket(remainingTime)
		if err != nil {
			return nil, err
		}

		if pkt.Tag == expectedTag {
			return pkt.Data, nil
		}

		// Wrong tag, continue waiting
		// Note: This will discard non-matching packets
	}
}

// ReceiveMultiple collects multiple packets within the specified duration
// This is useful for collecting periodic packets like S2C_SyncTime
func (c *TestWebSocketClient) ReceiveMultiple(duration time.Duration) ([]*ReceivedPacket, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	var packets []*ReceivedPacket
	deadline := time.Now().Add(duration)

	for time.Now().Before(deadline) {
		remainingTime := time.Until(deadline)
		if remainingTime <= 0 {
			break
		}

		pkt, err := c.ReceivePacket(remainingTime)
		if err != nil {
			// Timeout is expected when collecting multiple packets
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				break
			}
			// Check if it's a timeout error
			if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
				break
			}
			// Other errors should be returned
			return packets, err
		}

		packets = append(packets, pkt)
	}

	return packets, nil
}

// ReceiveMultipleWithTag collects multiple packets with the specified tag within the duration
func (c *TestWebSocketClient) ReceiveMultipleWithTag(tag string, duration time.Duration) ([]*ReceivedPacket, error) {
	allPackets, err := c.ReceiveMultiple(duration)
	if err != nil {
		return nil, err
	}

	var filtered []*ReceivedPacket
	for _, pkt := range allPackets {
		if pkt.Tag == tag {
			filtered = append(filtered, pkt)
		}
	}

	return filtered, nil
}

// SendPing sends a C2S_Ping packet and waits for S2C_Pong
func (c *TestWebSocketClient) SendPing(timeout time.Duration) error {
	if err := c.SendPacket(C2S_Ping, map[string]interface{}{}); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	_, err := c.ExpectPacket(S2C_Pong, timeout)
	if err != nil {
		return fmt.Errorf("failed to receive pong: %w", err)
	}

	return nil
}

// IsConnected returns true if the client is connected
func (c *TestWebSocketClient) IsConnected() bool {
	return c.conn != nil
}

// UnmarshalPacketData unmarshals the packet data into the provided struct
func UnmarshalPacketData(data json.RawMessage, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal packet data: %w", err)
	}
	return nil
}

// NewTestWebSocketClient creates a new WebSocket test client
func NewTestWebSocketClient() *TestWebSocketClient {
	return &TestWebSocketClient{}
}

// ConnectAndWait connects to the WebSocket server and waits for the connection to be established
func (c *TestWebSocketClient) ConnectAndWait(ctx context.Context, urlStr string, waitDuration time.Duration) error {
	if err := c.Connect(urlStr); err != nil {
		return err
	}

	// Wait a bit for the connection to be fully established
	select {
	case <-time.After(waitDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
