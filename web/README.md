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

## Publishing to S3 (CDN) and portal integration

`.github/workflows/wasm-publish.yaml` builds the static site and publishes it to the
`media.megaport.com` S3 bucket (fronted by CloudFront). Trigger it manually
(`workflow_dispatch`) or by pushing a `v*` tag.

### Layout

Each run publishes the whole static site under two prefixes inside `portal/megaport-cli/`:

- `portal/megaport-cli/<version>/`: immutable (`max-age=31536000, immutable`). The
  version comes from the tag, a manual input, or `git describe`.
- `portal/megaport-cli/latest/`: short TTL (`max-age=300`), refreshed on every run.

Every file under a prefix (including `wasm_exec.js`) inherits that prefix's cache
lifetime. This differs from the fixed-path `cmd/server` model in the Caching section
above, where `wasm_exec.js` is `no-cache`: here the `<version>/` path is unique so
immutable caching is safe, and `latest/`'s short TTL gives the same freshness `no-cache`
would.

### Stable filenames, not content-hashed

Unlike the `cmd/server` / Docker flow above, this flow does NOT run `cmd/wasmhash`. The
portal loads `megaport.wasm` and `wasm_exec.js` by static config URL and does not read
our `index.html`, so a changing filename would break it. The `<version>/` and `latest/`
prefixes provide the cache-busting instead: a given versioned URL never changes content,
so it is safe to serve immutable.

### What gets uploaded

The whole static site is synced, but the portal only loads two files by URL:
`megaport.wasm` (brotli-compressed, served with `Content-Encoding: br`) and `wasm_exec.js`.
`index.html` and the Vue `assets/` are uploaded as a byproduct of the sync, not as a
browsable site: the built `index.html` references its assets with absolute `/assets/...`
paths, which do not resolve under a versioned prefix.

Only the brotli copy is published as `megaport.wasm` (with `Content-Encoding: br`); no
gzip or identity fallback is uploaded, because the portal's browser audience always
accepts brotli.

### Portal integration

The portal's `wasmUrl` / `wasmExecUrl` point at the `latest/` prefix while the
integration is evolving, then pin to a specific `<version>/` prefix once it is stable:

```js
megaportCli: {
  wasmUrl:     'https://media.megaport.com/portal/megaport-cli/latest/megaport.wasm',
  wasmExecUrl: 'https://media.megaport.com/portal/megaport-cli/latest/wasm_exec.js',
}
```

### Required GitHub config

The workflow authenticates to AWS via OIDC (no long-lived keys) and fails early until
these are set:

| Setting | Kind | Value |
|---|---|---|
| `AWS_ROLE_ARN` | secret | OIDC role with write access to the bucket prefix |
| `AWS_REGION` | var | `ap-southeast-2` |
| `WASM_S3_BUCKET` | var | `media.megaport.com` |
| `WASM_S3_PREFIX` | var | `portal/megaport-cli` |
| `WASM_CLOUDFRONT_DISTRIBUTION_ID` | var | optional; if set, published paths are invalidated, otherwise `latest/` self-refreshes within its TTL |

## Development

The Charsm renderer uses the Charsm library (WebAssembly port of lipgloss) to provide styled terminal output in the browser.

To test manually in a browser:

1. Build the WASM binary: `./wasm.sh`
2. Start the server: `./start-server.sh`
3. Open http://localhost:8080 in your browser
