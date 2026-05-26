# Megaport CLI - WebAssembly (WASM) Browser Terminal

A browser-based terminal for the Megaport CLI powered by **WebAssembly (WASM)**, deployed with Docker. This enables customers to use the Megaport CLI directly in their web browser without installing anything locally.

## What is This?

This is the **WASM browser version** of the Megaport CLI that:

- **Runs in the browser** - No local installation required
- **Powered by WebAssembly** - Go code compiled to WASM runs directly in the browser
- **Deployed with Docker** - Easy deployment with a containerized web server
- **Session-based authentication** - Secure login using customer's Megaport credentials
- **XTerm.js Terminal** - Full-featured terminal emulator with ANSI support
- **Early Release** - Currently supports locations and ports commands (more coming soon!)

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

### Available Commands (Current)

⚠️ **Note**: This is a very early release of the WASM version. Currently supported commands:

- `locations list` - List all Megaport locations
- `locations get <locationId>` - Get details for a specific location
- `ports list` - List your ports
- `ports get <portId>` - Get details for a specific port

**Output Formats**: All standard output formats are supported:

- `--output table` (default) - Formatted table with styled output
- `--output json` - JSON format for programmatic use
- `--output csv` - CSV format for data export

**Example Commands**:

```bash
megaport-cli locations list
megaport-cli locations list --output json
megaport-cli ports list
megaport-cli ports get abc123
```

More commands will be added in future releases!

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
# Make sure you have Go 1.21 or later
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
