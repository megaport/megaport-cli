#!/usr/bin/env bash
# Build the browser/WASM CLI into a self-contained static site at web/vue-demo/.
# No Docker and no Go server — the output is ready to `aws s3 sync` to a CDN.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

# shellcheck source-path=SCRIPTDIR
# shellcheck source=lib/web-assets.sh
source "$repo_root/scripts/lib/web-assets.sh"

build_static_assets

echo ""
echo "Static site ready: web/vue-demo/"
echo "Publish with: aws s3 sync web/vue-demo/ s3://<bucket>/<prefix>/ --delete"
