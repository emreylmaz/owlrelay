# 🦉 OwlRelay

**Self-hosted browser control relay for AI agents.**

Let AI agents interact with your browser securely through a lightweight Chrome extension and relay server.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://hub.docker.com/)

## 🎯 What is OwlRelay?

OwlRelay bridges AI agents (like Claude, GPT, OpenClaw) with your browser. Unlike Puppeteer/Playwright that run headless browsers, OwlRelay uses your **existing browser session** — no re-login needed.

```
┌─────────────┐     WebSocket      ┌──────────────┐      REST      ┌──────────────┐
│   Chrome    │ ──────────────────>│    Relay     │<──────────────│   AI Agent   │
│  Extension  │     (outbound)     │    Server    │    (API)      │  (OpenClaw)  │
└─────────────┘                    └──────────────┘               └──────────────┘
```

## ✨ Features

- **🔒 Self-hosted** — Run on your own server, no cloud dependency
- **🆓 100% Free** — Open source, MIT license, no pricing tiers
- **⚡ Lightweight** — ~15MB Docker image, ~5MB RAM
- **🔐 Secure** — Token auth, banking site blacklist, TLS
- **🌐 Your Session** — Use existing browser logins, no re-auth

## 🚀 Quick Start

### 1. Deploy Relay Server

```bash
# Clone the repo
git clone https://github.com/emreylmaz/owlrelay.git
cd owlrelay/relay

# Start with Docker
docker-compose up -d

# Generate a token
docker exec owlrelay ./relay token create --name "my-agent"
# Output: owl_xxxxxxxxxxxxxxxxxxxxx
```

### 2. Install Chrome Extension

1. Download the [latest release](https://github.com/emreylmaz/owlrelay/releases)
2. Go to `chrome://extensions/`
3. Enable "Developer mode"
4. Click "Load unpacked" and select the extension folder
5. Click the OwlRelay icon and enter:
   - Relay URL: `https://relay.yourdomain.com`
   - Token: `owl_xxxxxxxxxxxxxxxxxxxxx`

### 3. Connect from AI Agent

```bash
# Check status
curl -H "Authorization: Bearer owl_xxx" \
  https://relay.yourdomain.com/api/v1/status

# Click a button
curl -X POST -H "Authorization: Bearer owl_xxx" \
  -H "Content-Type: application/json" \
  -d '{"tabId": "...", "action": {"kind": "click", "selector": "#submit"}}' \
  https://relay.yourdomain.com/api/v1/command
```

## 📦 Components

| Component | Description | Status |
|-----------|-------------|--------|
| `relay/` | Go relay server | 🚧 In Progress |
| `extension/` | Chrome extension (Manifest V3) | 🚧 In Progress |
| `sdk/` | TypeScript SDK for agents | 📋 Planned |
| `docs/` | Documentation | 📋 Planned |

## 🔧 API Reference

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/status` | Connection status |
| GET | `/api/v1/tabs` | List connected tabs |
| POST | `/api/v1/command` | Execute command (click, type, scroll) |
| POST | `/api/v1/screenshot` | Capture screenshot |
| POST | `/api/v1/snapshot` | Get DOM snapshot |

### Commands

```json
// Click
{"kind": "click", "selector": "#button"}

// Type
{"kind": "type", "selector": "#input", "text": "Hello"}

// Scroll
{"kind": "scroll", "direction": "down", "amount": 500}

// Screenshot
{"kind": "screenshot", "fullPage": false}

// DOM Snapshot
{"kind": "snapshot", "maxDepth": 10}
```

## 🔐 Security

- **Token Auth** — SHA-256 hashed tokens, never stored plaintext
- **Site Blacklist** — Banking, password managers, crypto sites blocked
- **TLS Required** — WSS/HTTPS enforced in production
- **Rate Limiting** — 100 requests/minute per token
- **User Control** — One-click disconnect, visual indicator

## 🛠️ Development

```bash
# Relay server
cd relay
go mod download
go run ./cmd/relay

# Extension
cd extension
npm install
npm run dev
```

## 📄 License

MIT License — see [LICENSE](LICENSE)

## 🤝 Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

---

Made with 🦉 by [Emre Yılmaz](https://github.com/emreylmaz)
