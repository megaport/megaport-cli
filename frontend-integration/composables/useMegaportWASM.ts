/**
 * Vue 3 Composable for Megaport CLI WASM Integration
 * Handles WASM loading, initialization, and command execution
 */

import { ref, onMounted, onUnmounted, readonly, triggerRef } from 'vue';
import type { Ref } from 'vue';
import {
  isMegaportCommandResult,
  isMegaportPromptRequest,
  hasWASMFunctions,
  hasWebAssemblySupport,
  isValidCommand,
  getErrorMessage,
} from '../utils/type-guards';

interface MegaportCommandResult {
  output?: string;
  error?: string;
}

interface MegaportWASMConfig {
  wasmPath?: string;
  wasmExecPath?: string;
  debug?: boolean;
  initTimeout?: number; // Timeout for WASM initialization in ms
  maxRetries?: number; // Maximum retry attempts for failed initialization
  retryDelay?: number; // Base delay between retries in ms
  onTelemetry?: (
    event: import('../types/megaport-wasm').TelemetryEvent
  ) => void; // Telemetry callback
}

// Constants
const DEFAULT_INIT_TIMEOUT = 30000; // 30 seconds
const INIT_STABILIZATION_DELAY = 100; // Wait for WASM to stabilize
const DEFAULT_MAX_RETRIES = 3; // Retry up to 3 times
const DEFAULT_RETRY_DELAY = 1000; // Start with 1 second delay

export function useMegaportWASM(config: MegaportWASMConfig = {}) {
  const {
    wasmPath = '/megaport.wasm',
    wasmExecPath = '/wasm_exec.js',
    debug = false,
    initTimeout = DEFAULT_INIT_TIMEOUT,
    maxRetries = DEFAULT_MAX_RETRIES,
    retryDelay = DEFAULT_RETRY_DELAY,
    onTelemetry,
  } = config;

  // State
  const isLoading: Ref<boolean> = ref(true);
  const isReady: Ref<boolean> = ref(false);
  const error: Ref<Error | null> = ref(null);
  const activeSpinners: Ref<Map<string, string>> = ref(new Map());

  // Counter for unique spinner IDs
  let spinnerCounter = 0;

  /**
   * Conditional logging helper - only logs when debug mode is enabled
   */
  const log = (message: string, ...args: any[]) => {
    if (debug) {
      console.log(message, ...args);
    }
  };

  const warn = (message: string, ...args: any[]) => {
    // Warnings should always be shown, regardless of debug mode
    console.warn(message, ...args);
  };

  /**
   * Emit telemetry event if callback is provided
   */
  const emitTelemetry = (
    type: import('../types/megaport-wasm').TelemetryEventType,
    metadata?: Record<string, any>,
    duration?: number
  ) => {
    if (onTelemetry) {
      onTelemetry({
        type,
        timestamp: Date.now(),
        duration,
        metadata,
      });
    }
  };

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
      // Mutate the Map directly and trigger reactivity manually
      activeSpinners.value.set(spinnerId, message);
      triggerRef(activeSpinners);

      log(`🔄 Spinner started: ${spinnerId} - ${message}`);
      emitTelemetry('spinner_start', { spinnerId, message });

      return spinnerId;
    };

    // Global spinner stop function
    (window as any).wasmStopSpinner = (spinnerId: string): void => {
      const message = activeSpinners.value.get(spinnerId);
      // Mutate the Map directly and trigger reactivity manually
      activeSpinners.value.delete(spinnerId);
      triggerRef(activeSpinners);

      if (message) {
        log(`⏹️ Spinner stopped: ${spinnerId} - ${message}`);
      }
      emitTelemetry('spinner_stop', { spinnerId, message });
    };

    log('✅ Spinner functions registered on window');
  };

  /**
   * Initialize WASM directly in main thread with timeout
   * Better for development and simpler integration
   */
  const initDirect = async (): Promise<void> => {
    const startTime = Date.now();
    emitTelemetry('wasm_init_start', { mode: 'direct' });

    // Verify WebAssembly support
    if (!hasWebAssemblySupport()) {
      const err = new Error('WebAssembly is not supported in this browser');
      error.value = err;
      isLoading.value = false;
      emitTelemetry(
        'wasm_init_error',
        {
          mode: 'direct',
          error: err.message,
        },
        Date.now() - startTime
      );
      throw err;
    }

    // Create timeout promise
    const timeoutPromise = new Promise<never>((_, reject) => {
      setTimeout(() => {
        reject(new Error(`WASM initialization timeout after ${initTimeout}ms`));
      }, initTimeout);
    });

    // Create initialization promise
    const initPromise = async () => {
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
        await new Promise((resolve) =>
          setTimeout(resolve, INIT_STABILIZATION_DELAY)
        );

        // Verify functions are available
        if (!hasWASMFunctions(window)) {
          throw new Error('WASM functions not exposed');
        }

        // Register prompt handler for interactive mode
        if (window.registerPromptHandler) {
          window.registerPromptHandler((promptRequest: any) => {
            // Validate prompt request with type guard
            if (!isMegaportPromptRequest(promptRequest)) {
              warn('⚠️ Invalid prompt request received:', promptRequest);
              return;
            }

            log('📝 Prompt requested:', promptRequest);

            // Note: The default handler does nothing - applications MUST register
            // their own prompt handler for interactive mode to work properly.
            // This prevents unwanted browser prompt() dialogs.
            // See MegaportTerminal.vue for an example of inline terminal prompts.
            warn(
              '⚠️ No custom prompt handler registered. Interactive commands require ' +
                'a prompt handler. Use registerPromptHandler() to provide one.'
            );
          });

          log(
            '✅ Default prompt handler registered (does nothing - override required)'
          );
        }

        isReady.value = true;
        isLoading.value = false;

        const duration = Date.now() - startTime;
        emitTelemetry('wasm_init_success', { mode: 'direct' }, duration);

        log('✅ Megaport WASM ready (direct mode)');
        log('Available functions:', {
          executeMegaportCommand: typeof window.executeMegaportCommand,
          executeMegaportCommandAsync:
            typeof window.executeMegaportCommandAsync,
          debugAuthInfo: typeof window.debugAuthInfo,
        });
      } catch (err) {
        error.value = err as Error;
        isLoading.value = false;
        const duration = Date.now() - startTime;
        emitTelemetry(
          'wasm_init_error',
          {
            mode: 'direct',
            error: (err as Error).message,
          },
          duration
        );
        throw err;
      }
    };

    // Race initialization against timeout
    try {
      await Promise.race([initPromise(), timeoutPromise]);
    } catch (err) {
      error.value = err as Error;
      isLoading.value = false;
      const duration = Date.now() - startTime;
      emitTelemetry(
        'wasm_init_error',
        {
          mode: 'direct',
          error: (err as Error).message,
        },
        duration
      );
      throw err;
    }
  };

  /**
   * Execute a CLI command
   */
  const execute = async (command: string): Promise<MegaportCommandResult> => {
    // Validate command with type guard
    if (!isValidCommand(command)) {
      const error =
        'Invalid command: must be a non-empty string without dangerous patterns';
      emitTelemetry('command_execute_error', { command, error }, 0);
      throw new Error(error);
    }

    if (!isReady.value) {
      throw new Error('WASM not ready');
    }

    const startTime = Date.now();
    emitTelemetry('command_execute_start', { command });

    return new Promise((resolve, reject) => {
      try {
        log(`🚀 Executing command: ${command}`);

        // Use async version for better reliability
        if (window.executeMegaportCommandAsync) {
          window.executeMegaportCommandAsync(command, (result) => {
            const duration = Date.now() - startTime;

            // Validate result with type guard
            if (!isMegaportCommandResult(result)) {
              const error = 'Invalid command result received from WASM';
              warn('⚠️ Invalid result:', result);
              emitTelemetry(
                'command_execute_error',
                { command, error },
                duration
              );
              reject(new Error(error));
              return;
            }

            log('📦 Command result:', result);

            if (result.error) {
              emitTelemetry(
                'command_execute_error',
                {
                  command,
                  error: result.error,
                },
                duration
              );
            } else {
              emitTelemetry('command_execute_success', { command }, duration);
            }

            resolve(result);
          });
        } else if (window.executeMegaportCommand) {
          // Fallback to sync version
          const result = window.executeMegaportCommand(command);
          const duration = Date.now() - startTime;

          // Validate result with type guard
          if (!isMegaportCommandResult(result)) {
            const error = 'Invalid command result received from WASM';
            warn('⚠️ Invalid result:', result);
            emitTelemetry(
              'command_execute_error',
              { command, error },
              duration
            );
            reject(new Error(error));
            return;
          }

          if (result.error) {
            emitTelemetry(
              'command_execute_error',
              {
                command,
                error: result.error,
              },
              duration
            );
          } else {
            emitTelemetry('command_execute_success', { command }, duration);
          }

          resolve(result);
        } else {
          const duration = Date.now() - startTime;
          emitTelemetry(
            'command_execute_error',
            {
              command,
              error: 'No WASM execute function available',
            },
            duration
          );
          reject(new Error('No WASM execute function available'));
        }
      } catch (err) {
        const duration = Date.now() - startTime;
        emitTelemetry(
          'command_execute_error',
          {
            command,
            error: getErrorMessage(err),
          },
          duration
        );
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

      log('🔑 Auth credentials set securely (in-memory only)');
      emitTelemetry('auth_set', { environment, success: result?.success });

      if (result && !result.success) {
        console.error('Failed to set credentials:', result.error); // Always log errors
      }
      if (window.debugAuthInfo) {
        log('Auth info:', window.debugAuthInfo());
      }
    } else {
      console.error(
        '❌ setAuthCredentials function not available. WASM may not be initialized.'
      ); // Always log errors
      emitTelemetry('auth_set', { environment, success: false });
    }
  };

  /**
   * Clear authentication credentials from memory
   */
  const clearAuth = (): void => {
    if (window.clearAuthCredentials) {
      window.clearAuthCredentials();
      log('🔓 Auth credentials cleared from memory');
      emitTelemetry('auth_clear', {});
    } else {
      console.error(
        '❌ clearAuthCredentials function not available. WASM may not be initialized.'
      ); // Always log errors
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
    warn('registerPromptHandler not available - WASM may not be initialized');
    return false;
  };

  /**
   * Initialize WASM with retry logic
   * Attempts initialization multiple times with exponential backoff
   */
  const initWithRetry = async (): Promise<void> => {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        log(`Attempt ${attempt}/${maxRetries}: Initializing Megaport WASM...`);

        await initDirect();

        log(`✅ WASM initialized successfully on attempt ${attempt}`);
        return; // Success!
      } catch (err) {
        lastError = err as Error;
        console.error(
          `❌ WASM initialization attempt ${attempt}/${maxRetries} failed:`,
          err
        );

        if (attempt < maxRetries) {
          // Calculate exponential backoff delay
          const delay = retryDelay * Math.pow(2, attempt - 1);
          log(`Retrying in ${delay}ms...`);

          await new Promise((resolve) => setTimeout(resolve, delay));
        }
      }
    }

    // All retries failed
    const finalError = new Error(
      `WASM initialization failed after ${maxRetries} attempts. Last error: ${lastError?.message}`
    );
    error.value = finalError;
    isLoading.value = false;
    throw finalError;
  };

  // Initialize on mount
  onMounted(async () => {
    try {
      await initWithRetry();
    } catch (err) {
      console.error(
        'Failed to initialize Megaport WASM after all retries:',
        err
      ); // Always log final failure
      error.value = err as Error;
    }
  });

  /**
   * Cleanup function - clears state and auth
   */
  const cleanup = () => {
    log('🧹 Cleaning up WASM resources');

    // Clear active spinners
    activeSpinners.value.clear();

    // Clear auth credentials from memory for security
    if (window.clearAuthCredentials) {
      window.clearAuthCredentials();
    }

    // Remove global spinner functions
    if (typeof window !== 'undefined') {
      delete (window as any).wasmStartSpinner;
      delete (window as any).wasmStopSpinner;
    }

    log('Cleanup complete');
  };

  // Cleanup on unmount
  onUnmounted(() => {
    cleanup();
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
    cleanup, // Expose cleanup for manual cleanup if needed
  };
}
