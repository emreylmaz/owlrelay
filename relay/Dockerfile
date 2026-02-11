# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install CA certificates
RUN apk add --no-cache ca-certificates

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo dev)" \
    -o /relay ./cmd/relay

# Final stage (scratch = minimal image)
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /relay /relay

# Data volume
VOLUME ["/data"]

# Environment defaults
ENV PORT=3000
ENV DB_PATH=/data/owlrelay.db
ENV SCREENSHOT_PATH=/data/screenshots
ENV LOG_LEVEL=info

EXPOSE 3000

ENTRYPOINT ["/relay", "serve"]
