#!/bin/bash
# Build and deploy the Megaport CLI WASM Docker container with Vue 3 Frontend

set -e  # Exit on error

echo "🏗️  Megaport CLI WASM - Vue 3 Frontend Deployment"
echo "=================================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is not installed"
    echo "Please install Node.js and npm to continue"
    exit 1
fi

# Build WASM first (to check for errors)
echo "📦 Building WASM binary..."
GOWORK=off GOOS=js GOARCH=wasm go build -mod=vendor -tags js,wasm -o web/megaport.wasm .
echo "✅ WASM build successful"
echo ""

# Copy wasm_exec.js to web directory
echo "📋 Copying wasm_exec.js..."
# Try new location (Go 1.25+) first, fall back to old location
if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
elif [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/
else
    echo "❌ Error: Could not find wasm_exec.js in Go installation"
    exit 1
fi
echo "✅ wasm_exec.js copied"
echo ""

# Build Vue frontend
echo "🎨 Building Vue 3 frontend..."
cd frontend-integration
npm install --quiet
npm run build:demo
cd ..
echo "✅ Vue frontend build successful"
echo ""

# Copy WASM files to Vue build output
echo "📦 Copying WASM files to Vue build..."
cp web/megaport.wasm web/vue-demo/
cp web/wasm_exec.js web/vue-demo/
echo "✅ WASM files copied to Vue build"
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
echo "🌐 Access the Vue 3 CLI at: http://localhost:8080"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📦 Frontend Build Info:"
echo "  - Framework: Vue 3 + TypeScript"
echo "  - Components: Reusable terminal component library"
echo "  - Build Tool: Vite"
echo "  - Output: web/vue-demo/"
echo ""
echo "Useful commands:"
echo "  📋 View logs:        docker logs -f megaport-cli-wasm"
echo "  🔄 Restart:          docker restart megaport-cli-wasm"
echo "  🛑 Stop:             docker stop megaport-cli-wasm"
echo "  🗑️  Remove:           docker rm megaport-cli-wasm"
echo ""
echo "To login:"
echo "  1. Open http://localhost:8080 in your browser"
echo "  2. Click 'Login' to authenticate with Megaport"
echo "  3. Start using the CLI in your browser!"
echo ""
echo "💡 Ready for Megaport Portal Integration"
echo "  This Vue 3 component can be easily integrated into the"
echo "  Megaport Portal by the frontend team. See HANDOFF.md"
echo "  for integration documentation."
echo ""

# Wait a moment and check if container is healthy
sleep 2
if docker ps --filter name=megaport-cli-wasm --format '{{.Status}}' | grep -q 'Up'; then
    echo "✅ Container is running and healthy"
else
    echo "⚠️  Container may have issues. Check logs: docker logs megaport-cli-wasm"
fi
