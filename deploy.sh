#!/bin/bash
# Build and deploy the Megaport CLI WASM Docker container

set -e  # Exit on error

echo "🏗️  Megaport CLI WASM - Docker Deployment"
echo "=========================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

# Build WASM first (to check for errors)
echo "📦 Building WASM binary..."
GOWORK=off GOOS=js GOARCH=wasm go build -mod=vendor -tags js,wasm -o web/megaport.wasm .
echo "✅ WASM build successful"
echo ""

# Build Docker image
echo "🐳 Building Docker image..."
docker build --no-cache -t megaport-cli-wasm:latest . > /tmp/docker-build.log 2>&1
echo "✅ Docker image built successfully"
echo ""

# Stop and remove existing container if running
if docker ps -a --format '{{.Names}}' | grep -q '^megaport-cli-wasm$'; then
    echo "🔄 Stopping existing container..."
    docker stop megaport-cli-wasm > /dev/null 2>&1 || true
    docker rm megaport-cli-wasm > /dev/null 2>&1 || true
fi

# Start the container
echo "🚀 Starting container..."
docker run -d \
    --name megaport-cli-wasm \
    -p 8080:8080 \
    megaport-cli-wasm:latest

echo ""
echo "✅ Deployment successful!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 Access the CLI at: http://localhost:8080"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Useful commands:"
echo "  📋 View logs:        docker logs -f megaport-cli-wasm"
echo "  🔄 Restart:          docker restart megaport-cli-wasm"
echo "  🛑 Stop:             docker stop megaport-cli-wasm"
echo "  🗑️  Remove:           docker rm megaport-cli-wasm"
echo ""
echo "To login:"
echo "  1. Open http://localhost:8080 in your browser"
echo "  2. Enter your Megaport Access Key and Secret Key"
echo "  3. Select your environment (production/staging/development)"
echo "  4. Click Login"
echo ""

# Wait a moment and check if container is healthy
sleep 2
if docker ps --filter name=megaport-cli-wasm --format '{{.Status}}' | grep -q 'Up'; then
    echo "✅ Container is running and healthy"
else
    echo "⚠️  Container may have issues. Check logs: docker logs megaport-cli-wasm"
fi
