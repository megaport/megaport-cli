#!/bin/bash
# Build and deploy the Megaport CLI WASM Docker container with Vue 3 Frontend

set -e  # Exit on error

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$repo_root"

echo "🏗️  Megaport CLI WASM - Vue 3 Frontend Deployment"
echo "=================================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

# Build the static assets (Vue + WASM) into web/vue-demo/. Shared with
# build-web.sh / make web-static so the two flows can't drift, and doubles as a
# fast pre-flight that surfaces build errors before the slow Docker build.
# shellcheck source-path=SCRIPTDIR
# shellcheck source=scripts/lib/web-assets.sh
source "$repo_root/scripts/lib/web-assets.sh"
build_static_assets
echo ""

# The Docker build compiles offline with -mod=vendor (the golang base image has
# no module-proxy access in restricted networks). vendor/ is gitignored, so
# regenerate it on the host before building the image.
echo "📦 Vendoring Go dependencies for the Docker build..."
GOWORK=off go mod vendor
echo ""

# Content-hash the wasm filename and point index.html at it, so the CDN can serve it
# immutable and only index.html is ever invalidated.
echo "🔖 Content-hashing WASM filename..."
HASHED_WASM=$(GOWORK=off go run -mod=vendor ./cmd/wasmhash web/vue-demo/megaport.wasm web/vue-demo/index.html)
echo "✅ Hashed WASM: $HASHED_WASM"
echo ""

# Pre-compress the hashed WASM for CDN serving (CloudFront skips auto-compression >10MB)
echo "🗜️  Pre-compressing WASM (brotli + gzip)..."
GOWORK=off go run -mod=vendor ./cmd/wasmcompress "$HASHED_WASM"
echo "✅ Compressed artifacts created (.br, .gz)"
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
    -p "127.0.0.1:${PORT:-8080}:8080" \
    megaport-cli-wasm:latest

echo ""
echo "✅ Deployment successful!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 Access the Vue 3 CLI at: http://localhost:${PORT:-8080}"
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
echo "  1. Open http://localhost:${PORT:-8080} in your browser"
echo "  2. Click 'Login' to authenticate with Megaport"
echo "  3. Start using the CLI in your browser!"
echo ""
echo "💡 Ready for Megaport Portal Integration"
echo "  This Vue 3 component can be easily integrated into the"
echo "  Megaport Portal. See frontend-integration/INTEGRATION_GUIDE.md"
echo "  for complete integration documentation."
echo ""

# Wait a moment and check if container is healthy
sleep 2
if docker ps --filter name=megaport-cli-wasm --format '{{.Status}}' | grep -q 'Up'; then
    echo "✅ Container is running and healthy"
else
    echo "⚠️  Container may have issues. Check logs: docker logs megaport-cli-wasm"
fi
