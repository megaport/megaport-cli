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
  environment: string;
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
