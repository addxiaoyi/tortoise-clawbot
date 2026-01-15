// Gateway - WebSocket gateway for Tortoise protocol

package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/tortoise/server/plugin"
	"github.com/tortoise/server/session"
	"github.com/tortoise/server/protocol"
)

// Config holds gateway configuration
type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PluginHost   *plugin.Host
}

// Gateway handles WebSocket connections
type Gateway struct {
	cfg      *Config
	server   *http.Server
	upgrader websocket.Upgrader
	sessions *session.Manager
	protocol *protocol.Manager
	wg       sync.WaitGroup
	mu       sync.RWMutex
	running  bool
}

// New creates a new gateway
func New(cfg *Config) *Gateway {
	g := &Gateway{
		cfg: cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024 * 4,
			WriteBufferSize: 1024 * 4,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		sessions: session.NewManager(10000),
		protocol: protocol.NewManager(cfg.PluginHost),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", g.handleWebSocket)
	mux.HandleFunc("/health", g.handleHealth)

	g.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return g
}

// Start the gateway
func (g *Gateway) Start(ctx context.Context) error {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return fmt.Errorf("gateway already running")
	}
	g.running = true
	g.mu.Unlock()

	err := g.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Stop the gateway
func (g *Gateway) Stop(ctx context.Context) error {
	g.mu.Lock()
	if !g.running {
		g.mu.Unlock()
		return nil
	}
	g.running = false
	g.mu.Unlock()

	return g.server.Shutdown(ctx)
}

// handleWebSocket handles WebSocket upgrade and connection
func (g *Gateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	g.wg.Add(1)
	go g.handleConnection(conn)
}

// handleConnection handles a single WebSocket connection
func (g *Gateway) handleConnection(conn *websocket.Conn) {
	defer g.wg.Done()
	defer conn.Close()

	// Create session
	sess := g.sessions.NewSession(conn)

	log.Info().Str("session_id", sess.ID()).Msg("New connection")

	// Handle messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn().Err(err).Str("session_id", sess.ID()).Msg("Connection error")
			}
			break
		}

		// Process message
		if err := g.protocol.ProcessMessage(sess, data); err != nil {
			log.Error().Err(err).Str("session_id", sess.ID()).Msg("Message processing error")
			g.sendError(conn, err)
		}
	}

	// Cleanup
	g.sessions.RemoveSession(sess.ID())
	log.Info().Str("session_id", sess.ID()).Msg("Connection closed")
}

// sendError sends an error message
func (g *Gateway) sendError(conn *websocket.Conn, err error) {
	// Implementation
}

// handleHealth handles health check
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","version":"0.1.0"}`))
}

// Stats returns gateway statistics
func (g *Gateway) Stats() SessionStats {
	return g.sessions.Stats()
}

// SessionStats holds session statistics
type SessionStats struct {
	Active int
	Total  int64
}
