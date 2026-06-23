# Megaport CLI WASM Frontend

This directory contains the WebAssembly frontend for the Megaport CLI, including the Charsm renderer for styled terminal output.

## Testing

### Setup

Install the dependencies:

```bash
npm install
```

### Running Tests

Run all tests:

```bash
npm test
```

Run tests in watch mode (automatically re-run on file changes):

```bash
npm test:watch
```

Run tests with coverage report:

```bash
npm test:coverage
```

### Test Structure

- `charsm-renderer.test.js` - Tests for the Charsm renderer module
- `jest.config.js` - Jest configuration
- `jest.setup.js` - Global test setup and mocks

### Coverage

Coverage reports are generated in the `coverage/` directory after running `npm test:coverage`.

## Files

- `charsm-renderer.js` - Main Charsm renderer implementation
- `xterm-terminal.js` - Xterm.js terminal integration
- `terminal-output.js` - Terminal output handling
- `session.js` - Session management
- `global-helpers.js` - Global helper functions
- `script.js` - Main application script
- `index.html` - Main HTML page
- `wasm_exec.js` - Go WASM execution runtime
- `megaport.wasm` - Compiled WASM binary

## Pre-compressed WASM artifacts

`megaport.wasm` is ~18.5 MiB raw, above CloudFront's 10 MB auto-compression cap, so the
build pre-compresses it at the origin. `cmd/wasmcompress` writes two sibling objects
next to the wasm:

- `megaport.wasm.br` — brotli, quality 11 (~2.95 MiB; preferred)
- `megaport.wasm.gz` — gzip -9 (~4.4 MiB; fallback)

This runs in the **build**, not the CDN sync — `make wasm-compress`, the `deploy.sh`
static build, and the Docker `go-builder` stage all invoke it. The raw identity object
is kept for clients that accept neither encoding.

Whatever uploads these to the CDN origin must serve the compressed objects with
`Content-Type: application/wasm` and the matching `Content-Encoding` (`br` / `gzip`),
shipping `.br`/`.gz`/identity as separate objects selected on `Accept-Encoding`. The
upload itself and the `Accept-Encoding` routing are owned by the infra/CDN tooling, not
this repo.

## Caching

The wasm is content-hashed at build time (`cmd/wasmhash`) into `megaport.<hash>.wasm`,
and `index.html` carries that hashed URL via `window.__MEGAPORT_WASM_URL__`. This only
works if the CDN respects the right cache lifetimes, so the origin server
(`cmd/server`) sets these and any CDN in front must preserve them:

- `megaport.<hash>.wasm` (and its `.br`/`.gz` siblings): `Cache-Control: public,
  max-age=31536000, immutable` — the hash changes when the content does, so it never
  needs revalidating.
- `index.html`: `no-cache` (short TTL at most) — it must be re-fetched so clients pick
  up a new hashed wasm URL.
- `wasm_exec.js`: `no-cache` — it is served from a fixed unhashed path, so it must
  revalidate to stay paired with the wasm's Go runtime after a toolchain upgrade.

A plain unhashed `megaport.wasm` must NOT be served immutable at a flat, unversioned
path. (The S3 publish flow below does serve an unhashed `megaport.wasm`, but only under
a per-version prefix where the path itself is unique, so immutable caching is safe
there.)

## Publishing to S3 (CDN)

`.github/workflows/wasm-publish.yaml` builds the static site and publishes it to a
CloudFront-fronted S3 bucket. It is manual-only (`workflow_dispatch`): a production
publish is always a deliberate action, never an automatic tag push.

Each run uploads the whole static site under two prefixes:

- `<version>/`: immutable, long-lived cache. Treat a version label as write-once, so use
  a fresh tag or `version` input per release.
- `latest/`: short TTL, refreshed on every run.

The portal loads `megaport.wasm` (brotli, served with `Content-Encoding: br`) and
`wasm_exec.js` by static config URL, so unlike the `cmd/server` flow above these
filenames are kept stable rather than content-hashed; the version/latest prefix is the
cache-buster instead.

Authentication uses GitHub OIDC (no long-lived keys). The workflow fails fast with a
clear message if its required repo configuration is missing; see the workflow file for
the expected secrets and variables.

## Development

The Charsm renderer uses the Charsm library (WebAssembly port of lipgloss) to provide styled terminal output in the browser.

To test manually in a browser:

1. Build the WASM binary: `./wasm.sh`
2. Start the server: `./start-server.sh`
3. Open http://localhost:8080 in your browser
