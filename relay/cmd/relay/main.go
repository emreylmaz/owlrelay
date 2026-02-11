package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/emreylmaz/owlrelay/relay/internal/config"
	"github.com/emreylmaz/owlrelay/relay/internal/database"
	"github.com/emreylmaz/owlrelay/relay/internal/hub"
	"github.com/emreylmaz/owlrelay/relay/internal/server"
	"github.com/emreylmaz/owlrelay/relay/internal/store"
)

var version = "0.1.0"

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		runServer()
	case "token":
		handleTokenCommand(os.Args[2:])
	case "version":
		fmt.Printf("owlrelay %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`🦉 OwlRelay - Browser Control Relay Server

Usage:
  relay serve              Start the relay server
  relay token create       Create a new token
  relay token list         List all tokens
  relay token revoke <id>  Revoke a token by ID
  relay version            Show version
  relay help               Show this help

Environment Variables:
  PORT            Server port (default: 3000)
  HOST            Server host (default: 0.0.0.0)
  DB_PATH         SQLite database path (default: ./data/owlrelay.db)
  SCREENSHOT_PATH Screenshot storage path (default: ./data/screenshots)
  LOG_LEVEL       Log level: debug, info, warn, error (default: info)
  
  RATE_LIMIT_DEFAULT     Requests per minute per token (default: 100)
  SCREENSHOT_TTL         Screenshot TTL in seconds (default: 30)
  COMMAND_TIMEOUT        Command timeout in ms (default: 30000)
  
  WS_PING_INTERVAL       WebSocket ping interval in seconds (default: 30)
  WS_PONG_TIMEOUT        WebSocket pong timeout in seconds (default: 10)

Examples:
  # Start server on default port
  relay serve

  # Create a token with custom name
  relay token create my-agent

  # List all tokens
  relay token list

  # Revoke a token
  relay token revoke 1`)
}

func runServer() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Set log level
	zerolog.SetGlobalLevel(cfg.GetLogLevel())

	// Initialize database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Create stores
	tokenStore := store.NewTokenStore(db)

	// Create hub
	h := hub.New(cfg, version)

	// Create and start server
	srv := server.New(cfg, h, tokenStore, version)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	// Start server
	if err := srv.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server error")
	}

	log.Info().Msg("Server stopped gracefully")
}

func handleTokenCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: relay token <create|list|revoke>")
		os.Exit(1)
	}

	// Load config and initialize database
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	tokenStore := store.NewTokenStore(db)

	switch args[0] {
	case "create":
		name := "default"
		if len(args) > 1 {
			name = args[1]
		}

		token, err := tokenStore.Create(name, cfg.RateLimitDefault)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating token: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("✅ Token created successfully!\n\n")
		fmt.Printf("Token: %s\n", token)
		fmt.Printf("Name:  %s\n", name)
		fmt.Println()
		fmt.Println("⚠️  Save this token securely. It won't be shown again.")
		fmt.Println()
		fmt.Println("To connect your extension, use:")
		fmt.Printf("  Relay URL: http://localhost:%d\n", cfg.Port)
		fmt.Printf("  Token:     %s\n", token)

	case "list":
		tokens, err := tokenStore.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing tokens: %v\n", err)
			os.Exit(1)
		}

		if len(tokens) == 0 {
			fmt.Println("No tokens found. Create one with: relay token create <name>")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tRATE LIMIT\tCREATED\tLAST USED\tSTATUS")
		fmt.Fprintln(w, "--\t----\t----------\t-------\t---------\t------")

		for _, t := range tokens {
			lastUsed := "never"
			if t.LastUsedAt != nil {
				lastUsed = t.LastUsedAt.Format("2006-01-02 15:04")
			}

			status := "active"
			if t.RevokedAt != nil {
				status = "revoked"
			}

			fmt.Fprintf(w, "%d\t%s\t%d/min\t%s\t%s\t%s\n",
				t.ID,
				t.Name,
				t.RateLimit,
				t.CreatedAt.Format("2006-01-02"),
				lastUsed,
				status,
			)
		}
		w.Flush()

	case "revoke":
		if len(args) < 2 {
			fmt.Println("Usage: relay token revoke <id>")
			os.Exit(1)
		}

		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid token ID: %s\n", args[1])
			os.Exit(1)
		}

		if err := tokenStore.Revoke(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error revoking token: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Token %d revoked successfully.\n", id)

	default:
		fmt.Printf("Unknown token command: %s\n", args[0])
		fmt.Println("Usage: relay token <create|list|revoke>")
		os.Exit(1)
	}
}
