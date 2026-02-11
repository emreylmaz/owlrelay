package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var version = "dev"

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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
	default:
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
  relay token revoke <id>  Revoke a token
  relay version            Show version

Environment:
  PORT            Server port (default: 3000)
  DB_PATH         SQLite database path (default: ./data/owlrelay.db)
  SCREENSHOT_PATH Screenshot storage path (default: ./data/screenshots)
  LOG_LEVEL       Log level: debug, info, warn, error (default: info)`)
}

func runServer() {
	log.Info().Str("version", version).Msg("Starting OwlRelay server...")

	// TODO: Initialize config, db, and start server
	// cfg := config.Load()
	// db := db.New(cfg.DBPath)
	// srv := server.New(cfg, db)
	// srv.Start()

	log.Info().Msg("Server started on :3000")
	log.Info().Msg("Waiting for connections...")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down...")
}

func handleTokenCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: relay token <create|list|revoke>")
		os.Exit(1)
	}

	switch args[0] {
	case "create":
		name := "default"
		if len(args) > 1 {
			name = args[1]
		}
		// TODO: Generate token and save to DB
		fmt.Printf("Token created: owl_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n")
		fmt.Printf("Name: %s\n", name)
		fmt.Println("\n⚠️  Save this token securely. It won't be shown again.")
	case "list":
		// TODO: List tokens from DB
		fmt.Println("ID\tName\tCreated\t\tLast Used")
		fmt.Println("1\tdefault\t2026-01-01\t2026-01-02")
	case "revoke":
		if len(args) < 2 {
			fmt.Println("Usage: relay token revoke <id>")
			os.Exit(1)
		}
		// TODO: Revoke token
		fmt.Printf("Token %s revoked.\n", args[1])
	default:
		fmt.Println("Usage: relay token <create|list|revoke>")
		os.Exit(1)
	}
}
