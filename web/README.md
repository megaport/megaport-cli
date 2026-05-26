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

`megaport.wasm` is ~32 MB raw, above CloudFront's 10 MB auto-compression cap, so the
build pre-compresses it at the origin. `cmd/wasmcompress` writes two sibling objects
next to the wasm:

- `megaport.wasm.br` — brotli, quality 11 (~4.8 MB; preferred)
- `megaport.wasm.gz` — gzip -9 (~7.1 MB; fallback)

This runs in the **build**, not the CDN sync — `make wasm-compress`, the `deploy.sh`
static build, and the Docker `go-builder` stage all invoke it. The raw identity object
is kept for clients that accept neither encoding.

Whatever uploads these to the CDN origin must serve the compressed objects with
`Content-Type: application/wasm` and the matching `Content-Encoding` (`br` / `gzip`),
shipping `.br`/`.gz`/identity as separate objects selected on `Accept-Encoding`. The
upload itself and the `Accept-Encoding` routing are owned by the infra/CDN tooling, not
this repo.

## Development

The Charsm renderer uses the Charsm library (WebAssembly port of lipgloss) to provide styled terminal output in the browser.

To test manually in a browser:

1. Build the WASM binary: `./wasm.sh`
2. Start the server: `./start-server.sh`
3. Open http://localhost:8080 in your browser
