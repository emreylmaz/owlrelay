// Package hub manages WebSocket connections from extensions
package hub

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/emreylmaz/owlrelay/relay/internal/config"
	"github.com/emreylmaz/owlrelay/relay/internal/models"
)

// Hub manages all WebSocket connections
type Hub struct {
	cfg *config.Config

	// Connections indexed by token hash
	sessions   map[string]*Connection
	sessionsMu sync.RWMutex

	// Pending commands waiting for response
	pending   map[string]chan *models.CommandResponse
	pendingMu sync.RWMutex

	// Server version for handshake
	version string
}

// Connection represents a WebSocket connection from an extension
type Connection struct {
	Session *models.Session
	Conn    *websocket.Conn
	Send    chan []byte
	hub     *Hub
	done    chan struct{}
}

// New creates a new Hub
func New(cfg *config.Config, version string) *Hub {
	return &Hub{
		cfg:      cfg,
		sessions: make(map[string]*Connection),
		pending:  make(map[string]chan *models.CommandResponse),
		version:  version,
	}
}

// Register adds a new connection
func (h *Hub) Register(conn *websocket.Conn, tokenHash, tokenName string) *Connection {
	session := &models.Session{
		ID:          uuid.New().String(),
		TokenHash:   tokenHash,
		TokenName:   tokenName,
		Tabs:        make(map[string]*models.Tab),
		ConnectedAt: time.Now().UTC(),
		LastPingAt:  time.Now().UTC(),
	}

	c := &Connection{
		Session: session,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		hub:     h,
		done:    make(chan struct{}),
	}

	h.sessionsMu.Lock()
	// Close existing connection for this token if any
	if existing, ok := h.sessions[tokenHash]; ok {
		close(existing.done)
		existing.Conn.Close()
	}
	h.sessions[tokenHash] = c
	h.sessionsMu.Unlock()

	log.Info().
		Str("session_id", session.ID).
		Str("token_name", tokenName).
		Msg("Extension connected")

	// Send connect ack
	ack := models.ConnectAck{
		Type:          "connect_ack",
		SessionID:     session.ID,
		ServerTime:    time.Now().UnixMilli(),
		ServerVersion: h.version,
	}
	if data, err := json.Marshal(ack); err == nil {
		c.Send <- data
	}

	return c
}

// Unregister removes a connection
func (h *Hub) Unregister(c *Connection) {
	h.sessionsMu.Lock()
	if existing, ok := h.sessions[c.Session.TokenHash]; ok && existing == c {
		delete(h.sessions, c.Session.TokenHash)
	}
	h.sessionsMu.Unlock()

	close(c.done)
	c.Conn.Close()

	log.Info().
		Str("session_id", c.Session.ID).
		Str("token_name", c.Session.TokenName).
		Msg("Extension disconnected")
}

// GetSession returns the session for a token hash
func (h *Hub) GetSession(tokenHash string) *models.Session {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()

	if c, ok := h.sessions[tokenHash]; ok {
		return c.Session
	}
	return nil
}

// GetConnection returns the connection for a token hash
func (h *Hub) GetConnection(tokenHash string) *Connection {
	h.sessionsMu.RLock()
	defer h.sessionsMu.RUnlock()
	return h.sessions[tokenHash]
}

// SendCommand sends a command to the extension and waits for response
func (h *Hub) SendCommand(ctx context.Context, tokenHash string, cmd *models.CommandRequest) (*models.CommandResponse, error) {
	h.sessionsMu.RLock()
	c, ok := h.sessions[tokenHash]
	h.sessionsMu.RUnlock()

	if !ok || c == nil {
		return nil, ErrNotConnected
	}

	// Create response channel
	respChan := make(chan *models.CommandResponse, 1)
	h.pendingMu.Lock()
	h.pending[cmd.ID] = respChan
	h.pendingMu.Unlock()

	defer func() {
		h.pendingMu.Lock()
		delete(h.pending, cmd.ID)
		h.pendingMu.Unlock()
	}()

	// Send command
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	select {
	case c.Send <- data:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, ErrNotConnected
	}

	// Wait for response
	timeout := time.Duration(cmd.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = time.Duration(h.cfg.CommandTimeout) * time.Millisecond
	}

	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, ErrNotConnected
	}
}

// HandleResponse handles a command response from the extension
func (h *Hub) HandleResponse(resp *models.CommandResponse) {
	h.pendingMu.RLock()
	ch, ok := h.pending[resp.ID]
	h.pendingMu.RUnlock()

	if ok {
		select {
		case ch <- resp:
		default:
		}
	}
}

// Run starts the read and write pumps for a connection
func (c *Connection) Run(ctx context.Context) {
	go c.writePump(ctx)
	c.readPump(ctx)
}

func (c *Connection) readPump(ctx context.Context) {
	defer c.hub.Unregister(c)

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(time.Duration(c.hub.cfg.WSPingInterval+c.hub.cfg.WSPongTimeout) * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(time.Duration(c.hub.cfg.WSPingInterval+c.hub.cfg.WSPongTimeout) * time.Second))
		c.Session.LastPingAt = time.Now().UTC()
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
		}

		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn().Err(err).Str("session_id", c.Session.ID).Msg("WebSocket read error")
			}
			return
		}

		c.handleMessage(message)
	}
}

func (c *Connection) writePump(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(c.hub.cfg.WSPingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(c.hub.cfg.WSWriteTimeout) * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Warn().Err(err).Str("session_id", c.Session.ID).Msg("WebSocket write error")
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(c.hub.cfg.WSWriteTimeout) * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) handleMessage(data []byte) {
	var msg models.WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Err(err).Str("session_id", c.Session.ID).Msg("Failed to parse message")
		return
	}

	switch msg.Type {
	case "tab_attach":
		var attach models.TabAttach
		if err := json.Unmarshal(data, &attach); err != nil {
			return
		}
		c.Session.Tabs[attach.TabID] = &models.Tab{
			ID:         attach.TabID,
			URL:        attach.URL,
			Title:      attach.Title,
			FavIconURL: attach.FavIconURL,
			AttachedAt: time.Now().UTC(),
		}
		log.Debug().Str("tab_id", attach.TabID).Str("url", attach.URL).Msg("Tab attached")

	case "tab_detach":
		var detach models.TabDetach
		if err := json.Unmarshal(data, &detach); err != nil {
			return
		}
		delete(c.Session.Tabs, detach.TabID)
		log.Debug().Str("tab_id", detach.TabID).Msg("Tab detached")

	case "tab_update":
		var update models.TabUpdate
		if err := json.Unmarshal(data, &update); err != nil {
			return
		}
		if tab, ok := c.Session.Tabs[update.TabID]; ok {
			if update.URL != "" {
				tab.URL = update.URL
			}
			if update.Title != "" {
				tab.Title = update.Title
			}
		}

	case "pong":
		var pong models.Pong
		if err := json.Unmarshal(data, &pong); err != nil {
			return
		}
		c.Session.LastPingAt = time.Now().UTC()

	case "command_response":
		var resp models.CommandResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return
		}
		c.hub.HandleResponse(&resp)

	default:
		log.Debug().Str("type", msg.Type).Msg("Unknown message type")
	}
}

// Errors
var (
	ErrNotConnected = &HubError{Code: "EXTENSION_OFFLINE", Message: "Extension is not connected"}
	ErrTimeout      = &HubError{Code: "TIMEOUT", Message: "Command timed out"}
)

// HubError represents a hub-related error
type HubError struct {
	Code    string
	Message string
}

func (e *HubError) Error() string {
	return e.Message
}
