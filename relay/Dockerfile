# OwlRelay Relay Server - Multi-stage Dockerfile
# Produces a ~15MB scratch image with a single static binary

# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache ca-certificates

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary (CGO disabled for pure Go SQLite)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /relay ./cmd/relay

# --- Production stage ---
FROM scratch

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /relay /relay

# Create data directories (will be volume-mounted in production)
VOLUME ["/data"]

# Environment defaults
ENV PORT=3000
ENV HOST=0.0.0.0
ENV DB_PATH=/data/owlrelay.db
ENV SCREENSHOT_PATH=/data/screenshots
ENV LOG_LEVEL=info
ENV RATE_LIMIT_DEFAULT=100
ENV SCREENSHOT_TTL=30
ENV COMMAND_TIMEOUT=30000

EXPOSE 3000

ENTRYPOINT ["/relay"]
CMD ["serve"]
