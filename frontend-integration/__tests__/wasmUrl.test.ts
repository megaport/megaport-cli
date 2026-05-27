import { describe, it, expect, afterEach } from 'vitest';
import { resolveWasmUrl, DEFAULT_WASM_PATH } from '@/utils/wasmUrl';

describe('resolveWasmUrl', () => {
  afterEach(() => {
    delete window.__MEGAPORT_WASM_URL__;
  });

  it('falls back to the fixed path when no global is injected', () => {
    expect(resolveWasmUrl()).toBe(DEFAULT_WASM_PATH);
    expect(DEFAULT_WASM_PATH).toBe('/megaport.wasm');
  });

  it('prefers the build-time-injected content-hashed URL', () => {
    window.__MEGAPORT_WASM_URL__ = '/megaport.5f560da7.wasm';
    expect(resolveWasmUrl()).toBe('/megaport.5f560da7.wasm');
  });

  it('ignores an empty injected value and falls back', () => {
    window.__MEGAPORT_WASM_URL__ = '';
    expect(resolveWasmUrl()).toBe(DEFAULT_WASM_PATH);
  });
});
