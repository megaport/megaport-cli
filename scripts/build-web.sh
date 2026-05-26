#!/usr/bin/env bash
# Build the browser/WASM CLI into a self-contained static site at web/vue-demo/.
# No Docker and no Go server — the output is ready to `aws s3 sync` to a CDN.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

echo "==> Building WASM binary (web/megaport.wasm)"
# GOWORK=off builds this module against its pinned deps rather than the go.work
# workspace — required when run from a git worktree or the monorepo root.
GOWORK=off GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .

echo "==> Copying wasm_exec.js"
goroot="$(go env GOROOT)"
if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
  cp "$goroot/lib/wasm/wasm_exec.js" web/
elif [ -f "$goroot/misc/wasm/wasm_exec.js" ]; then
  cp "$goroot/misc/wasm/wasm_exec.js" web/
else
  echo "error: wasm_exec.js not found under $goroot" >&2
  exit 1
fi

echo "==> Building Vue front end (frontend-integration -> web/vue-demo)"
(
  cd frontend-integration
  # The lockfile is gitignored, so fall back to npm install on a clean checkout
  # where npm ci would have nothing to install from.
  if [ -f package-lock.json ]; then
    npm ci --quiet
  else
    npm install --quiet
  fi
  npm run build:demo
)

echo "==> Copying WASM files into the published dir"
cp web/megaport.wasm web/wasm_exec.js web/vue-demo/

echo ""
echo "Static site ready: web/vue-demo/"
echo "Publish with: aws s3 sync web/vue-demo/ s3://<bucket>/<prefix>/ --delete"
