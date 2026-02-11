// Package handlers contains HTTP request handlers
package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/emreylmaz/owlrelay/relay/internal/config"
	"github.com/emreylmaz/owlrelay/relay/internal/hub"
	"github.com/emreylmaz/owlrelay/relay/internal/middleware"
	"github.com/emreylmaz/owlrelay/relay/internal/models"
	"github.com/emreylmaz/owlrelay/relay/internal/store"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	cfg        *config.Config
	hub        *hub.Hub
	tokenStore *store.TokenStore
	version    string
	startTime  time.Time
}

// New creates a new Handlers instance
func New(cfg *config.Config, h *hub.Hub, tokenStore *store.TokenStore, version string) *Handlers {
	return &Handlers{
		cfg:        cfg,
		hub:        h,
		tokenStore: tokenStore,
		version:    version,
		startTime:  time.Now(),
	}
}

// Health returns server health status
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	resp := models.HealthResponse{
		Status:  "ok",
		Version: h.version,
		Uptime:  int64(time.Since(h.startTime).Seconds()),
	}
	writeJSON(w, http.StatusOK, resp)
}

// Status returns connection status for the authenticated token
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	tokenHash := middleware.TokenHashFromContext(r.Context())

	if token == nil || tokenHash == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
		return
	}

	session := h.hub.GetSession(tokenHash)

	resp := models.StatusResponse{
		Connected: session != nil,
	}

	if session != nil {
		resp.LastSeen = session.LastPingAt.Format(time.RFC3339)
		resp.ExtensionVersion = session.ExtensionVer
		resp.TabCount = len(session.Tabs)
	}

	writeJSON(w, http.StatusOK, resp)
}

// Tabs returns list of attached tabs
func (h *Handlers) Tabs(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	tokenHash := middleware.TokenHashFromContext(r.Context())

	if token == nil || tokenHash == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
		return
	}

	session := h.hub.GetSession(tokenHash)
	if session == nil {
		writeError(w, http.StatusServiceUnavailable, "EXTENSION_OFFLINE", "Extension is not connected")
		return
	}

	tabs := make([]*models.Tab, 0, len(session.Tabs))
	for _, tab := range session.Tabs {
		tabs = append(tabs, tab)
	}

	writeJSON(w, http.StatusOK, models.TabsResponse{Tabs: tabs})
}

// Command executes a command on the browser
func (h *Handlers) Command(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	tokenHash := middleware.TokenHashFromContext(r.Context())

	if token == nil || tokenHash == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
		return
	}

	var req models.CommandAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}

	if req.TabID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tabId is required")
		return
	}

	if req.Action.Kind == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "action.kind is required")
		return
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5000 // Default 5 seconds
	}

	cmd := &models.CommandRequest{
		Type:    "command",
		ID:      uuid.New().String(),
		Action:  req.Action,
		TabID:   req.TabID,
		Timeout: timeout,
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	resp, err := h.hub.SendCommand(ctx, tokenHash, cmd)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		if hubErr, ok := err.(*hub.HubError); ok {
			statusCode := http.StatusServiceUnavailable
			if hubErr.Code == "TIMEOUT" {
				statusCode = http.StatusGatewayTimeout
			}
			writeError(w, statusCode, hubErr.Code, hubErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	apiResp := models.CommandAPIResponse{
		Success: resp.Success,
		Result:  resp.Result,
		Error:   resp.Error,
	}
	apiResp.Timing.Total = elapsed

	writeJSON(w, http.StatusOK, apiResp)
}

// Screenshot captures a screenshot
func (h *Handlers) Screenshot(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	tokenHash := middleware.TokenHashFromContext(r.Context())

	if token == nil || tokenHash == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
		return
	}

	var req models.ScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}

	if req.TabID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tabId is required")
		return
	}

	format := req.Format
	if format == "" {
		format = "png"
	}

	cmd := &models.CommandRequest{
		Type:  "command",
		ID:    uuid.New().String(),
		TabID: req.TabID,
		Action: models.CommandAction{
			Kind:     "screenshot",
			FullPage: req.FullPage,
			Format:   format,
			Quality:  req.Quality,
		},
		Timeout: h.cfg.CommandTimeout,
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.cfg.CommandTimeout)*time.Millisecond)
	defer cancel()

	resp, err := h.hub.SendCommand(ctx, tokenHash, cmd)
	if err != nil {
		if hubErr, ok := err.(*hub.HubError); ok {
			writeError(w, http.StatusServiceUnavailable, hubErr.Code, hubErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if !resp.Success {
		writeError(w, http.StatusBadRequest, resp.Error.Code, resp.Error.Message)
		return
	}

	// Extract base64 data from result
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid response format")
		return
	}

	data, _ := result["data"].(string)
	width, _ := result["width"].(float64)
	height, _ := result["height"].(float64)

	// Save to file
	filename := uuid.New().String() + "." + format
	filePath := filepath.Join(h.cfg.ScreenshotPath, filename)

	// Decode base64 and save
	if err := saveBase64ToFile(data, filePath); err != nil {
		log.Error().Err(err).Msg("Failed to save screenshot")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save screenshot")
		return
	}

	fileInfo, _ := os.Stat(filePath)
	fileSize := 0
	if fileInfo != nil {
		fileSize = int(fileInfo.Size())
	}

	expiresAt := time.Now().Add(time.Duration(h.cfg.ScreenshotTTL) * time.Second)

	// Schedule cleanup
	go func() {
		time.Sleep(time.Duration(h.cfg.ScreenshotTTL) * time.Second)
		os.Remove(filePath)
	}()

	writeJSON(w, http.StatusOK, models.ScreenshotResponse{
		URL:       "/screenshots/" + filename,
		Width:     int(width),
		Height:    int(height),
		Size:      fileSize,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// Snapshot captures a DOM snapshot
func (h *Handlers) Snapshot(w http.ResponseWriter, r *http.Request) {
	token := middleware.TokenFromContext(r.Context())
	tokenHash := middleware.TokenHashFromContext(r.Context())

	if token == nil || tokenHash == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
		return
	}

	var req models.SnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}

	if req.TabID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tabId is required")
		return
	}

	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 10
	}

	maxLength := req.MaxLength
	if maxLength <= 0 {
		maxLength = 100 * 1024 // 100KB
	}

	cmd := &models.CommandRequest{
		Type:  "command",
		ID:    uuid.New().String(),
		TabID: req.TabID,
		Action: models.CommandAction{
			Kind:      "snapshot",
			MaxDepth:  maxDepth,
			MaxLength: maxLength,
		},
		Timeout: h.cfg.CommandTimeout,
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(h.cfg.CommandTimeout)*time.Millisecond)
	defer cancel()

	resp, err := h.hub.SendCommand(ctx, tokenHash, cmd)
	if err != nil {
		if hubErr, ok := err.(*hub.HubError); ok {
			writeError(w, http.StatusServiceUnavailable, hubErr.Code, hubErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if !resp.Success {
		writeError(w, http.StatusBadRequest, resp.Error.Code, resp.Error.Message)
		return
	}

	// Parse result
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid response format")
		return
	}

	html, _ := result["html"].(string)
	url, _ := result["url"].(string)
	title, _ := result["title"].(string)
	truncated, _ := result["truncated"].(bool)

	writeJSON(w, http.StatusOK, models.SnapshotResponse{
		HTML:      html,
		URL:       url,
		Title:     title,
		Truncated: truncated,
	})
}

// ServeScreenshots serves screenshot files
func (h *Handlers) ServeScreenshots() http.Handler {
	return http.StripPrefix("/screenshots/", http.FileServer(http.Dir(h.cfg.ScreenshotPath)))
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIError{
		Error: struct {
			Code       string `json:"code"`
			Message    string `json:"message"`
			RetryAfter int    `json:"retryAfter,omitempty"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func saveBase64ToFile(base64Data, filePath string) error {
	// Remove data URL prefix if present
	if strings.Contains(base64Data[:min(100, len(base64Data))], ",") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) == 2 {
			base64Data = parts[1]
		}
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, decoded, 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RegisterRoutes registers all API routes
func (h *Handlers) RegisterRoutes(r chi.Router, tokenStore *store.TokenStore) {
	r.Get("/health", h.Health)
	r.Handle("/screenshots/*", h.ServeScreenshots())

	r.Route("/api/v1", func(r chi.Router) {
		// These routes require authentication
		r.Use(middleware.Auth(tokenStore))

		rateLimiter := middleware.NewRateLimiter()
		r.Use(rateLimiter.RateLimit(tokenStore))

		r.Get("/status", h.Status)
		r.Get("/tabs", h.Tabs)
		r.Post("/command", h.Command)
		r.Post("/screenshot", h.Screenshot)
		r.Post("/snapshot", h.Snapshot)
	})
}
