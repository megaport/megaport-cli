#!/bin/bash
# Quick restart script for WASM server

echo "🔄 Restarting Megaport CLI WASM Server..."
echo ""

# Kill any existing server on port 8080
echo "Stopping existing server..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || true

# Wait a moment
sleep 1

# Navigate to project root
cd "$(dirname "$0")"

# Start server in background
echo "Starting server on port 8080..."
go run ./cmd/server/server.go --port 8080 --dir web &

# Store PID
SERVER_PID=$!
echo "Server started with PID: $SERVER_PID"
echo ""
echo "✅ Server running at: http://localhost:8080"
echo ""
echo "Available page:"
echo "  - http://localhost:8080/index.html"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Wait for server process
wait $SERVER_PID
