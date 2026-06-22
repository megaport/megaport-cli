# Megaport CLI - WebAssembly (WASM) Browser Terminal

A browser-based terminal for the Megaport CLI powered by **WebAssembly (WASM)**. The
CLI is compiled to a `.wasm` binary that runs entirely in the browser, so customers can
use it without installing anything locally.

## What is This?

- **Runs in the browser** - No local installation required
- **Powered by WebAssembly** - Go code compiled to WASM runs directly in the browser
- **In-browser authentication** - Credentials stay in browser memory; the WASM calls the Megaport API directly, with no server-side component
- **XTerm.js Terminal** - Full-featured terminal emulator with ANSI support
- **Early Release** - Covers the main resource modules (ports, MCR, MVE, VXC, locations, partners, service keys). Auth, config, completion, `generate-docs`, and `version` are not applicable in the browser.

## How Customers Use It

The browser CLI is embedded in the Megaport Portal, which loads the WASM binary from the
CDN (`s3://media.megaport.com/portal/megaport-cli/`) and runs it in-page. There is no
server to deploy: the front end loads static assets and the WASM authenticates against
the Megaport API directly.

### Available Commands

The WASM build registers the following modules. Each module exposes the same subcommands it provides in the native CLI (so `partners` is still `list` / `find`, `locations` is still `list` / `get`, and so on), subject to the constraints of running in the browser:

- `locations` - Find Megaport locations and metros
- `ports` - Manage Megaport ports (including LAG ports)
- `mcr` - Manage Megaport Cloud Routers
- `mve` - Manage Megaport Virtual Edge devices
- `vxc` - Manage Virtual Cross Connects
- `partners` - Look up cloud partner ports
- `servicekeys` - Manage service keys

Any module not in that list is unavailable in the WASM build, including `auth`, `config`, `completion`, `generate-docs`, `version`, `nat-gateway`, `ix`, `users`, `status`, `topology`, `apply`, `product`, `managed-account`, and `billing-market`. Some rely on the local filesystem (profile storage, completion scripts); others have simply not been wired into the browser build yet.

**Output Formats**: The WASM build supports the following output formats:

- `--output table` (default) - Formatted table with styled output
- `--output json` - JSON format for programmatic use
- `--output csv` - CSV format for data export
- `--output xml` - XML format

`--output go-template` is not supported in WASM/browser builds.

**Example Commands**:

```bash
megaport-cli locations list
megaport-cli locations list --output json
megaport-cli ports list
megaport-cli ports get <portUID>
megaport-cli vxc list --status LIVE
megaport-cli partners list --company-name "Amazon Web Services"
```

## Authentication

Authentication happens entirely in the browser. There is no login server, session
token, or API proxy.

- **Standalone demo**: a credential form calls the WASM's `setAuthCredentials`, which
  holds your access key and secret key in memory for the lifetime of the page. The WASM
  uses them to sign requests to the Megaport API directly.
- **Portal integration**: the host page passes an existing session token via the WASM's
  `setAuthToken`, so the user doesn't re-enter credentials.

Credentials and tokens live only in browser memory and are cleared on reload or when the
page calls `clearAuthCredentials`.

## Building

The browser CLI is two pieces: the WASM binary and the Vue front end that hosts it.

```bash
# WASM binary only (writes web/megaport.wasm)
make wasm

# Full static site: WASM + Vue front end, assembled into web/vue-demo/
make web-static          # or: ./scripts/build-web.sh
```

`web-static` needs the Go toolchain and Node/npm on `PATH`. It produces a self-contained
**`web/vue-demo/`** directory (Vue build + `megaport.wasm` + `wasm_exec.js`) ready to
publish to a CDN. See [`web/README.md`](web/README.md) for the wasm pre-compression and
cache-header details.

## Local Development

The front end lives in `frontend-integration/` and has its own Vite dev server. The dev
server serves files from that directory's root and has no `public/` dir, so build the wasm
and copy the loader into `frontend-integration/` first, then start the server:

```bash
# From the repo root: build the wasm + loader into the dev server's root.
GOOS=js GOARCH=wasm go build -tags js,wasm -o frontend-integration/megaport.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" frontend-integration/wasm_exec.js

cd frontend-integration
npm install
npm run dev:demo
```

Vite serves `megaport.wasm` with the correct `application/wasm` MIME type and reloads the
front end on change. Rerun the build command above after changing Go code.

## Publishing to the Portal

The Portal loads the WASM binary from `s3://media.megaport.com/portal/megaport-cli/`. Publishing is currently manual.

### Prerequisites

- AWS CLI configured with SSO for the `ProductionDeveloper` role.

### Steps

```bash
# 1. Ensure AWS SSO auth is active (login if needed)
aws sso login
aws sts get-caller-identity
# 2. Build the WASM binary and copy the matching wasm_exec.js loader. The loader is
#    not checked in: copy it fresh from the toolchain so it stays paired with the Go
#    version that built the wasm (./wasm.sh does both steps if you prefer).
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js

# 3. Upload the WASM binary and the wasm_exec.js loader.
#    `--content-type application/wasm` is required so the file isn't served as
#    application/octet-stream: `WebAssembly.instantiateStreaming` rejects anything
#    else, which would break loading the WASM in the browser.
aws s3 cp web/megaport.wasm s3://media.megaport.com/portal/megaport-cli/megaport.wasm \
    --content-type application/wasm
aws s3 cp web/wasm_exec.js  s3://media.megaport.com/portal/megaport-cli/wasm_exec.js
```

For CDN hosting (S3 + CloudFront) of the full static site, sync the assembled
`web/vue-demo/` directory and serve it from the site root:

```bash
make web-static
aws s3 sync web/vue-demo/ s3://<bucket>/<prefix>/ --delete
```

`--delete` prunes stale hashed assets from old builds, so point it at a prefix dedicated
to this site, since it removes anything else under that prefix.

## Troubleshooting

### Can't build WASM

```bash
# Make sure you have Go 1.25 or later (matches go.mod)
go version

# Check WASM support
GOOS=js GOARCH=wasm go version
```

### Can't connect to Megaport API

The browser calls the Megaport API directly, so debug from the browser:

- open the dev tools Network tab and inspect the failing request
- confirm the access key / secret key are correct for the selected environment

## License

See LICENSE file for details.
