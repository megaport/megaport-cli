/**
 * Runtime type guards for WASM integration
 * Provides safe runtime checks for TypeScript types
 */

import type {
  MegaportCommandResult,
  MegaportAuthInfo,
  MegaportPromptRequest,
  TelemetryEvent,
  TelemetryEventType,
} from '../types/megaport-wasm';

/**
 * Type guard for MegaportCommandResult
 */
export function isMegaportCommandResult(
  value: unknown
): value is MegaportCommandResult {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const result = value as Partial<MegaportCommandResult>;

  // Must have at least one of output or error
  if (!('output' in result) && !('error' in result)) {
    return false;
  }

  // If present, output must be string or undefined
  if (
    'output' in result &&
    result.output !== undefined &&
    typeof result.output !== 'string'
  ) {
    return false;
  }

  // If present, error must be string or undefined
  if (
    'error' in result &&
    result.error !== undefined &&
    typeof result.error !== 'string'
  ) {
    return false;
  }

  return true;
}

/**
 * Type guard for MegaportAuthInfo
 */
export function isMegaportAuthInfo(value: unknown): value is MegaportAuthInfo {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const info = value as Partial<MegaportAuthInfo>;

  return (
    typeof info.accessKeySet === 'boolean' &&
    typeof info.accessKeyPreview === 'string' &&
    typeof info.secretKeySet === 'boolean' &&
    typeof info.secretKeyPreview === 'string' &&
    typeof info.environment === 'string'
  );
}

/**
 * Type guard for MegaportPromptRequest
 */
export function isMegaportPromptRequest(
  value: unknown
): value is MegaportPromptRequest {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const request = value as Partial<MegaportPromptRequest>;

  if (
    typeof request.id !== 'string' ||
    typeof request.message !== 'string' ||
    typeof request.type !== 'string'
  ) {
    return false;
  }

  // resourceType is optional
  if (
    'resourceType' in request &&
    request.resourceType !== undefined &&
    typeof request.resourceType !== 'string'
  ) {
    return false;
  }

  return true;
}

/**
 * Type guard for TelemetryEventType
 */
export function isTelemetryEventType(
  value: unknown
): value is TelemetryEventType {
  const validTypes: TelemetryEventType[] = [
    'wasm_init_start',
    'wasm_init_success',
    'wasm_init_error',
    'command_execute_start',
    'command_execute_success',
    'command_execute_error',
    'auth_set',
    'auth_clear',
    'spinner_start',
    'spinner_stop',
    'prompt_requested',
    'prompt_submitted',
    'prompt_cancelled',
  ];

  return (
    typeof value === 'string' &&
    validTypes.includes(value as TelemetryEventType)
  );
}

/**
 * Type guard for TelemetryEvent
 */
export function isTelemetryEvent(value: unknown): value is TelemetryEvent {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const event = value as Partial<TelemetryEvent>;

  if (!isTelemetryEventType(event.type)) {
    return false;
  }

  if (typeof event.timestamp !== 'number') {
    return false;
  }

  // duration is optional
  if (
    'duration' in event &&
    event.duration !== undefined &&
    typeof event.duration !== 'number'
  ) {
    return false;
  }

  // metadata is optional
  if (
    'metadata' in event &&
    event.metadata !== undefined &&
    (typeof event.metadata !== 'object' || event.metadata === null)
  ) {
    return false;
  }

  return true;
}

/**
 * Type guard for Window WASM functions availability
 */
export function hasWASMFunctions(win: Window): boolean {
  return (
    typeof win.executeMegaportCommand === 'function' ||
    typeof win.executeMegaportCommandAsync === 'function'
  );
}

/**
 * Type guard for Worker support
 */
export function hasWorkerSupport(): boolean {
  return typeof Worker !== 'undefined';
}

/**
 * Type guard for WebAssembly support
 */
export function hasWebAssemblySupport(): boolean {
  return (
    typeof WebAssembly !== 'undefined' &&
    typeof WebAssembly.instantiate === 'function'
  );
}

/**
 * Type guard for string
 */
export function isString(value: unknown): value is string {
  return typeof value === 'string';
}

/**
 * Type guard for non-empty string
 */
export function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}

/**
 * Type guard for Error object
 */
export function isError(value: unknown): value is Error {
  return value instanceof Error;
}

/**
 * Safe error message extraction
 */
export function getErrorMessage(error: unknown): string {
  if (isError(error)) {
    return error.message;
  }
  if (isString(error)) {
    return error;
  }
  return String(error);
}

/**
 * Type guard for object with specific key
 */
export function hasKey<K extends string>(
  obj: unknown,
  key: K
): obj is Record<K, unknown> {
  return typeof obj === 'object' && obj !== null && key in obj;
}

/**
 * Type guard for callable function
 */
export function isFunction(value: unknown): value is Function {
  return typeof value === 'function';
}

/**
 * Safe callback invocation with type checking
 */
export function safeInvoke<T extends any[], R>(
  fn: unknown,
  ...args: T
): R | undefined {
  if (isFunction(fn)) {
    try {
      return fn(...args) as R;
    } catch (error) {
      console.error('Error invoking function:', getErrorMessage(error));
      return undefined;
    }
  }
  return undefined;
}

/**
 * Validate and sanitize command string
 */
export function isValidCommand(command: unknown): command is string {
  if (!isNonEmptyString(command)) {
    return false;
  }

  // Basic command validation - should not contain dangerous patterns
  const dangerousPatterns = [
    /rm\s+-rf\s+\//i, // Dangerous rm command
    /:\(\)\{/i, // Fork bomb pattern
    /eval\(/i, // eval() call
    /<script>/i, // Script tag
  ];

  return !dangerousPatterns.some((pattern) => pattern.test(command));
}
