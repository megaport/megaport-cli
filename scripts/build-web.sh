#!/usr/bin/env bash
# Build the browser/WASM CLI's static assets (wasm + wasm_exec.js) into web/dist/.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib/web-assets.sh
source "$repo_root/scripts/lib/web-assets.sh"

build_static_assets

echo ""
echo "Static assets ready: web/dist/"
echo "Publish with the wasm-publish workflow. A plain 'aws s3 sync' uploads the wasm"
echo "uncompressed and with no Cache-Control."
