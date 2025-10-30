# Multi-stage build for Megaport CLI WASM
FROM golang:1.24-alpine AS builder

# Install build dependencies (use HTTP to bootstrap)
RUN sed -i 's/https/http/g' /etc/apk/repositories && \
    apk add --no-cache git make ca-certificates openssl && \
    sed -i 's/http/https/g' /etc/apk/repositories && \
    update-ca-certificates 2>/dev/null || true

# Set SSL/TLS environment for Go
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

WORKDIR /app

# Copy source code (includes go.mod, go.sum, and vendor directory)
COPY . .

# Build the WASM binary using vendored dependencies
RUN GOOS=js GOARCH=wasm go build -mod=vendor -tags js,wasm -o web/megaport.wasm .

# Build the server binary using vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o server ./cmd/server/server.go

# Copy wasm_exec.js to web directory
RUN cp $(go env GOROOT)/misc/wasm/wasm_exec.js web/ || echo "wasm_exec.js not copied"

# Final stage - minimal runtime image
FROM alpine:latest

# Use HTTP repositories temporarily to bootstrap ca-certificates
RUN sed -i 's/https/http/g' /etc/apk/repositories && \
    apk update --no-cache && \
    apk add --no-cache ca-certificates tzdata wget && \
    sed -i 's/http/https/g' /etc/apk/repositories

WORKDIR /app

# Copy server binary
COPY --from=builder /app/server .

# Copy web assets (includes wasm_exec.js copied in builder stage)
COPY --from=builder /app/web ./web

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run server
CMD ["./server", "--port", "8080", "--dir", "web"]
