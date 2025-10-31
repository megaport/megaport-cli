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

## Development

The Charsm renderer uses the Charsm library (WebAssembly port of lipgloss) to provide styled terminal output in the browser.

To test manually in a browser:

1. Build the WASM binary: `./wasm.sh`
2. Start the server: `./start-server.sh`
3. Open http://localhost:8080 in your browser
