#!/usr/bin/env bash
# Build the browser/WASM CLI into a self-contained static site at web/vue-demo/.
# No Docker and no Go server — the output is ready to `aws s3 sync` to a CDN.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

publish_dir="web/vue-demo"

echo "==> Building Vue front end (frontend-integration -> $publish_dir)"
(
  cd frontend-integration
  # package-lock.json is gitignored, so npm ci (which requires a lockfile and
  # errors out without one) won't work on a clean checkout — fall back to npm install.
  if [ -f package-lock.json ]; then
    npm ci --quiet
  else
    npm install --quiet
  fi
  npm run build:demo
)

# Build the WASM files straight into the publish dir. vite's emptyOutDir wipes
# web/vue-demo first, so this has to run after the front-end build — and writing
# here (rather than into the tracked web/ files) keeps the working tree clean.
echo "==> Building WASM binary ($publish_dir/megaport.wasm)"
# GOWORK=off builds this module against its pinned deps rather than the go.work
# workspace — required when run from a git worktree or the monorepo root.
GOWORK=off GOOS=js GOARCH=wasm go build -tags js,wasm -o "$publish_dir/megaport.wasm" .

echo "==> Copying wasm_exec.js"
goroot="$(go env GOROOT)"
if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
  cp "$goroot/lib/wasm/wasm_exec.js" "$publish_dir/"
elif [ -f "$goroot/misc/wasm/wasm_exec.js" ]; then
  cp "$goroot/misc/wasm/wasm_exec.js" "$publish_dir/"
else
  echo "error: wasm_exec.js not found under $goroot" >&2
  exit 1
fi

echo ""
echo "Static site ready: $publish_dir/"
echo "Publish with: aws s3 sync $publish_dir/ s3://<bucket>/<prefix>/ --delete"
