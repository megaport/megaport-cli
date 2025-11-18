/**
 * Web Worker for Megaport CLI WASM
 * Runs WASM module off the main thread for better performance
 */

/// <reference lib="webworker" />
declare const self: DedicatedWorkerGlobalScope;

// Worker state
let wasmReady = false;
let go: any = null;

// Message handler
self.addEventListener('message', async (e: MessageEvent) => {
  const {
    type,
    id,
    command,
    wasmPath,
    wasmExecPath,
    accessKey,
    secretKey,
    environment,
  } = e.data;

  try {
    switch (type) {
      case 'INIT':
        await initializeWASM(wasmPath, wasmExecPath);
        self.postMessage({ type: 'READY', id });
        break;

      case 'SET_AUTH':
        setAuthentication(accessKey, secretKey, environment);
        self.postMessage({ type: 'AUTH_SET', id });
        break;

      case 'EXECUTE':
        await executeCommand(command, id);
        break;

      case 'RESET':
        resetBuffers();
        self.postMessage({ type: 'RESET_COMPLETE', id });
        break;

      default:
        self.postMessage({
          type: 'ERROR',
          id,
          error: `Unknown message type: ${type}`,
        });
    }
  } catch (error) {
    self.postMessage({
      type: 'ERROR',
      id,
      error: error instanceof Error ? error.message : String(error),
    });
  }
});

/**
 * Initialize WASM module in worker
 */
async function initializeWASM(
  wasmPath: string,
  wasmExecPath: string
): Promise<void> {
  if (wasmReady) {
    console.log('WASM already initialized');
    return;
  }

  try {
    console.log('🚀 Worker: Loading wasm_exec.js from', wasmExecPath);

    // Load wasm_exec.js
    self.importScripts(wasmExecPath); // Initialize Go runtime
    go = new (self as any).Go();

    console.log('🚀 Worker: Fetching WASM from', wasmPath);

    // Fetch and instantiate WASM
    const response = await fetch(wasmPath);
    const buffer = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(buffer, go.importObject);

    console.log('🚀 Worker: Running Go WASM instance');

    // Run the Go program
    go.run(result.instance);

    // Wait for initialization
    await new Promise((resolve) => setTimeout(resolve, 200));

    wasmReady = true;
    console.log('✅ Worker: WASM initialized successfully');
  } catch (error) {
    console.error('❌ Worker: WASM initialization failed:', error);
    throw error;
  }
}

/**
 * Set authentication credentials securely (in-memory only)
 * Uses the secure setAuthCredentials function - no localStorage
 */
function setAuthentication(
  accessKey: string,
  secretKey: string,
  environment: string
): void {
  // Use the secure setAuthCredentials function
  if ((self as any).setAuthCredentials) {
    const result = (self as any).setAuthCredentials(
      accessKey,
      secretKey,
      environment
    );

    if (result && result.success) {
      console.log('🔑 Worker: Auth credentials set securely (in-memory)');
    } else {
      console.error('❌ Worker: Failed to set credentials:', result?.error);
      throw new Error(result?.error || 'Failed to set credentials');
    }
  } else {
    throw new Error('setAuthCredentials function not available');
  }
}

/**
 * Execute a CLI command
 */
async function executeCommand(
  command: string,
  messageId: string
): Promise<void> {
  if (!wasmReady) {
    throw new Error('WASM not ready');
  }

  console.log(`🚀 Worker: Executing command: ${command}`);

  return new Promise((resolve, reject) => {
    try {
      // Reset buffers before execution
      if ((self as any).resetWasmOutput) {
        (self as any).resetWasmOutput();
      }

      // Check if async function is available
      if ((self as any).executeMegaportCommandAsync) {
        (self as any).executeMegaportCommandAsync(command, (result: any) => {
          console.log('📦 Worker: Command result:', result);

          self.postMessage({
            type: 'RESULT',
            id: messageId,
            result,
          });

          resolve();
        });
      } else if ((self as any).executeMegaportCommand) {
        // Fallback to sync version
        const result = (self as any).executeMegaportCommand(command);
        console.log('📦 Worker: Command result (sync):', result);

        self.postMessage({
          type: 'RESULT',
          id: messageId,
          result,
        });

        resolve();
      } else {
        throw new Error('No WASM execute function available');
      }
    } catch (error) {
      console.error('❌ Worker: Command execution failed:', error);
      reject(error);
    }
  });
}

/**
 * Reset output buffers
 */
function resetBuffers(): void {
  if ((self as any).resetWasmOutput) {
    (self as any).resetWasmOutput();
  }
}

// Export type for TypeScript
export type {};
