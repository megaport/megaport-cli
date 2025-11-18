import { vi, beforeEach, afterEach } from 'vitest';

// Proper Go constructor for WASM
class MockGo {
  run = vi.fn();
  importObject = {};
}

// Setup global test environment
beforeEach(() => {
  // Mock Go constructor - must be available before any imports
  (window as any).Go = MockGo;
  (global as any).Go = MockGo;

  // Mock window WASM functions
  (window as any).executeMegaportCommandAsync = vi.fn();
  (window as any).executeMegaportCommand = vi.fn();
  (window as any).setAuthCredentials = vi.fn(() => ({ success: true }));
  (window as any).clearAuthCredentials = vi.fn(() => ({ success: true }));
  (window as any).resetWasmOutput = vi.fn();
  (window as any).toggleWasmDebug = vi.fn(() => false);
  (window as any).debugAuthInfo = vi.fn(() => ({
    accessKeySet: false,
    accessKeyPreview: '',
    secretKeySet: false,
    secretKeyPreview: '',
    environment: 'staging',
  }));
  (window as any).getAuthInfo = vi.fn(() => ({
    accessKeySet: false,
    accessKeyPreview: '',
    secretKeySet: false,
    secretKeyPreview: '',
    environment: 'staging',
  }));

  // Mock fetch for WASM loading - synchronous resolution
  global.fetch = vi.fn(() =>
    Promise.resolve({
      ok: true,
      arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
    } as Response)
  );

  // Mock WebAssembly - synchronous resolution
  global.WebAssembly = {
    instantiate: vi.fn(() =>
      Promise.resolve({
        instance: {},
        module: {},
      })
    ),
    instantiateStreaming: vi.fn(),
  } as any;

  // Mock Worker constructor
  global.Worker = vi.fn(function (this: any, url: string) {
    this.url = url;
    this.postMessage = vi.fn();
    this.addEventListener = vi.fn();
    this.removeEventListener = vi.fn();
    this.terminate = vi.fn();
    return this;
  }) as any;

  // Mock document.createElement for script loading - synchronous Go availability
  const originalCreateElement = document.createElement.bind(document);
  document.createElement = vi.fn((tag: string) => {
    const element = originalCreateElement(tag);
    if (tag === 'script') {
      // Make Go available immediately instead of async
      (window as any).Go = MockGo;
      // Trigger onload synchronously in next tick
      setTimeout(() => {
        if (element.onload) {
          element.onload(new Event('load'));
        }
      }, 0);
    }
    return element;
  }) as any;
});

afterEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});
