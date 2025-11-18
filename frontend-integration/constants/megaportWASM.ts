/**
 * Megaport WASM Configuration Constants
 * Centralized configuration for WASM initialization and terminal behavior
 */

/**
 * WASM Initialization Configuration
 */
export const WASM_CONFIG = {
  /** Maximum time to wait for WASM initialization (milliseconds) */
  INIT_TIMEOUT: 30000,
  
  /** Delay after initialization to allow WASM to stabilize (milliseconds) */
  INIT_STABILIZATION_DELAY: 100,
  
  /** Maximum number of retry attempts for failed operations */
  MAX_RETRIES: 3,
  
  /** Initial delay between retry attempts (milliseconds) */
  RETRY_DELAY: 1000,
  
  /** Interval for checking WASM ready state (milliseconds) */
  READY_CHECK_INTERVAL: 100,
  
  /** Maximum time to wait for WASM ready state (milliseconds) */
  READY_TIMEOUT: 30000,
} as const;

/**
 * Terminal Configuration
 */
export const TERMINAL_CONFIG = {
  /** Font size for terminal display (pixels) */
  FONT_SIZE: 14,
  
  /** Font family for terminal display */
  FONT_FAMILY: 'Menlo, Monaco, "Courier New", monospace',
  
  /** Delay for debouncing terminal resize events (milliseconds) */
  RESIZE_DEBOUNCE_DELAY: 150,
  
  /** Maximum number of commands to keep in history */
  MAX_HISTORY_SIZE: 100,
} as const;

/**
 * Type exports for better IDE support
 */
export type WASMConfig = typeof WASM_CONFIG;
export type TerminalConfig = typeof TERMINAL_CONFIG;
