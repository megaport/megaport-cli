# Megaport CLI - WebAssembly (WASM) Browser Terminal

A browser-based terminal for the Megaport CLI powered by **WebAssembly (WASM)**, deployed with Docker. This enables customers to use the Megaport CLI directly in their web browser without installing anything locally.

## What is This?

This is the **WASM browser version** of the Megaport CLI that:

- **Runs in the browser** - No local installation required
- **Powered by WebAssembly** - Go code compiled to WASM runs directly in the browser
- **Deployed with Docker** - Easy deployment with a containerized web server
- **Session-based authentication** - Secure login using customer's Megaport credentials
- **XTerm.js Terminal** - Full-featured terminal emulator with ANSI support
- **Early Release** - Covers all registered resource modules (see Available Commands below). Auth, config, completion, `generate-docs`, and `version` are not applicable in the browser.

## Quick Start (One Command!)

### Deploy with the Deployment Script

```bash
# Clone the repository
git clone https://github.com/megaport/megaport-cli.git
cd megaport-cli

# Run the deployment script
./deploy.sh
```

That's it! The script will:

1. ✅ Build the WASM binary
2. ✅ Build the Docker image
3. ✅ Stop any existing container
4. ✅ Start the new container
5. ✅ Verify it's running

**Open your browser to http://localhost:8080 and login!**

### Login to the Web Terminal

1. Open **http://localhost:8080** in your browser
2. Enter your Megaport credentials:
   - **Access Key** - Your Megaport API access key
   - **Secret Key** - Your Megaport API secret key
   - **Environment** - Select production, staging, or development
3. Click **Login**
4. Start using the CLI!

### Available Commands

The WASM build registers the following modules. Each module exposes the same subcommands it provides in the native CLI (so `partners` is still `list` / `find`, `locations` is still `list` / `get`, and so on), subject to the constraints of running in the browser:

- `locations` - Find Megaport locations and metros
- `ports` - Manage Megaport ports (including LAG ports)
- `mcr` - Manage Megaport Cloud Routers
- `mve` - Manage Megaport Virtual Edge devices
- `vxc` - Manage Virtual Cross Connects
- `nat-gateway` - Manage NAT gateways
- `partners` - Look up cloud partner ports
- `product` - Inspect provisioned products
- `servicekeys` - Manage service keys
- `status` - Check resource provisioning status
- `topology` - View network topology

The registered list is the source of truth in `cmd/megaport/modules_wasm.go`. Any module not registered there is unavailable in the WASM build, including `auth`, `config`, `completion`, `generate-docs`, `version`, `ix`, `users`, `apply`, `managed-account`, and `billing-market`. Some rely on the local filesystem (profile storage, completion scripts); others have simply not been wired into the browser build yet. Authentication in WASM is session-based via the web UI login form.

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

### Session Management

- **Session Duration**: By default, sessions last **1 hour**
- **Auto-expiration**: When your session expires, you'll be automatically redirected to the login page
- **Logout Button**: Click the **Logout** button in the top-right corner to end your session manually
- **Session Storage**: Your credentials are stored securely in memory and cleared on logout or expiration

### Managing the Container

```bash
# View logs
docker logs -f megaport-cli-wasm

# Restart
docker restart megaport-cli-wasm

# Stop
docker stop megaport-cli-wasm

# Remove
docker rm megaport-cli-wasm

# Rebuild and redeploy
./deploy.sh
```

## Static Build (CDN Hosting)

The Docker flow above runs a Go server that does more than serve files — it
also handles the login/session endpoints and proxies API calls (see
`cmd/server/server.go`). To host the browser CLI on a CDN (S3 + CloudFront)
instead, build just the static front-end assets and sync the output dir:

```bash
make web-static          # or: ./scripts/build-web.sh
```

Needs the Go toolchain and Node/npm on `PATH` — the build compiles the WASM
binary and bundles the Vue front end.

This produces a self-contained **`web/vue-demo/`** directory (Vue build +
`megaport.wasm` + `wasm_exec.js`). Publish it with:

```bash
aws s3 sync web/vue-demo/ s3://<bucket>/<prefix>/ --delete
```

`--delete` prunes stale hashed assets from old builds, so point it at a prefix
dedicated to this site — it removes anything else under that prefix.

### Notes for the CDN/S3 side

- A static deployment serves the **front-end assets only**. It does not include
  the login/session and API-proxy endpoints that the Docker server provides, so
  the auth/API path for a server-less deployment has to be handled separately
  (out of scope here — see the infra ticket).
- The build assumes the app is served from the **site root**. The demo's vite
  config doesn't set `base`, so the bundled `assets/` resolve from root. Serving
  under a path (e.g. `media.megaport.com/cli/`) needs source changes, not just
  config: setting `base` in `frontend-integration/vite.demo.config.ts` would
  rewrite the bundled `assets/`, but `megaport.wasm` and `wasm_exec.js` are
  fetched from hardcoded absolute paths (`wasm-path`/`wasm-exec-path` in
  `frontend-integration/demo/App.vue`) and would still 404. Confirm root
  hosting, or budget for those edits.
- `megaport.wasm` is ~32 MB uncompressed — serve it compressed (brotli `-q11`
  gets it to ~4.7 MB over the wire, gzip `-9` ~6.8 MB). S3 must set
  `Content-Type: application/wasm` explicitly; it won't be inferred.
- The wasm file keeps a fixed name (it isn't content-hashed like vite's
  `assets/`), so invalidate it on every deploy. Serve `index.html` `no-cache`;
  the hashed files under `assets/` can cache long/immutable.

## Configuration

### Environment Variables

| Variable           | Default | Description                                        |
| ------------------ | ------- | -------------------------------------------------- |
| `PORT`             | `8080`  | Port to expose the server on                       |
| `SESSION_DURATION` | `1h`    | How long customer sessions last (30m, 1h, 2h, 24h) |
| `TZ`               | `UTC`   | Timezone for logs                                  |

### Session Duration

The session duration determines how long a customer can stay logged in before needing to re-authenticate:

```bash
# 30 minutes (good for demos)
SESSION_DURATION=30m

# 1 hour (default, good for interactive use)
SESSION_DURATION=1h

# 8 hours (good for long work sessions)
SESSION_DURATION=8h

# 24 hours (maximum recommended)
SESSION_DURATION=24h
```

## How It Works

### Architecture Overview

```
Customer Browser                Docker Container
    │                               │
    │  1. Login with credentials    │
    ├──────────────────────────────>│
    │                               │
    │  2. Session token returned    │  3. Validates with
    │<──────────────────────────────┤     Megaport API
    │                               │
    │  4. Load WASM binary          │
    │<──────────────────────────────┤
    │                               │
    │  5. Execute CLI commands      │
    │     (WASM runs in browser)    │
    │                               │
    │  6. API calls via proxy       │
    ├──────────────────────────────>│  7. Proxy to
    │     (with session token)      │     Megaport API
    │                               │
```

### WebAssembly Execution

1. **Go CLI compiled to WASM** - The entire Megaport CLI is compiled to a `.wasm` file
2. **Runs in browser** - WASM binary executes directly in the customer's browser
3. **Commands run client-side** - All CLI logic runs in the browser, not on the server
4. **API calls proxied** - Only API requests go through the Docker server for authentication

### Authentication Flow

1. **Customer logs in** through the web UI with their Megaport credentials
2. **Docker server validates** credentials with Megaport API
3. **Server creates session** and returns a session token to the browser
4. **Browser stores token** in localStorage (credentials are NOT stored)
5. **WASM binary uses credentials** - Credentials stored in JavaScript global for WASM to access
6. **Session auto-expires** - After configured duration, redirects to login
7. **Logout clears session** - Session token and credentials removed from browser

### Important Notes

> **⚠️ Config Commands Not Available**: The WASM/browser version does **NOT** support `config` commands
> (create-profile, use-profile, etc.). These are only available in the standard CLI.
>
> Authentication in WASM is **session-based** via the web UI login form, not profile-based.

## Development

### Enabling a Module for WASM

Every module follows the same recipe to go from native-only to browser-enabled:

1. **Register it.** Add the module to `registerModules()` in
   `cmd/megaport/modules_wasm.go` (import the package and call
   `moduleRegistry.Register(<module>.NewModule())`). This file is the source of
   truth for what ships in the browser build.

2. **Build the guard.** Run `make wasm-build-guard`
   (`GOOS=js GOARCH=wasm go build -tags js,wasm -o /dev/null .`). This fails if
   the module pulls in anything that does not compile under `js/wasm` (native
   filesystem, `os/exec`, `syscall`, etc.). CI runs the same step on every PR.

3. **Add a `_wasm` override only if needed.** Most commands work unchanged
   because the API call already routes through the WASM fetch transport. A
   module needs a `<module>_actions_wasm.go` (`//go:build js && wasm`) only
   when its native path does something the browser can't, such as reading the
   local filesystem or using a custom HTTP client. The pattern is to override
   the action's function variable in `init()`, using
   `config.NewUnauthenticatedClient()` for public endpoints or
   `config.Login(ctx)` for authenticated ones. See
   `internal/commands/locations/locations_actions_wasm.go` for a minimal
   example.

4. **Smoke-test it.** Run `make wasm-smoke` to round-trip a command through the
   browser fetch transport against a live API (defaults to `locations list`
   against staging, no credentials needed). Requires Node.js 20+ on `PATH`. To
   exercise a specific module:

   ```bash
   WASM_SMOKE_COMMAND='locations list --output json' make wasm-smoke
   ```

   The command must emit JSON (`--output json`) and return a non-empty JSON
   array (use a `list` subcommand, not a single-resource `get`). For anything
   that needs auth, set `MEGAPORT_ACCESS_KEY` / `MEGAPORT_SECRET_KEY` in the
   environment first (local use only; do not add credentials to the CI smoke
   job).

### Local Development with Hot Reload

```bash
# Uncomment volume mount in docker-compose.yml
# - ./web:/app/web:ro

# Rebuild WASM locally
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .

# Container will serve the updated files immediately
```

### Building Manually

```bash
# Build WASM
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .

# Build server
go build -o server ./cmd/server/server.go

# Run locally
./server --port 8080 --dir web --session-duration 1h
```

The server binds `127.0.0.1` by default. To expose it on other interfaces (e.g. inside a container), pass `--bind 0.0.0.0`.

## Publishing to the Portal

The Portal loads the WASM binary from `s3://media.megaport.com/portal/megaport-cli/`. Publishing is currently manual.

### Prerequisites

- AWS CLI configured with SSO for the `ProductionDeveloper` role.

### Steps

```bash
# 1. Ensure AWS SSO auth is active (login if needed)
aws sso login
aws sts get-caller-identity
# 2. Build the WASM binary
GOOS=js GOARCH=wasm go build -tags js,wasm -o web/megaport.wasm .

# 3. Upload the WASM binary and the (already checked-in) wasm_exec.js loader.
#    `--content-type application/wasm` is required so the file isn't served as
#    application/octet-stream — `WebAssembly.instantiateStreaming` rejects
#    anything else, which would break the standalone loader in web/script.js.
aws s3 cp web/megaport.wasm s3://media.megaport.com/portal/megaport-cli/megaport.wasm \
    --content-type application/wasm
aws s3 cp web/wasm_exec.js  s3://media.megaport.com/portal/megaport-cli/wasm_exec.js
```

## API Endpoints

### Authentication

```bash
# Login
POST /auth/login
Content-Type: application/json
{
  "accessKey": "your-key",
  "secretKey": "your-secret",
  "environment": "production"  # or "staging", "development"
}

Response:
{
  "sessionToken": "abc123...",
  "expiresIn": 3600,
  "environment": "production"
}

# Logout
POST /auth/logout
X-Session-Token: abc123...

# Check session
GET /auth/check
X-Session-Token: abc123...
```

### Authenticated API Proxy

```bash
# All Megaport API calls go through /api/
GET /api/v2/locations
X-Session-Token: abc123...

GET /api/v2/products
X-Session-Token: abc123...
```

## Troubleshooting

### Can't build WASM

```bash
# Make sure you have Go 1.25 or later (matches go.mod)
go version

# Check WASM support
GOOS=js GOARCH=wasm go version
```

### Docker build fails

```bash
# Clean build
docker-compose build --no-cache

# Check logs
docker-compose logs
```

### Session expires immediately

```bash
# Check session duration
docker-compose exec megaport-cli-wasm env | grep SESSION

# Increase duration
# Edit .env file: SESSION_DURATION=24h
docker-compose up -d
```

### Can't connect to Megaport API

```bash
# Check network connectivity from container
docker-compose exec megaport-cli-wasm wget -O- https://api.megaport.com/

# Check logs for authentication errors
docker-compose logs | grep "Authentication failed"
```

## License

See LICENSE file for details.
