/** Served path used when the build hasn't injected a content-hashed wasm URL. */
export const DEFAULT_WASM_PATH = '/megaport.wasm';

/**
 * Resolve the wasm URL, preferring the build-time-injected content-hashed name
 * (`window.__MEGAPORT_WASM_URL__`) so the CDN can serve it immutable. SSR-safe:
 * falls back to the fixed path when there's no window or no injected value.
 */
export function resolveWasmUrl(): string {
  if (typeof window !== 'undefined' && window.__MEGAPORT_WASM_URL__) {
    return window.__MEGAPORT_WASM_URL__;
  }
  return DEFAULT_WASM_PATH;
}
