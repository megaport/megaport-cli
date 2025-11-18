/**
 * Vue 3 Composable for Megaport CLI WASM Integration
 * Handles WASM loading, initialization, and command execution
 */

import { ref, onMounted, readonly } from 'vue';
import type { Ref } from 'vue';

interface MegaportCommandResult {
  output?: string;
  error?: string;
}

interface MegaportWASMConfig {
  wasmPath?: string;
  wasmExecPath?: string;
  debug?: boolean;
  useWorker?: boolean;
}

export function useMegaportWASM(config: MegaportWASMConfig = {}) {
  const {
    wasmPath = '/megaport.wasm',
    wasmExecPath = '/wasm_exec.js',
    debug = false,
    useWorker = true,
  } = config;

  // State
  const isLoading: Ref<boolean> = ref(true);
  const isReady: Ref<boolean> = ref(false);
  const error: Ref<Error | null> = ref(null);
  const worker: Ref<Worker | null> = ref(null);
  const activeSpinners: Ref<Map<string, string>> = ref(new Map());

  // Counter for unique spinner IDs
  let spinnerCounter = 0;

  /**
   * Load the wasm_exec.js script
   */
  const loadWasmExec = (): Promise<void> => {
    return new Promise((resolve, reject) => {
      if (window.Go) {
        resolve();
        return;
      }

      const script = document.createElement('script');
      script.src = wasmExecPath;
      script.onload = () => resolve();
      script.onerror = () => reject(new Error('Failed to load wasm_exec.js'));
      document.head.appendChild(script);
    });
  };

  /**
   * Setup global spinner functions for WASM
   */
  const setupSpinnerFunctions = (): void => {
    // Global spinner start function
    (window as any).wasmStartSpinner = (message: string): string => {
      const spinnerId = `spinner_${Date.now()}_${spinnerCounter++}`;
      activeSpinners.value.set(spinnerId, message);

      if (debug) {
        console.log(`🔄 Spinner started: ${spinnerId} - ${message}`);
      }

      return spinnerId;
    };

    // Global spinner stop function
    (window as any).wasmStopSpinner = (spinnerId: string): void => {
      const message = activeSpinners.value.get(spinnerId);
      activeSpinners.value.delete(spinnerId);

      if (debug && message) {
        console.log(`⏹️ Spinner stopped: ${spinnerId} - ${message}`);
      }
    };

    if (debug) {
      console.log('✅ Spinner functions registered on window');
    }
  };

  /**
   * Initialize WASM in Web Worker (recommended for production)
   */
  const initWithWorker = async (): Promise<void> => {
    try {
      // Create worker from inline code
      const workerCode = `
        let wasmReady = false;
        let go = null;

        self.addEventListener('message', async (e) => {
          const { type, command, wasmPath, wasmExecPath } = e.data;

          if (type === 'INIT') {
            try {
              // Load wasm_exec.js in worker
              importScripts(wasmExecPath);
              
              // Initialize Go runtime
              go = new Go();
              
              // Fetch and instantiate WASM
              const result = await WebAssembly.instantiateStreaming(
                fetch(wasmPath),
                go.importObject
              );
              
              // Run the Go program
              go.run(result.instance);
              wasmReady = true;
              
              self.postMessage({ type: 'READY' });
            } catch (err) {
              self.postMessage({ 
                type: 'ERROR', 
                error: err.message 
              });
            }
          } else if (type === 'EXECUTE') {
            if (!wasmReady) {
              self.postMessage({ 
                type: 'RESULT', 
                result: { error: 'WASM not ready' } 
              });
              return;
            }

            // Since we can't easily expose window.executeMegaportCommandAsync in worker,
            // we need to use a different approach or use direct WASM without worker
            self.postMessage({ 
              type: 'RESULT', 
              result: { error: 'Worker execution not yet implemented. Use direct mode.' } 
            });
          }
        });
      `;

      const blob = new Blob([workerCode], { type: 'application/javascript' });
      const workerUrl = URL.createObjectURL(blob);
      worker.value = new Worker(workerUrl);

      // Set up message handler
      worker.value.addEventListener('message', (e: MessageEvent) => {
        const { type, error: workerError } = e.data;

        if (type === 'READY') {
          isReady.value = true;
          isLoading.value = false;
          if (debug) console.log('✅ Megaport WASM Worker ready');
        } else if (type === 'ERROR') {
          error.value = new Error(workerError);
          isLoading.value = false;
          console.error('❌ WASM Worker error:', workerError);
        }
      });

      // Initialize worker
      worker.value.postMessage({
        type: 'INIT',
        wasmPath,
        wasmExecPath,
      });
    } catch (err) {
      error.value = err as Error;
      isLoading.value = false;
      throw err;
    }
  };

  /**
   * Initialize WASM directly in main thread
   * Better for development and simpler integration
   */
  const initDirect = async (): Promise<void> => {
    try {
      // Setup spinner functions first
      setupSpinnerFunctions();

      // Load wasm_exec.js
      await loadWasmExec();

      if (!window.Go) {
        throw new Error('Go WASM runtime not loaded');
      }

      // Initialize Go runtime
      const go = new window.Go();

      // Fetch and instantiate WASM
      const response = await fetch(wasmPath);
      const buffer = await response.arrayBuffer();
      const result = await WebAssembly.instantiate(buffer, go.importObject);

      // Run the Go program
      go.run(result.instance);

      // Wait a bit for initialization
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Verify functions are available
      if (!window.executeMegaportCommandAsync) {
        throw new Error('WASM functions not exposed');
      }

      // Register prompt handler for interactive mode
      if (window.registerPromptHandler) {
        window.registerPromptHandler((promptRequest: any) => {
          if (debug) {
            console.log('📝 Prompt requested:', promptRequest);
          }

          // Note: The default handler does nothing - applications MUST register
          // their own prompt handler for interactive mode to work properly.
          // This prevents unwanted browser prompt() dialogs.
          // See MegaportTerminal.vue for an example of inline terminal prompts.
          console.warn(
            '⚠️ No custom prompt handler registered. Interactive commands require ' +
              'a prompt handler. Use registerPromptHandler() to provide one.'
          );
        });

        if (debug) {
          console.log(
            '✅ Default prompt handler registered (does nothing - override required)'
          );
        }
      }

      isReady.value = true;
      isLoading.value = false;

      if (debug) {
        console.log('✅ Megaport WASM ready (direct mode)');
        console.log('Available functions:', {
          executeMegaportCommand: typeof window.executeMegaportCommand,
          executeMegaportCommandAsync:
            typeof window.executeMegaportCommandAsync,
          debugAuthInfo: typeof window.debugAuthInfo,
        });
      }
    } catch (err) {
      error.value = err as Error;
      isLoading.value = false;
      throw err;
    }
  };

  /**
   * Execute a CLI command
   */
  const execute = async (command: string): Promise<MegaportCommandResult> => {
    if (!isReady.value) {
      throw new Error('WASM not ready');
    }

    return new Promise((resolve, reject) => {
      try {
        if (debug) {
          console.log(`🚀 Executing command: ${command}`);
        }

        // Use async version for better reliability
        if (window.executeMegaportCommandAsync) {
          window.executeMegaportCommandAsync(command, (result) => {
            if (debug) {
              console.log('📦 Command result:', result);
            }
            resolve(result);
          });
        } else if (window.executeMegaportCommand) {
          // Fallback to sync version
          const result = window.executeMegaportCommand(command);
          resolve(result);
        } else {
          reject(new Error('No WASM execute function available'));
        }
      } catch (err) {
        reject(err);
      }
    });
  };

  /**
   * Set authentication credentials (secure, in-memory only)
   *
   * Available WASM functions:
   * - executeMegaportCommand: 'function' - Synchronous command execution
   * - executeMegaportCommandAsync: 'function' - Asynchronous command execution with callback
   * - debugAuthInfo: 'function' - Get current auth state for debugging
   * - setAuthCredentials: 'function' - Secure in-memory credential storage
   * - clearAuthCredentials: 'function' - Clear credentials from memory
   *
   * This function uses setAuthCredentials to store credentials securely in-memory.
   * Credentials are NOT persisted and will be cleared on page refresh.
   * This prevents XSS attacks that could steal credentials from localStorage.
   */
  const setAuth = (
    accessKey: string,
    secretKey: string,
    environment = 'staging'
  ): void => {
    // Use secure in-memory credential storage
    if (window.setAuthCredentials) {
      const result = window.setAuthCredentials(
        accessKey,
        secretKey,
        environment
      );

      if (debug) {
        console.log('🔑 Auth credentials set securely (in-memory only)');
        if (result && !result.success) {
          console.error('Failed to set credentials:', result.error);
        }
        if (window.debugAuthInfo) {
          console.log('Auth info:', window.debugAuthInfo());
        }
      }
    } else {
      console.error(
        '❌ setAuthCredentials function not available. WASM may not be initialized.'
      );
    }
  };

  /**
   * Clear authentication credentials from memory
   */
  const clearAuth = (): void => {
    if (window.clearAuthCredentials) {
      window.clearAuthCredentials();

      if (debug) {
        console.log('🔓 Auth credentials cleared from memory');
      }
    } else {
      console.error(
        '❌ clearAuthCredentials function not available. WASM may not be initialized.'
      );
    }
  };

  /**
   * Get authentication status
   */
  const getAuthInfo = () => {
    if (window.debugAuthInfo) {
      return window.debugAuthInfo();
    }
    return null;
  };

  /**
   * Reset output buffers
   */
  const resetOutput = (): void => {
    if (window.resetWasmOutput) {
      window.resetWasmOutput();
    }
  };

  /**
   * Toggle debug mode
   */
  const toggleDebug = (): boolean => {
    if (window.toggleWasmDebug) {
      return window.toggleWasmDebug();
    }
    return false;
  };

  /**
   * Register a custom prompt handler for interactive commands
   * This allows applications to provide their own UI for prompts
   * instead of using the default browser prompt()
   *
   * @param callback - Function to handle prompt requests
   * @returns true if registered successfully
   *
   * @example
   * ```typescript
   * registerPromptHandler((request) => {
   *   // Show custom UI for the prompt
   *   showCustomPrompt(request.message).then(response => {
   *     if (response) {
   *       window.submitPromptResponse(request.id, response);
   *     } else {
   *       window.cancelPrompt(request.id);
   *     }
   *   });
   * });
   * ```
   */
  const registerPromptHandler = (callback: (request: any) => void): boolean => {
    if (window.registerPromptHandler) {
      return window.registerPromptHandler(callback);
    }
    console.warn(
      'registerPromptHandler not available - WASM may not be initialized'
    );
    return false;
  };

  // Initialize on mount
  onMounted(async () => {
    try {
      if (useWorker) {
        console.warn(
          '⚠️ Worker mode not fully implemented. Falling back to direct mode.'
        );
        await initDirect();
      } else {
        await initDirect();
      }
    } catch (err) {
      console.error('Failed to initialize Megaport WASM:', err);
      error.value = err as Error;
    }
  });

  return {
    // State (readonly refs)
    isLoading: readonly(isLoading),
    isReady: readonly(isReady),
    error: readonly(error),
    activeSpinners: readonly(activeSpinners),

    // Methods
    execute,
    setAuth,
    clearAuth,
    getAuthInfo,
    resetOutput,
    toggleDebug,
    registerPromptHandler,
  };
}
