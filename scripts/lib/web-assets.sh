# shellcheck shell=bash
# Shared build step for the browser/WASM CLI. Source this and call
# build_static_assets from the repo root; build-web.sh wraps it to produce the
# static site for CDN publishing.

build_static_assets() {
  local publish_dir="web/vue-demo"

  local tool
  for tool in go npm; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      echo "error: '$tool' is required but was not found on PATH" >&2
      return 1
    fi
  done

  echo "==> Building Vue front end (frontend-integration -> $publish_dir)"
  (
    # set -e so a failed step aborts the subshell regardless of the caller's options.
    set -e
    cd frontend-integration
    npm ci --quiet
    npm run build:demo
  ) || return 1

  echo "==> Building WASM binary ($publish_dir/megaport.wasm)"
  # GOWORK=off so the build uses the module's pinned deps, not the go.work workspace.
  GOWORK=off GOOS=js GOARCH=wasm go build -tags js,wasm -o "$publish_dir/megaport.wasm" . || return 1

  echo "==> Copying wasm_exec.js"
  local goroot
  if ! goroot="$(go env GOROOT)" || [ -z "$goroot" ]; then
    echo "error: could not determine GOROOT via 'go env GOROOT'" >&2
    return 1
  fi
  if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
    cp "$goroot/lib/wasm/wasm_exec.js" "$publish_dir/" || return 1
  elif [ -f "$goroot/misc/wasm/wasm_exec.js" ]; then
    cp "$goroot/misc/wasm/wasm_exec.js" "$publish_dir/" || return 1
  else
    echo "error: wasm_exec.js not found under $goroot" >&2
    return 1
  fi
}
