# 🦉 OwlRelay - Relay Server

Self-hosted relay server for the OwlRelay browser control system. Connects AI agents to browser extensions via WebSocket/REST API.

## Features

- **WebSocket Hub**: Real-time bidirectional communication with browser extensions
- **REST API**: Simple HTTP API for AI agents to control browsers
- **Token Authentication**: Secure SHA-256 hashed token system
- **Rate Limiting**: In-memory per-token rate limiting (default 100 req/min)
- **SQLite Storage**: Zero-dependency pure Go SQLite (no CGO)
- **Minimal Footprint**: ~15MB Docker image, ~10MB memory usage
- **Graceful Shutdown**: Clean connection handling on shutdown

## Quick Start

### Local Development

```bash
# Clone and build
cd relay
go build ./cmd/relay

# Create a token
./relay token create my-agent

# Start the server
./relay serve
```

### Docker

```bash
# Build and run with docker-compose
docker-compose up -d

# Create a token (in running container)
docker exec owlrelay /relay token create my-agent

# View logs
docker logs -f owlrelay
```

## CLI Commands

```bash
# Start server
relay serve

# Token management
relay token create <name>   # Create new token
relay token list            # List all tokens
relay token revoke <id>     # Revoke a token by ID

# Info
relay version               # Show version
relay help                  # Show help
```

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `DB_PATH` | `./data/owlrelay.db` | SQLite database path |
| `SCREENSHOT_PATH` | `./data/screenshots` | Screenshot storage path |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `RATE_LIMIT_DEFAULT` | `100` | Requests per minute per token |
| `SCREENSHOT_TTL` | `30` | Screenshot TTL in seconds |
| `COMMAND_TIMEOUT` | `30000` | Command timeout in milliseconds |
| `WS_PING_INTERVAL` | `30` | WebSocket ping interval (seconds) |
| `WS_PONG_TIMEOUT` | `10` | WebSocket pong timeout (seconds) |

## API Reference

### Public Endpoints

#### `GET /health`
Health check (no auth required).

```json
{"status":"ok","version":"0.1.0","uptime":123}
```

### Authenticated Endpoints

All `/api/v1/*` endpoints require `Authorization: Bearer owl_xxxxx` header.

#### `GET /api/v1/status`
Check extension connection status.

```json
{"connected":true,"lastSeen":"2026-01-01T12:00:00Z","tabCount":2}
```

#### `GET /api/v1/tabs`
List attached browser tabs.

```json
{
  "tabs": [
    {"id":"abc123","url":"https://example.com","title":"Example","attachedAt":"2026-01-01T12:00:00Z"}
  ]
}
```

#### `POST /api/v1/command`
Execute a browser command.

```json
{
  "tabId": "abc123",
  "action": {
    "kind": "click",
    "selector": "#submit-button"
  },
  "timeout": 5000
}
```

Supported action kinds:
- `click` - Click an element (by selector or coordinates)
- `type` - Type text into an input
- `scroll` - Scroll the page or element
- `navigate` - Navigate to a URL
- `evaluate` - Execute JavaScript

#### `POST /api/v1/screenshot`
Capture a screenshot.

```json
{
  "tabId": "abc123",
  "fullPage": false,
  "format": "png",
  "quality": 90
}
```

Response includes a temporary URL (expires in 30s by default).

#### `POST /api/v1/snapshot`
Capture a DOM snapshot.

```json
{
  "tabId": "abc123",
  "maxDepth": 10,
  "maxLength": 102400
}
```

### WebSocket Connection

Extensions connect via WebSocket:

```
wss://your-relay.com/ws?token=owl_xxxxx
```

Or with header: `Authorization: Bearer owl_xxxxx`

## Project Structure

```
relay/
├── cmd/relay/           # CLI entry point
├── internal/
│   ├── config/          # Environment configuration
│   ├── database/        # SQLite database
│   ├── handlers/        # HTTP handlers
│   ├── hub/             # WebSocket hub
│   ├── middleware/      # Auth & rate limiting
│   ├── models/          # Data types
│   ├── server/          # HTTP server setup
│   └── store/           # Data access layer
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Production Deployment

### With TLS (Recommended)

Use a reverse proxy like Caddy, Nginx, or Traefik for TLS termination:

```yaml
# docker-compose with Caddy
services:
  owlrelay:
    build: .
    expose:
      - "3000"
    volumes:
      - owlrelay-data:/data
    networks:
      - web

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
    networks:
      - web

# Caddyfile
relay.example.com {
    reverse_proxy owlrelay:3000
}
```

### Resource Requirements

- **CPU**: Minimal (< 0.1 core idle)
- **Memory**: ~10-20MB idle, scales with connections
- **Disk**: ~50MB for database + screenshots
- **Network**: Low latency preferred for WebSocket

## Development

```bash
# Run with hot reload (requires air)
air

# Run tests
go test ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o relay-linux-amd64 ./cmd/relay
GOOS=darwin GOARCH=arm64 go build -o relay-darwin-arm64 ./cmd/relay
GOOS=windows GOARCH=amd64 go build -o relay-windows-amd64.exe ./cmd/relay
```

## License

MIT
