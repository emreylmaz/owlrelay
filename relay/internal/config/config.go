// Package config handles environment-based configuration
package config

import (
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

// Config holds all configuration values
type Config struct {
	// Server
	Port     int    `envconfig:"PORT" default:"3000"`
	Host     string `envconfig:"HOST" default:"0.0.0.0"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Database
	DBPath string `envconfig:"DB_PATH" default:"./data/owlrelay.db"`

	// Screenshots
	ScreenshotPath string `envconfig:"SCREENSHOT_PATH" default:"./data/screenshots"`
	ScreenshotTTL  int    `envconfig:"SCREENSHOT_TTL" default:"30"` // seconds

	// Rate Limiting
	RateLimitDefault int `envconfig:"RATE_LIMIT_DEFAULT" default:"100"` // requests per minute

	// WebSocket
	WSPingInterval    int `envconfig:"WS_PING_INTERVAL" default:"30"`    // seconds
	WSPongTimeout     int `envconfig:"WS_PONG_TIMEOUT" default:"10"`     // seconds
	WSWriteTimeout    int `envconfig:"WS_WRITE_TIMEOUT" default:"10"`    // seconds
	WSReadBufferSize  int `envconfig:"WS_READ_BUFFER_SIZE" default:"1024"`
	WSWriteBufferSize int `envconfig:"WS_WRITE_BUFFER_SIZE" default:"1024"`

	// Command
	CommandTimeout int `envconfig:"COMMAND_TIMEOUT" default:"30000"` // milliseconds
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.ScreenshotPath, 0755); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetLogLevel returns the zerolog log level
func (c *Config) GetLogLevel() zerolog.Level {
	switch c.LogLevel {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
