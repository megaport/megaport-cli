/**
 * Megaport CLI WebAssembly Module - TypeScript Definitions
 * For Vue 3 + Vite Integration
 */

export interface MegaportCommandResult {
  output?: string;
  error?: string;
}

export interface MegaportAuthInfo {
  accessKeySet: boolean;
  accessKeyPreview: string;
  secretKeySet: boolean;
  secretKeyPreview: string;
  accessTokenSet: boolean;
  accessTokenPreview: string;
  environment: string;
  apiURL: string;
  authMethod: 'token' | 'apikey' | 'none';
}

export interface MegaportBufferDump {
  stdout: string;
  stderr: string;
  direct: string;
}

/**
 * Prompt request from WASM for interactive commands
 */
export interface MegaportPromptRequest {
  id: string;
  message: string;
  type: string; // "text", "confirm", "resource"
  resourceType?: string; // for resource prompts
}

/**
 * Telemetry event types for tracking WASM operations
 */
export type TelemetryEventType =
  | 'wasm_init_start'
  | 'wasm_init_success'
  | 'wasm_init_error'
  | 'command_execute_start'
  | 'command_execute_success'
  | 'command_execute_error'
  | 'auth_set'
  | 'auth_clear'
  | 'spinner_start'
  | 'spinner_stop'
  | 'prompt_requested'
  | 'prompt_submitted'
  | 'prompt_cancelled';

/**
 * Telemetry event data
 */
export interface TelemetryEvent {
  type: TelemetryEventType;
  timestamp: number;
  duration?: number; // milliseconds
  metadata?: Record<string, any>;
}

/**
 * Telemetry callback function
 */
export type TelemetryCallback = (event: TelemetryEvent) => void;

/**
 * Global WASM interface exposed by the Megaport CLI
 * Available after WASM module initialization
 */
export interface MegaportWASM {
  /**
   * Execute a CLI command synchronously (LEGACY - may not work with async operations)
   * @param command - Full command string (e.g., "port list --output json")
   * @returns Result object with output or error
   * @deprecated Use executeMegaportCommandAsync for better reliability
   */
  executeMegaportCommand(command: string): MegaportCommandResult;

  /**
   * Execute a CLI command asynchronously (RECOMMENDED)
   * @param command - Full command string (e.g., "port list --output json")
   * @param callback - Callback function to receive the result
   */
  executeMegaportCommandAsync(
    command: string,
    callback: (result: MegaportCommandResult) => void
  ): void;

  /**
   * Read a config file from localStorage
   * @param filename - Name of the file to read
   */
  readConfigFile(filename: string): { content?: string; error?: string };

  /**
   * Write a config file to localStorage
   * @param filename - Name of the file to write
   * @param content - Content to write
   */
  writeConfigFile(filename: string, content: string): { success: boolean };

  /**
   * Get authentication information
   */
  debugAuthInfo(): MegaportAuthInfo;

  /**
   * Save data to localStorage
   */
  saveToLocalStorage(key: string, value: string): boolean;

  /**
   * Load data from localStorage
   */
  loadFromLocalStorage(key: string): string;

  /**
   * Set authentication credentials securely (in-memory only, recommended)
   * Stores credentials in Go environment variables and JavaScript global
   * Does NOT use localStorage to prevent XSS attacks
   * @param accessKey - Megaport API access key
   * @param secretKey - Megaport API secret key
   * @param environment - Environment (production, staging, development)
   * @returns Result object with success status
   */
  setAuthCredentials(
    accessKey: string,
    secretKey: string,
    environment: string
  ): { success: boolean; error?: string };

  /**
   * Set authentication using an existing token from the portal session,
   * bypassing the OAuth flow. Use this when the host page already holds a
   * valid Megaport access token (typically via SSO into the portal).
   *
   * ## Environment resolution
   *
   * The environment is resolved in this order:
   *
   * 1. The explicit `environment` argument, if non-empty.
   * 2. The environment derived from `hostname` per the Megaport conventions:
   *    - `megaport.com`, `www.megaport.com`, and any `<app>.megaport.com`
   *      (single-word app, no hyphens) → `"production"`.
   *    - `<app>-<env>.megaport.com` → `<env>` (env may contain further
   *      hyphens, so `api-mpone-dev.megaport.com` resolves to `"mpone-dev"`).
   *
   * If neither yields a value (e.g. `hostname` is `"localhost"`, a private IP,
   * or a non-Megaport domain), **the call fails**. The function never
   * silently falls back to production.
   *
   * ## API URL
   *
   * The API URL is always built from the resolved environment:
   * - `"production"` → `https://api.megaport.com/`.
   * - anything else → `https://api-<env>.megaport.com/`.
   *
   * ## Validation
   *
   * The explicit `environment` argument must match `/^[a-z0-9-]+$/` — any
   * other value (containing `/`, `.`, `@`, `:`, uppercase, etc.) is rejected
   * with an error to prevent hostname injection into the API URL.
   *
   * @param token - The access token from the portal session
   * @param hostname - The current hostname, e.g. `window.location.hostname`
   * @param environment - Optional explicit environment override; supersedes the hostname-derived value. Useful when `hostname` is `"localhost"` or a non-portal host, or when the portal needs to talk to a specific backend regardless of where it's served
   * @returns On success: `{ success: true, environment, hostname, apiURL }` where `environment` is the resolved env name (e.g. `"qa"`) and `apiURL` is the matching `api-<env>.megaport.com/` URL. On failure: `{ success: false, error }` with a human-readable message; the caller should surface the message to guide the user
   *
   * @example
   * // Portal served from a recognised host — no override needed.
   * setAuthToken(token, window.location.hostname);
   *
   * @example
   * // Local development against the qa backend.
   * setAuthToken(token, window.location.hostname, "qa");
   */
  setAuthToken(
    token: string,
    hostname: string,
    environment?: string
  ): { success: boolean; error?: string; environment?: string; hostname?: string; apiURL?: string };

  /**
   * Clear authentication credentials from memory
   * @returns Result object with success status
   */
  clearAuthCredentials(): { success: boolean };

  /**
   * Reset WASM output buffers
   */
  resetWasmOutput(): boolean;

  /**
   * Get current WASM output
   */
  getWasmOutput(): string;

  /**
   * Toggle WASM debug mode
   */
  toggleWasmDebug(): boolean;

  /**
   * Dump all buffer contents for debugging
   */
  dumpBuffers(): MegaportBufferDump;

  /**
   * Check if WASM debug mode is enabled
   */
  wasmDebug(): boolean;

  /**
   * Log location command debug information
   */
  logLocationCommand(message: string): void;

  /**
   * Register a prompt handler for interactive commands
   * @param callback - Function to call when user input is needed
   */
  registerPromptHandler(
    callback: (request: MegaportPromptRequest) => void
  ): boolean;

  /**
   * Submit a response to a pending prompt
   * @param id - Prompt ID
   * @param response - User's response
   */
  submitPromptResponse(id: string, response: string): void;

  /**
   * Cancel a pending prompt
   * @param id - Prompt ID
   */
  cancelPrompt(id: string): void;

  /**
   * Get list of pending prompts (for debugging)
   */
  getPendingPrompts(): MegaportPromptRequest[];
}

/**
 * Go WASM runtime
 */
export interface GoWASM {
  run(instance: WebAssembly.Instance): void;
  importObject: WebAssembly.Imports;
  _exitPromise?: Promise<void>;
  _resolveExitPromise?: () => void;
  _pendingEvent?: { id: number; this: any; args: any[] };
}

declare global {
  interface Window {
    executeMegaportCommand?: (command: string) => MegaportCommandResult;
    executeMegaportCommandAsync?: (
      command: string,
      callback: (result: MegaportCommandResult) => void
    ) => void;
    readConfigFile?: (filename: string) => { content?: string; error?: string };
    writeConfigFile?: (
      filename: string,
      content: string
    ) => { success: boolean };
    debugAuthInfo?: () => MegaportAuthInfo;
    saveToLocalStorage?: (key: string, value: string) => boolean;
    loadFromLocalStorage?: (key: string) => string;
    setAuthCredentials?: (
      accessKey: string,
      secretKey: string,
      environment: string
    ) => { success: boolean; error?: string };
    /**
     * Global shim mirroring {@link MegaportWASM.setAuthToken}. See the
     * interface JSDoc for the full contract, hostname conventions, validation
     * rules, and examples.
     */
    setAuthToken?: (
      token: string,
      hostname: string,
      environment?: string
    ) => { success: boolean; error?: string; environment?: string; hostname?: string; apiURL?: string };
    clearAuthCredentials?: () => { success: boolean };
    resetWasmOutput?: () => boolean;
    getWasmOutput?: () => string;
    toggleWasmDebug?: () => boolean;
    dumpBuffers?: () => MegaportBufferDump;
    wasmDebug?: () => boolean;
    logLocationCommand?: (message: string) => void;
    registerPromptHandler?: (
      callback: (request: MegaportPromptRequest) => void
    ) => boolean;
    submitPromptResponse?: (id: string, response: string) => void;
    cancelPrompt?: (id: string) => void;
    getPendingPrompts?: () => MegaportPromptRequest[];
    Go?: new () => GoWASM;
  }
}

export {};
