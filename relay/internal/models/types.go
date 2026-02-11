// Package models defines shared data structures
package models

import "time"

// Token represents an API token stored in the database
type Token struct {
	ID         int64      `json:"id"`
	Hash       string     `json:"-"` // SHA-256 hash, never exposed
	Name       string     `json:"name"`
	RateLimit  int        `json:"rateLimit"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
}

// Tab represents a browser tab connected via the extension
type Tab struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	FavIconURL string    `json:"favIconUrl,omitempty"`
	AttachedAt time.Time `json:"attachedAt"`
}

// Session represents an extension connection
type Session struct {
	ID             string    `json:"id"`
	TokenHash      string    `json:"-"`
	TokenName      string    `json:"tokenName"`
	Tabs           map[string]*Tab `json:"tabs"`
	ExtensionVer   string    `json:"extensionVersion,omitempty"`
	ConnectedAt    time.Time `json:"connectedAt"`
	LastPingAt     time.Time `json:"lastPingAt"`
}

// --- WebSocket Messages ---

// WSMessage is the base structure for all WebSocket messages
type WSMessage struct {
	Type string `json:"type"`
}

// ConnectAck is sent after successful connection
type ConnectAck struct {
	Type          string `json:"type"` // "connect_ack"
	SessionID     string `json:"sessionId"`
	ServerTime    int64  `json:"serverTime"`
	ServerVersion string `json:"serverVersion"`
}

// ConnectError is sent when connection fails
type ConnectError struct {
	Type    string `json:"type"` // "connect_error"
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TabAttach is received when a tab is attached
type TabAttach struct {
	Type       string `json:"type"` // "tab_attach"
	TabID      string `json:"tabId"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	FavIconURL string `json:"favIconUrl,omitempty"`
}

// TabDetach is received when a tab is detached
type TabDetach struct {
	Type  string `json:"type"` // "tab_detach"
	TabID string `json:"tabId"`
}

// TabUpdate is received when tab info changes
type TabUpdate struct {
	Type  string `json:"type"` // "tab_update"
	TabID string `json:"tabId"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

// Ping is sent to check connection health
type Ping struct {
	Type      string `json:"type"` // "ping"
	Timestamp int64  `json:"timestamp"`
}

// Pong is received in response to ping
type Pong struct {
	Type      string `json:"type"` // "pong"
	Timestamp int64  `json:"timestamp"`
	TabCount  int    `json:"tabCount"`
}

// CommandRequest is sent to execute a command
type CommandRequest struct {
	Type    string        `json:"type"` // "command"
	ID      string        `json:"id"`
	Action  CommandAction `json:"action"`
	TabID   string        `json:"tabId"`
	Timeout int           `json:"timeout"` // ms
}

// CommandAction defines the action to perform
type CommandAction struct {
	Kind        string   `json:"kind"` // click, type, scroll, screenshot, snapshot, navigate, evaluate
	Selector    string   `json:"selector,omitempty"`
	Coordinates *Point   `json:"coordinates,omitempty"`
	Button      string   `json:"button,omitempty"`
	Modifiers   []string `json:"modifiers,omitempty"`
	Text        string   `json:"text,omitempty"`
	Clear       bool     `json:"clear,omitempty"`
	Delay       int      `json:"delay,omitempty"`
	Direction   string   `json:"direction,omitempty"`
	Amount      int      `json:"amount,omitempty"`
	FullPage    bool     `json:"fullPage,omitempty"`
	Clip        *Rect    `json:"clip,omitempty"`
	Quality     int      `json:"quality,omitempty"`
	Format      string   `json:"format,omitempty"`
	MaxDepth    int      `json:"maxDepth,omitempty"`
	MaxLength   int      `json:"maxLength,omitempty"`
	URL         string   `json:"url,omitempty"`
	WaitUntil   string   `json:"waitUntil,omitempty"`
	Script      string   `json:"script,omitempty"`
}

// Point represents x,y coordinates
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Rect represents a rectangle
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CommandResponse is received after command execution
type CommandResponse struct {
	Type    string         `json:"type"` // "command_response"
	ID      string         `json:"id"`
	Success bool           `json:"success"`
	Result  interface{}    `json:"result,omitempty"`
	Error   *CommandError  `json:"error,omitempty"`
	Timing  *CommandTiming `json:"timing,omitempty"`
}

// CommandError contains error details
type CommandError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CommandTiming contains timing information
type CommandTiming struct {
	Received  int64 `json:"received"`
	Completed int64 `json:"completed"`
}

// --- REST API Types ---

// HealthResponse for GET /health
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"` // seconds
}

// StatusResponse for GET /api/v1/status
type StatusResponse struct {
	Connected        bool   `json:"connected"`
	LastSeen         string `json:"lastSeen,omitempty"`
	ExtensionVersion string `json:"extensionVersion,omitempty"`
	TabCount         int    `json:"tabCount,omitempty"`
}

// TabsResponse for GET /api/v1/tabs
type TabsResponse struct {
	Tabs []*Tab `json:"tabs"`
}

// CommandAPIRequest for POST /api/v1/command
type CommandAPIRequest struct {
	TabID   string        `json:"tabId"`
	Action  CommandAction `json:"action"`
	Timeout int           `json:"timeout,omitempty"` // Default 5000ms
}

// CommandAPIResponse for POST /api/v1/command
type CommandAPIResponse struct {
	Success bool          `json:"success"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *CommandError `json:"error,omitempty"`
	Timing  struct {
		Total int64 `json:"total"` // ms
	} `json:"timing,omitempty"`
}

// ScreenshotRequest for POST /api/v1/screenshot
type ScreenshotRequest struct {
	TabID    string `json:"tabId"`
	FullPage bool   `json:"fullPage,omitempty"`
	Format   string `json:"format,omitempty"` // png or jpeg
	Quality  int    `json:"quality,omitempty"` // 0-100 for jpeg
}

// ScreenshotResponse for POST /api/v1/screenshot
type ScreenshotResponse struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size      int    `json:"size"` // bytes
	ExpiresAt string `json:"expiresAt"`
}

// SnapshotRequest for POST /api/v1/snapshot
type SnapshotRequest struct {
	TabID     string `json:"tabId"`
	MaxDepth  int    `json:"maxDepth,omitempty"`  // Default 10
	MaxLength int    `json:"maxLength,omitempty"` // Default 100KB
	Format    string `json:"format,omitempty"`    // html or simplified
}

// SnapshotResponse for POST /api/v1/snapshot
type SnapshotResponse struct {
	HTML                string               `json:"html"`
	URL                 string               `json:"url"`
	Title               string               `json:"title"`
	Truncated           bool                 `json:"truncated"`
	InteractiveElements []InteractiveElement `json:"interactiveElements,omitempty"`
}

// InteractiveElement represents a clickable/interactive element
type InteractiveElement struct {
	Selector    string `json:"selector"`
	Type        string `json:"type"` // button, link, input, select
	Text        string `json:"text,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Error struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		RetryAfter int    `json:"retryAfter,omitempty"` // seconds, for rate limiting
	} `json:"error"`
}
