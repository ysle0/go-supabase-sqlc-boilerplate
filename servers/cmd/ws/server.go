package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/shared"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/ws_example/packet_handler"
	"github.com/your-org/go-monorepo-boilerplate/servers/internal/ws_example/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

type Server struct {
	logger   *slog.Logger
	sessions sync.Map // map[sessionID]*session.Session
	router   *chi.Mux
	server   *http.Server
}

func NewServer(logger *slog.Logger) *Server {
	s := &Server{
		logger: logger,
	}

	s.router = chi.NewRouter()
	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ws", s.handleWebSocket)
}

type serverCloser struct {
	server   *Server
	httpServ *http.Server
}

func (sc *serverCloser) Close(ctx context.Context) error {
	sc.server.logger.Info("Shutting down WebSocket server...")

	// Close all active sessions
	sc.server.sessions.Range(func(key, value interface{}) bool {
		if sess, ok := value.(*session.Session); ok {
			sess.Close()
		}
		return true
	})

	return sc.httpServ.Shutdown(ctx)
}

func (s *Server) Start(ctx context.Context, port string, listener net.Listener) (shared.Closer, error) {
	s.server = &http.Server{
		Handler: s.router,
	}

	go func() {
		s.logger.Info("WebSocket server listening", slog.String("addr", listener.Addr().String()))
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", "error", err)
		}
	}()

	return &serverCloser{
		server:   s,
		httpServ: s.server,
	}, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("failed to upgrade connection", "error", err)
		return
	}

	ctx := r.Context()
	sess := session.NewSession(ctx, s.logger, conn)
	sessionID := sess.GetRemoteAddr()

	s.logger.Info("new WebSocket connection", "sessionID", sessionID)
	s.sessions.Store(sessionID, sess)

	// Handle packets
	go s.handlePackets(ctx, sess, sessionID)
}

func (s *Server) handlePackets(ctx context.Context, sess *session.Session, sessionID string) {
	defer func() {
		s.logger.Info("session handler exiting", "sessionID", sessionID)
		s.sessions.Delete(sessionID)
		sess.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context canceled", "sessionID", sessionID)
			return
		case data, ok := <-sess.RecvPacketCh:
			if !ok {
				s.logger.Info("receive channel closed", "sessionID", sessionID)
				return
			}

			if err := packet_handler.HandlePacket(ctx, sess, data, s.logger); err != nil {
				s.logger.Error("error handling packet", "error", err, "sessionID", sessionID)
				// Continue processing other packets instead of closing
			}
		}
	}
}
