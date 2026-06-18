#!/usr/bin/env node
// Browser-fetch-transport smoke check for the WASM build.
//
// Loads the compiled megaport.wasm under Node's WebAssembly runtime and global
// fetch, then runs a read-only command end-to-end against a live API. This is
// the same WasmHTTPTransport the browser uses, so it proves a command actually
// round-trips through fetch, not just that the binary links.
//
// Defaults to `locations list` against staging, which uses an unauthenticated
// client and needs no credentials. Override via env vars:
//   MEGAPORT_ENVIRONMENT  target environment (default: staging)
//   WASM_SMOKE_COMMAND    command to run (default: locations list --output json)
//   WASM_BINARY           path to megaport.wasm (default: web/megaport.wasm)
//   WASM_EXEC_JS          path to wasm_exec.js (default: web/wasm_exec.js)

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, '..');

const wasmPath = process.env.WASM_BINARY || path.join(repoRoot, 'web', 'megaport.wasm');
const wasmExecPath = process.env.WASM_EXEC_JS || path.join(repoRoot, 'web', 'wasm_exec.js');
const environment = process.env.MEGAPORT_ENVIRONMENT || 'staging';
const command = process.env.WASM_SMOKE_COMMAND || 'locations list --output json';

function fail(message) {
  console.error(`wasm-smoke: FAIL: ${message}`);
  process.exit(1);
}

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

for (const [label, p] of [['wasm binary', wasmPath], ['wasm_exec.js', wasmExecPath]]) {
  if (!fs.existsSync(p)) fail(`${label} not found at ${p} (run 'make wasm' first)`);
}

// wasm_exec.js is a plain script that assigns globalThis.Go; execute it to load.
new Function(fs.readFileSync(wasmExecPath, 'utf8'))();
if (typeof globalThis.Go !== 'function') fail('wasm_exec.js did not define globalThis.Go');

const go = new globalThis.Go();
// os.Getenv inside the Go wasm process reads from go.env, not the host's env.
// Pass only the keys the WASM module actually reads rather than the full host env.
go.env = {
  MEGAPORT_ENVIRONMENT: environment,
  MEGAPORT_ACCESS_KEY: process.env.MEGAPORT_ACCESS_KEY || '',
  MEGAPORT_SECRET_KEY: process.env.MEGAPORT_SECRET_KEY || '',
};

const { instance } = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);

// main() blocks on a channel and never returns, so start it without awaiting.
go.run(instance);

// Wait for main() to register the JS bridge before calling into it.
// WASM startup typically completes in <500ms; 5s (100 x 50ms) is a generous ceiling.
let ready = false;
for (let i = 0; i < 100 && !ready; i++) {
  ready = typeof globalThis.executeMegaportCommandAsync === 'function';
  if (!ready) await sleep(50);
}
if (!ready) fail('executeMegaportCommandAsync was not registered by the WASM module');

function runCommand(cmd, timeoutMs = 60000) {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error(`timed out after ${timeoutMs}ms`)), timeoutMs);
    globalThis.executeMegaportCommandAsync(cmd, (result) => {
      clearTimeout(timer);
      resolve(result || {});
    });
  });
}

console.log(`wasm-smoke: running "${command}" against ${environment} via fetch transport`);

let result;
try {
  result = await runCommand(command);
} catch (e) {
  fail(e.message);
}

if (result.error) fail(`command returned an error: ${result.error}`);

const output = (result.output || '').trim();
if (!output) fail('command produced no output');

let parsed;
try {
  parsed = JSON.parse(output);
} catch (e) {
  fail(`output was not valid JSON: ${e.message}\n--- output (first 500 chars) ---\n${output.slice(0, 500)}`);
}

if (!Array.isArray(parsed) || parsed.length === 0) {
  fail(`expected a non-empty JSON array, got: ${output.slice(0, 200)}`);
}

console.log(`wasm-smoke: OK, round-tripped ${parsed.length} records through the fetch transport`);
process.exit(0);
