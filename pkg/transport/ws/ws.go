package ws

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/krigsherre/aerocast/pkg/fanout"
	"go.uber.org/zap"
)

type Config struct {
	ListenAddr         string
	PingInterval       time.Duration
	PongTimeout        time.Duration
	MaxFrameSize       int
	WriteBufferPerConn int
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:         "127.0.0.1:9100",
		PingInterval:       30 * time.Second,
		PongTimeout:        10 * time.Second,
		MaxFrameSize:       1024,
		WriteBufferPerConn: 64,
	}
}

type Callbacks interface {
	OnSubscribe(subID uint64, lat, lng, radiusM float64)
	OnUnsubscribe(subID uint64)
	OnFollow(subID uint64, entityID uint32)
	OnUnfollow(subID uint64, entityID uint32)
	OnDisconnect(subID uint64)
	Channels() *fanout.ChannelManager
}

type Server struct {
	cfg      Config
	cb       Callbacks
	upgrader websocket.Upgrader
	srv      *http.Server

	mu    sync.RWMutex
	conns map[*websocket.Conn]uint64
	subID uint64

	totalConns     atomic.Uint64
	broadcastBytes atomic.Uint64

	logger *zap.Logger
}

func NewServer(cfg Config, cb Callbacks, logger *zap.Logger) (*Server, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Server{
		cfg: cfg,
		cb:  cb,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		conns:  make(map[*websocket.Conn]uint64, 1024),
		logger: logger,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)

	s.srv = &http.Server{
		Addr:    s.cfg.ListenAddr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("ws server listening", zap.String("addr", s.cfg.ListenAddr))
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("ws server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.srv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) ConnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

func (s *Server) TotalConnections() uint64 {
	return s.totalConns.Load()
}

func (s *Server) BroadcastBytes() uint64 {
	return s.broadcastBytes.Load()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("ws upgrade failed", zap.Error(err))
		return
	}

	s.mu.Lock()
	s.subID++
	id := s.subID
	s.conns[conn] = id
	s.mu.Unlock()

	s.totalConns.Add(1)

	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		lat = 0.0
	}
	lng, err := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if err != nil {
		lng = 0.0
	}
	radius, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	if err != nil {
		radius = 1000000.0
	}

	s.cb.OnSubscribe(id, lat, lng, radius)

	go s.writeLoop(conn, id)

	s.readLoop(conn, id)

	s.mu.Lock()
	delete(s.conns, conn)
	s.mu.Unlock()

	s.cb.OnDisconnect(id)
	conn.Close()
}

func (s *Server) readLoop(conn *websocket.Conn, id uint64) {
	conn.SetReadLimit(int64(s.cfg.MaxFrameSize))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) writeLoop(conn *websocket.Conn, id uint64) {
	ch := s.cb.Channels().Get(fanout.SubscriberID(id))
	if ch == nil {
		return
	}

	ticker := time.NewTicker(s.cfg.PingInterval)
	defer ticker.Stop()
	defer conn.Close()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(s.cfg.PongTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case frame, ok := <-ch:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(s.cfg.PongTimeout))
			if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
				return
			}
			s.broadcastBytes.Add(uint64(len(frame)))
		}
	}
}
