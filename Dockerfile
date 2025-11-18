# Multi-stage build for Megaport CLI WASM with Vue 3 Frontend
# 
# Using base Debian image and installing Node.js from binary distribution
# 
FROM debian:bookworm-slim AS frontend-builder

WORKDIR /app/frontend

# Install Node.js 25 from official binaries (avoids image vulnerabilities and SSL issues)
# Detect architecture and download appropriate binary
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates wget xz-utils libatomic1 && \
    ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then NODE_ARCH="x64"; else NODE_ARCH="arm64"; fi && \
    wget -q https://nodejs.org/dist/v25.2.1/node-v25.2.1-linux-${NODE_ARCH}.tar.xz && \
    tar -xJf node-v25.2.1-linux-${NODE_ARCH}.tar.xz -C /usr/local --strip-components=1 && \
    rm node-v25.2.1-linux-${NODE_ARCH}.tar.xz && \
    npm install -g npm@latest && \
    apt-get remove -y wget xz-utils && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy frontend source
COPY frontend-integration/package*.json ./

# Use npm ci
RUN npm ci --quiet --no-audit --no-fund

COPY frontend-integration/ ./

# Build the Vue frontend demo
RUN npm run build:demo

# Go builder stage
FROM golang:1.24-bookworm AS go-builder

WORKDIR /app

# Copy source code (includes go.mod, go.sum, and vendor directory)
COPY . .

# Build the WASM binary using vendored dependencies
RUN GOOS=js GOARCH=wasm go build -mod=vendor -tags js,wasm -o web/megaport.wasm .

# Build the server binary using vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o server ./cmd/server/server.go

# Copy wasm_exec.js to web directory (handle both Go 1.25+ and older versions)
RUN if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
        cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/; \
    elif [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
        cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/; \
    else \
        echo "Warning: wasm_exec.js not found"; \
    fi

# Final stage - minimal runtime image
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata wget && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy server binary from Go builder
COPY --from=go-builder /app/server .

# Copy Vue frontend build from frontend builder (vite builds to ../web/vue-demo from frontend-integration)
COPY --from=frontend-builder /app/web/vue-demo ./web

# Copy WASM files from Go builder to Vue build directory
COPY --from=go-builder /app/web/megaport.wasm ./web/
COPY --from=go-builder /app/web/wasm_exec.js ./web/

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run server
CMD ["./server", "--port", "8080", "--dir", "web"]
