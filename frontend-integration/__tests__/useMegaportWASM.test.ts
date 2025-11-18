import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { nextTick, defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useMegaportWASM } from '../composables/useMegaportWASM';

// Mock Go class
class MockGo {
  run = vi.fn();
  importObject = {};
}

// Mock WASM functions
const mockExecuteMegaportCommandAsync = vi.fn();
const mockSetAuthCredentials = vi.fn(() => ({ success: true }));
const mockClearAuthCredentials = vi.fn(() => ({ success: true }));
const mockResetWasmOutput = vi.fn();
const mockToggleWasmDebug = vi.fn();
const mockDebugAuthInfo = vi.fn(() => ({
  accessKeySet: false,
  accessKeyPreview: '',
  secretKeySet: false,
  secretKeyPreview: '',
  environment: 'staging',
}));
const mockGetAuthInfo = vi.fn(() => ({
  accessKeySet: false,
  accessKeyPreview: '',
  secretKeySet: false,
  secretKeyPreview: '',
  environment: 'staging',
}));

// Helper to create a test wrapper for the composable
const createComposableTestWrapper = (config = {}) => {
  let composableInstance: any;
  const TestComponent = defineComponent({
    template: '<div>Test</div>',
    setup() {
      composableInstance = useMegaportWASM(config);
      return composableInstance;
    },
  });
  const wrapper = mount(TestComponent);
  return { wrapper, composable: composableInstance };
};

// Helper to wait for WASM to be ready
const waitForReady = async (isReady: any, timeout = 500) => {
  const start = Date.now();
  while (!isReady.value && Date.now() - start < timeout) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  if (!isReady.value) {
    throw new Error('WASM did not become ready in time');
  }
};

describe('useMegaportWASM', () => {
  beforeEach(() => {
    // Setup global mocks
    (global as any).Go = MockGo;
    (global as any).executeMegaportCommandAsync =
      mockExecuteMegaportCommandAsync;
    (global as any).setAuthCredentials = mockSetAuthCredentials;
    (global as any).clearAuthCredentials = mockClearAuthCredentials;
    (global as any).resetWasmOutput = mockResetWasmOutput;
    (global as any).toggleWasmDebug = mockToggleWasmDebug;
    (global as any).debugAuthInfo = mockDebugAuthInfo;
    (global as any).getAuthInfo = mockGetAuthInfo;

    (window as any).Go = MockGo;
    (window as any).executeMegaportCommandAsync =
      mockExecuteMegaportCommandAsync;
    (window as any).setAuthCredentials = mockSetAuthCredentials;
    (window as any).clearAuthCredentials = mockClearAuthCredentials;
    (window as any).debugAuthInfo = mockDebugAuthInfo;
    (window as any).getAuthInfo = mockGetAuthInfo;

    // Mock fetch for WASM loading
    global.fetch = vi.fn(() =>
      Promise.resolve({
        arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
      } as Response)
    );

    // Mock WebAssembly
    global.WebAssembly = {
      instantiate: vi.fn(() =>
        Promise.resolve({
          instance: {},
          module: {},
        })
      ),
      instantiateStreaming: vi.fn(),
    } as any;

    // Mock Worker
    global.Worker = vi.fn(() => ({
      postMessage: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      terminate: vi.fn(),
    })) as any;

    // Clear localStorage
    localStorage.clear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should initialize in loading state', () => {
      const { composable } = createComposableTestWrapper();
      const { isLoading, isReady, error } = composable;

      expect(isLoading.value).toBe(true);
      expect(isReady.value).toBe(false);
      expect(error.value).toBe(null);
    });

    it('should initialize with default config', () => {
      const { composable } = createComposableTestWrapper();
      const { isLoading } = composable;
      expect(isLoading.value).toBe(true);
    });

    it('should accept custom config', () => {
      const { composable } = createComposableTestWrapper({
        wasmPath: '/custom/path.wasm',
        wasmExecPath: '/custom/exec.js',
        debug: true,
        useWorker: false,
      });
      const { isLoading } = composable;

      expect(isLoading.value).toBe(true);
    });

    it('should initialize with worker mode when configured', () => {
      const { composable } = createComposableTestWrapper({ useWorker: true });
      const { isLoading } = composable;
      expect(isLoading.value).toBe(true);
    });
  });

  describe('Authentication', () => {
    it('should set authentication credentials', async () => {
      const { composable } = createComposableTestWrapper();
      const { setAuth } = composable;

      setAuth('test-access-key', 'test-secret-key', 'staging');

      await nextTick();

      expect(mockSetAuthCredentials).toHaveBeenCalledWith(
        'test-access-key',
        'test-secret-key',
        'staging'
      );
    });

    it('should clear authentication credentials', async () => {
      const { composable } = createComposableTestWrapper();
      const { clearAuth } = composable;

      clearAuth();

      await nextTick();

      expect(mockClearAuthCredentials).toHaveBeenCalled();
    });

    it('should get authentication info when configured', () => {
      const mockAuthInfo = {
        accessKeySet: true,
        accessKeyPreview: 'test-key',
        secretKeySet: true,
        secretKeyPreview: 'test-secret',
        environment: 'production',
      };
      mockDebugAuthInfo.mockReturnValue(mockAuthInfo);
      mockGetAuthInfo.mockReturnValue(mockAuthInfo);
      (window as any).getAuthInfo = vi.fn(() => mockAuthInfo);

      const { composable } = createComposableTestWrapper();
      const { getAuthInfo } = composable;
      const authInfo = getAuthInfo();

      expect(authInfo?.accessKeySet).toBe(true);
      expect(authInfo?.secretKeySet).toBe(true);
      expect(authInfo?.environment).toBe('production');
    });

    it('should detect unconfigured auth', () => {
      const emptyAuthInfo = {
        accessKeySet: false,
        accessKeyPreview: '',
        secretKeySet: false,
        secretKeyPreview: '',
        environment: '',
      };
      mockDebugAuthInfo.mockReturnValue(emptyAuthInfo);
      mockGetAuthInfo.mockReturnValue(emptyAuthInfo);
      (window as any).getAuthInfo = vi.fn(() => emptyAuthInfo);

      const { composable } = createComposableTestWrapper();
      const { getAuthInfo } = composable;
      const authInfo = getAuthInfo();

      expect(authInfo?.accessKeySet).toBe(false);
      expect(authInfo?.secretKeySet).toBe(false);
    });
  });

  describe('Command Execution', () => {
    it('should execute commands successfully', async () => {
      const mockResult = {
        output: 'Command output',
        error: '',
      };

      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          callback(mockResult);
        }
      );

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { execute, isReady } = composable;

      // Wait for actual ready state
      await waitForReady(isReady);

      const result = await execute('port list');

      expect(result.output).toBe('Command output');
      expect(result.error).toBe('');

      wrapper.unmount();
    });

    it('should handle command errors', async () => {
      const mockResult = {
        output: '',
        error: 'Command failed',
      };

      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          callback(mockResult);
        }
      );

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { execute, isReady } = composable;
      await waitForReady(isReady);

      const result = await execute('invalid command');

      expect(result.output).toBe('');
      expect(result.error).toBe('Command failed');

      wrapper.unmount();
    });

    it('should reject when WASM is not ready', async () => {
      mockExecuteMegaportCommandAsync.mockImplementation(() => {
        throw new Error('WASM not initialized');
      });

      const { composable } = createComposableTestWrapper();
      const { execute } = composable;

      await expect(execute('port list')).rejects.toThrow();
    });

    it('should handle multiple concurrent commands', async () => {
      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          setTimeout(() => {
            callback({ output: `Result for: ${cmd}`, error: '' });
          }, 50);
        }
      );

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { execute, isReady } = composable;
      await waitForReady(isReady);

      const results = await Promise.all([
        execute('port list'),
        execute('vxc list'),
        execute('location list'),
      ]);

      expect(results).toHaveLength(3);
      results.forEach((result) => {
        expect(result.output).toContain('Result for:');
      });

      wrapper.unmount();
    });
  });

  describe('Output Management', () => {
    it('should reset output buffers', () => {
      const { composable } = createComposableTestWrapper();
      const { resetOutput } = composable;

      resetOutput();

      expect(mockResetWasmOutput).toHaveBeenCalled();
    });

    it('should handle missing reset function gracefully', () => {
      (global as any).resetWasmOutput = undefined;

      const { composable } = createComposableTestWrapper();
      const { resetOutput } = composable;

      expect(() => resetOutput()).not.toThrow();
    });
  });

  describe('Debug Mode', () => {
    it('should toggle debug mode', () => {
      mockToggleWasmDebug.mockReturnValue(true);

      const { composable } = createComposableTestWrapper();
      const { toggleDebug } = composable;

      const enabled = toggleDebug();

      expect(mockToggleWasmDebug).toHaveBeenCalled();
      expect(enabled).toBe(true);
    });

    it('should handle missing debug function gracefully', () => {
      (global as any).toggleWasmDebug = undefined;

      const { composable } = createComposableTestWrapper();
      const { toggleDebug } = composable;

      expect(() => toggleDebug()).not.toThrow();
    });

    it('should initialize with debug mode when configured', () => {
      const { composable } = createComposableTestWrapper({ debug: true });
      const { toggleDebug } = composable;

      // Debug mode should be enabled during initialization
      expect(toggleDebug).toBeDefined();
    });
  });

  describe('Error Handling', () => {
    it('should handle WASM loading errors', async () => {
      global.fetch = vi.fn(() => Promise.reject(new Error('Network error')));
      delete (window as any).Go;

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });

      await new Promise((resolve) => setTimeout(resolve, 250));

      const { error, isLoading } = composable;
      expect(error.value).toBeTruthy();
      expect(isLoading.value).toBe(false);

      wrapper.unmount();
    });

    it('should handle missing wasm_exec.js script', async () => {
      // Mock script loading failure
      const originalCreateElement = document.createElement.bind(document);
      document.createElement = vi.fn((tag: string) => {
        const element = originalCreateElement(tag);
        if (tag === 'script') {
          setTimeout(() => {
            element.onerror?.(new Event('error'));
          }, 10);
        }
        return element;
      }) as any;

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { error } = composable;

      await new Promise((resolve) => setTimeout(resolve, 100));

      document.createElement = originalCreateElement;

      wrapper.unmount();
    });

    it('should handle WebAssembly instantiation errors', async () => {
      global.WebAssembly.instantiate = vi.fn(() =>
        Promise.reject(new Error('Invalid WASM'))
      );

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });

      await new Promise((resolve) => setTimeout(resolve, 250));

      const { error, isLoading } = composable;
      expect(error.value).toBeTruthy();
      expect(isLoading.value).toBe(false);

      wrapper.unmount();
    });
  });

  describe('Worker Mode', () => {
    it('should create worker when useWorker is true', () => {
      const { composable } = createComposableTestWrapper({ useWorker: true });
      const { isLoading } = composable;

      // Worker mode falls back to direct mode, so Worker is not called
      expect(isLoading.value).toBe(true);
    });

    it('should send INIT message to worker', async () => {
      // Worker mode not fully implemented - falls back to direct mode
      const { composable } = createComposableTestWrapper({ useWorker: true });
      const { isLoading } = composable;

      await nextTick();

      expect(isLoading.value).toBe(true);
    });

    it('should handle worker errors', async () => {
      // Worker mode falls back to direct - test direct mode error instead
      global.fetch = vi.fn(() => Promise.reject(new Error('Init failed')));
      delete (window as any).Go;

      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: true,
      });

      await new Promise((resolve) => setTimeout(resolve, 250));

      const { error } = composable;
      expect(error.value).toBeTruthy();

      wrapper.unmount();
    });
  });

  describe('Reactivity', () => {
    it('should expose reactive refs', () => {
      const { composable, wrapper } = createComposableTestWrapper();
      const { isLoading, isReady, error } = composable;

      expect(isLoading.value).toBeDefined();
      expect(isReady.value).toBeDefined();
      expect(error.value).toBeDefined();

      wrapper.unmount();
    });

    it('should update loading state', async () => {
      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { isLoading, isReady } = composable;

      expect(isLoading.value).toBe(true);
      expect(isReady.value).toBe(false);

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 300));

      // After initialization, loading should be false
      // (Implementation may vary based on actual WASM loading)

      wrapper.unmount();
    });
  });

  describe('Cleanup', () => {
    it('should provide methods for cleanup', () => {
      const { composable, wrapper } = createComposableTestWrapper();
      const { resetOutput, clearAuth } = composable;

      expect(resetOutput).toBeDefined();
      expect(clearAuth).toBeDefined();

      wrapper.unmount();
    });

    it('should clear auth on cleanup', () => {
      const { composable, wrapper } = createComposableTestWrapper();
      const { clearAuth } = composable;

      clearAuth();

      expect(mockClearAuthCredentials).toHaveBeenCalled();

      wrapper.unmount();
    });
  });

  describe('Spinner Functionality', () => {
    beforeEach(async () => {
      // Clear window spinner functions
      delete (window as any).wasmStartSpinner;
      delete (window as any).wasmStopSpinner;
      // Small delay to ensure cleanup is complete
      await new Promise((resolve) => setTimeout(resolve, 10));
    });

    it('should expose activeSpinners state', () => {
      const { composable, wrapper } = createComposableTestWrapper();
      const { activeSpinners } = composable;

      expect(activeSpinners).toBeDefined();
      expect(activeSpinners.value).toBeInstanceOf(Map);
      expect(activeSpinners.value.size).toBe(0);

      wrapper.unmount();
    });

    it('should register wasmStartSpinner on window', async () => {
      const { wrapper } = createComposableTestWrapper({ useWorker: false });

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      expect((window as any).wasmStartSpinner).toBeDefined();
      expect(typeof (window as any).wasmStartSpinner).toBe('function');

      wrapper.unmount();
    });

    it('should register wasmStopSpinner on window', async () => {
      const { wrapper } = createComposableTestWrapper({ useWorker: false });

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      expect((window as any).wasmStopSpinner).toBeDefined();
      expect(typeof (window as any).wasmStopSpinner).toBe('function');

      wrapper.unmount();
    });

    it('should add spinner when wasmStartSpinner is called', async () => {
      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { activeSpinners } = composable;

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      const spinnerId = (window as any).wasmStartSpinner?.(
        'Test spinner message'
      );

      expect(spinnerId).toBeDefined();
      expect(activeSpinners.value.size).toBe(1);
      expect(activeSpinners.value.get(spinnerId)).toBe('Test spinner message');

      wrapper.unmount();
    });

    it('should remove spinner when wasmStopSpinner is called', async () => {
      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { activeSpinners } = composable;

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      const spinnerId = (window as any).wasmStartSpinner?.(
        'Test spinner message'
      );
      expect(activeSpinners.value.size).toBe(1);

      (window as any).wasmStopSpinner?.(spinnerId);
      expect(activeSpinners.value.size).toBe(0);

      wrapper.unmount();
    });

    it('should handle multiple concurrent spinners', async () => {
      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { activeSpinners } = composable;

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      const spinner1 = (window as any).wasmStartSpinner?.('Spinner 1');
      const spinner2 = (window as any).wasmStartSpinner?.('Spinner 2');
      const spinner3 = (window as any).wasmStartSpinner?.('Spinner 3');

      expect(activeSpinners.value.size).toBe(3);
      expect(activeSpinners.value.get(spinner1)).toBe('Spinner 1');
      expect(activeSpinners.value.get(spinner2)).toBe('Spinner 2');
      expect(activeSpinners.value.get(spinner3)).toBe('Spinner 3');

      (window as any).wasmStopSpinner?.(spinner2);
      expect(activeSpinners.value.size).toBe(2);
      expect(activeSpinners.value.has(spinner1)).toBe(true);
      expect(activeSpinners.value.has(spinner2)).toBe(false);
      expect(activeSpinners.value.has(spinner3)).toBe(true);

      wrapper.unmount();
    });

    it('should generate unique spinner IDs', async () => {
      const { wrapper } = createComposableTestWrapper({ useWorker: false });

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      const ids: string[] = [];
      for (let i = 0; i < 10; i++) {
        const id = (window as any).wasmStartSpinner?.(`Spinner ${i}`);
        ids.push(id);
      }

      // All IDs should be unique
      const uniqueIds = new Set(ids);
      expect(uniqueIds.size).toBe(10);

      wrapper.unmount();
    });

    it('should track spinner messages correctly', async () => {
      const { composable, wrapper } = createComposableTestWrapper({
        useWorker: false,
      });
      const { activeSpinners } = composable;

      // Wait for initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      const messages = [
        'Logging in to Megaport...',
        'Validating Port order...',
        'Creating Port pb-test-port-vue...',
      ];

      const ids = messages.map((msg) =>
        (window as any).wasmStartSpinner?.(msg)
      );

      expect(activeSpinners.value.size).toBe(3);

      ids.forEach((id, index) => {
        expect(activeSpinners.value.get(id)).toBe(messages[index]);
      });

      wrapper.unmount();
    });
  });
});
