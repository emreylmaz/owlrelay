# OwlRelay Chrome Extension

Connect your browser to AI agents securely.

## Development

### Prerequisites

- Node.js 18+
- npm or pnpm

### Setup

```bash
cd extension
npm install
```

### Development Mode

```bash
npm run dev
```

This will start Vite with hot-reload support via CRXJS.

### Load Extension in Chrome

1. Open Chrome and go to `chrome://extensions`
2. Enable "Developer mode" (toggle in top right)
3. Click "Load unpacked"
4. Select the `extension/dist` folder (after running `npm run dev` or `npm run build`)

### Production Build

```bash
npm run build
```

The built extension will be in the `dist` folder.

## Configuration

### Relay URL

Enter the base URL of your OwlRelay server (the extension will automatically append `/ws`).

Examples:
- Local: `ws://localhost:3000`
- Self-hosted: `wss://your-domain.com`

> **Note:** You don't need to add `/ws` to the URL - the extension handles this automatically.

### Token

Get your token from the relay server:

```bash
docker exec owlrelay npx owl-cli token:create --name "my-browser"
```

## Features

- **WebSocket Connection**: Persistent connection to relay server with auto-reconnect
- **Tab Management**: Attach/detach individual tabs for remote control
- **Commands**:
  - Click (selector or coordinates)
  - Type text into inputs
  - Scroll (up/down/left/right)
  - Take screenshots
  - Capture DOM snapshots
- **Security**:
  - Banking and sensitive site blacklist
  - Visual badge indicator when attached
  - Per-tab explicit permission

## Security

The extension includes a hardcoded blacklist of banking and sensitive sites that cannot be controlled. This includes:

- Major banks (Chase, Bank of America, Wells Fargo, etc.)
- Turkish banks (Garanti, İşbank, Yapı Kredi, etc.)
- Password managers (1Password, LastPass, Bitwarden)
- Crypto exchanges (Coinbase, Binance)
- Auth pages (Google, Microsoft, Apple)

You can see the full list in `src/shared/constants.ts`.

## Architecture

```
extension/
├── src/
│   ├── background/    # Service Worker (WebSocket, commands)
│   ├── content/       # Content Script (DOM manipulation)
│   ├── popup/         # Preact UI
│   └── shared/        # Types, constants, utilities
├── manifest.json      # Manifest V3
└── vite.config.ts     # Vite + CRXJS config
```

## License

MIT
