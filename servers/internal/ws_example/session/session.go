package session

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Session struct {
	c            *websocket.Conn
	SendPacketCh chan any
	RecvPacketCh chan []byte
	isRecvClosed int32 // Use atomic for thread-safe access
	isSendClosed int32 // Use atomic for thread-safe access
	logger       *slog.Logger
	mu           sync.RWMutex // Protect connection state
	closed       int32        // Connection closed flag

	// Session data
	UserID   string // Example: user identifier
	Metadata map[string]interface{}
}

// NewSession creates a new session instance
func NewSession(ctx context.Context, logger *slog.Logger, c *websocket.Conn) *Session {
	s := &Session{
		c:            c,
		logger:       logger,
		SendPacketCh: make(chan any, 100),    // Buffered channel to prevent blocking
		RecvPacketCh: make(chan []byte, 100), // Buffered channel
		Metadata:     make(map[string]interface{}),
	}

	go s.send(ctx)
	go s.recv(ctx)

	return s
}

// IsClosed returns true if the session is closed
func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

// IsHealthy checks if the session is healthy and ready to send/receive
func (s *Session) IsHealthy() bool {
	return !s.IsClosed() &&
		atomic.LoadInt32(&s.isSendClosed) == 0 &&
		atomic.LoadInt32(&s.isRecvClosed) == 0
}

// GetRemoteAddr returns the remote address of the connection
func (s *Session) GetRemoteAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.c != nil {
		return s.c.RemoteAddr().String()
	}
	return ""
}

// Close closes the entire session gracefully
func (s *Session) Close() {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return // Already closed
	}

	s.logger.Info("closing session", "remoteAddr", s.GetRemoteAddr())

	// Close WebSocket connection first
	s.mu.Lock()
	if s.c != nil {
		if err := s.c.Close(); err != nil {
			s.logger.Error("failed closing connection", "error", err)
		}
	}
	s.mu.Unlock()

	s.CloseSend()
	s.CloseRead()
}

func (s *Session) CloseSend() {
	if !atomic.CompareAndSwapInt32(&s.isSendClosed, 0, 1) {
		return // Already closed
	}

	s.logger.Info("closing session send channel")
	close(s.SendPacketCh)
}

func (s *Session) CloseRead() {
	if !atomic.CompareAndSwapInt32(&s.isRecvClosed, 0, 1) {
		return // Already closed
	}

	s.logger.Info("closing session recv channel")
	close(s.RecvPacketCh)
}

func (s *Session) send(ctx context.Context) {
	defer func() {
		s.logger.Info("send goroutine exited")
		s.CloseSend()
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context canceled, exiting send loop")
			return
		case p, ok := <-s.SendPacketCh:
			if !ok {
				s.logger.Info("send channel closed, exiting send loop")
				return
			}

			if s.IsClosed() {
				s.logger.Info("session closed, discarding packet")
				return
			}

			if err := s.sendPacket(p); err != nil {
				s.logger.Error("error sending packet", "error", err)
				s.Close()
				return
			}
		}
	}
}

func (s *Session) recv(ctx context.Context) {
	defer func() {
		s.logger.Info("recv goroutine exited")
		s.CloseRead()
	}()

	for {
		if s.IsClosed() {
			s.logger.Info("session closed, exiting recv loop")
			return
		}

		select {
		case <-ctx.Done():
			s.logger.Info("context canceled, exiting recv loop")
			return
		default:
			s.mu.RLock()
			conn := s.c
			s.mu.RUnlock()

			if conn == nil {
				s.logger.Info("connection is nil, exiting recv loop")
				return
			}

			_, bytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure) {
					s.logger.Info("connection closed by client",
						"client", conn.RemoteAddr(),
						"error", err)
				} else {
					s.logger.Error("error reading message", "error", err)
				}
				s.Close()
				return
			}

			// Try to send to channel with non-blocking send
			select {
			case s.RecvPacketCh <- bytes:
				// Message sent successfully
			case <-ctx.Done():
				s.logger.Info("context canceled while sending to recv channel")
				return
			default:
				s.logger.Warn("recv channel is full, dropping message")
			}
		}
	}
}

func (s *Session) sendPacket(p any) error {
	if s.IsClosed() {
		return websocket.ErrCloseSent
	}

	s.mu.RLock()
	conn := s.c
	s.mu.RUnlock()

	if conn == nil {
		return websocket.ErrCloseSent
	}

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		s.logger.Error("failed getting next writer", "error", err)
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			s.logger.Error("failed closing writer", "error", err)
		}
	}()

	data, err := json.Marshal(p)
	if err != nil {
		s.logger.Error("failed marshaling packet", "error", err)
		return err
	}

	if _, err = w.Write(data); err != nil {
		s.logger.Error("failed writing packet", "error", err)
		return err
	}

	return nil
}
