// Package server provides the HTTP and WebSocket server
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/emreylmaz/owlrelay/relay/internal/config"
	"github.com/emreylmaz/owlrelay/relay/internal/handlers"
	"github.com/emreylmaz/owlrelay/relay/internal/hub"
	"github.com/emreylmaz/owlrelay/relay/internal/store"
)

// Server represents the HTTP server
type Server struct {
	cfg        *config.Config
	httpServer *http.Server
	hub        *hub.Hub
	tokenStore *store.TokenStore
	version    string
}

// New creates a new Server
func New(cfg *config.Config, h *hub.Hub, tokenStore *store.TokenStore, version string) *Server {
	return &Server{
		cfg:        cfg,
		hub:        h,
		tokenStore: tokenStore,
		version:    version,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, restrict this
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// WebSocket endpoint
	r.Get("/ws", s.handleWebSocket)

	// Register HTTP handlers
	h := handlers.New(s.cfg, s.hub, s.tokenStore, s.version)
	h.RegisterRoutes(r, s.tokenStore)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext:  func(l net.Listener) context.Context { return ctx },
	}

	log.Info().
		Str("addr", addr).
		Str("version", s.version).
		Msg("Starting server")

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		return s.shutdown()
	case err := <-errCh:
		return err
	}
}

func (s *Server) shutdown() error {
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin
	},
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" || !strings.HasPrefix(token, "owl_") {
		http.Error(w, `{"type":"connect_error","code":"INVALID_TOKEN","message":"Missing or invalid token"}`, http.StatusUnauthorized)
		return
	}

	// Validate token
	tokenData, err := s.tokenStore.Validate(token)
	if err != nil || tokenData == nil {
		http.Error(w, `{"type":"connect_error","code":"INVALID_TOKEN","message":"Invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	// Register connection with hub
	tokenHash := store.HashToken(token)
	c := s.hub.Register(conn, tokenHash, tokenData.Name)

	// Run connection pumps
	c.Run(r.Context())
}
