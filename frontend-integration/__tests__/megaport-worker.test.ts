import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock global worker scope
const mockPostMessage = vi.fn();
const mockImportScripts = vi.fn();

global.self = {
  postMessage: mockPostMessage,
  importScripts: mockImportScripts,
  addEventListener: vi.fn(),
} as any;

// Mock Go runtime as a class
class MockGo {
  run = vi.fn();
  importObject = {};
}

// Mock WASM functions
const mockExecuteMegaportCommandAsync = vi.fn();
const mockSetAuthCredentials = vi.fn(() => ({ success: true }));
const mockClearAuthCredentials = vi.fn(() => ({ success: true }));
const mockResetWasmOutput = vi.fn();

describe('Megaport Worker', () => {
  let messageHandler: Function;

  beforeEach(() => {
    vi.clearAllMocks();

    // Capture message handler
    (global.self.addEventListener as any).mockImplementation(
      (event: string, handler: Function) => {
        if (event === 'message') {
          messageHandler = handler;
        }
      }
    );

    // Setup worker environment
    (global.self as any).Go = MockGo;
    (global.self as any).executeMegaportCommandAsync =
      mockExecuteMegaportCommandAsync;
    (global.self as any).setAuthCredentials = mockSetAuthCredentials;
    (global.self as any).clearAuthCredentials = mockClearAuthCredentials;
    (global.self as any).resetWasmOutput = mockResetWasmOutput;

    // Mock fetch
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
    } as any;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should register message event listener', async () => {
      // Import worker to trigger addEventListener
      await import('../workers/megaport-worker');

      expect(global.self.addEventListener).toHaveBeenCalledWith(
        'message',
        expect.any(Function)
      );
    });

    it('should handle INIT message', async () => {
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: '1',
          wasmPath: '/megaport.wasm',
          wasmExecPath: '/wasm_exec.js',
        },
      };

      await messageHandler(initMessage);

      expect(mockImportScripts).toHaveBeenCalledWith('/wasm_exec.js');
      expect(global.fetch).toHaveBeenCalledWith('/megaport.wasm');
    });

    it('should post READY message after initialization', async () => {
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: 'init-1',
          wasmPath: '/test.wasm',
          wasmExecPath: '/test_exec.js',
        },
      };

      await messageHandler(initMessage);

      // Wait for async operations
      await new Promise((resolve) => setTimeout(resolve, 250));

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'READY',
        id: 'init-1',
      });
    });

    it('should handle initialization errors', async () => {
      // Setup error condition before importing worker
      global.fetch = vi.fn(() => Promise.reject(new Error('Network error')));

      // Clear mocks and reset modules to get fresh worker state
      vi.clearAllMocks();
      vi.resetModules();

      // Re-setup addEventListener to capture new message handler
      let errorTestMessageHandler: Function = () => {};
      (global.self.addEventListener as any).mockImplementation(
        (event: string, handler: Function) => {
          if (event === 'message') {
            errorTestMessageHandler = handler;
          }
        }
      );

      // Import fresh worker instance
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: 'error-1',
          wasmPath: '/fail.wasm',
          wasmExecPath: '/fail_exec.js',
        },
      };

      await errorTestMessageHandler(initMessage);

      // Wait for async error handling
      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'ERROR',
        id: 'error-1',
        error: expect.stringContaining('Network error'),
      });
    });

    it('should not reinitialize if already ready', async () => {
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: '1',
          wasmPath: '/test.wasm',
          wasmExecPath: '/test_exec.js',
        },
      };

      await messageHandler(initMessage);
      await new Promise((resolve) => setTimeout(resolve, 250));

      const callCount = mockImportScripts.mock.calls.length;

      // Send init again
      await messageHandler(initMessage);
      await new Promise((resolve) => setTimeout(resolve, 250));

      // Should not call importScripts again
      expect(mockImportScripts.mock.calls.length).toBe(callCount);
    });
  });

  describe('Authentication', () => {
    it('should handle SET_AUTH message', async () => {
      await import('../workers/megaport-worker');

      const authMessage = {
        data: {
          type: 'SET_AUTH',
          id: 'auth-1',
          accessKey: 'test-key',
          secretKey: 'test-secret',
          environment: 'staging',
        },
      };

      await messageHandler(authMessage);

      expect(mockSetAuthCredentials).toHaveBeenCalledWith(
        'test-key',
        'test-secret',
        'staging'
      );
    });

    it('should post AUTH_SET confirmation', async () => {
      await import('../workers/megaport-worker');

      const authMessage = {
        data: {
          type: 'SET_AUTH',
          id: 'auth-confirm',
          accessKey: 'key',
          secretKey: 'secret',
          environment: 'production',
        },
      };

      await messageHandler(authMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'AUTH_SET',
        id: 'auth-confirm',
      });
    });

    it('should handle missing auth function', async () => {
      (global.self as any).setAuthCredentials = undefined;

      await import('../workers/megaport-worker');

      const authMessage = {
        data: {
          type: 'SET_AUTH',
          id: 'auth-error',
          accessKey: 'key',
          secretKey: 'secret',
          environment: 'staging',
        },
      };

      await messageHandler(authMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'ERROR',
        id: 'auth-error',
        error: 'setAuthCredentials function not available',
      });
    });
  });

  describe('Command Execution', () => {
    beforeEach(async () => {
      // Initialize worker first
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: 'init',
          wasmPath: '/test.wasm',
          wasmExecPath: '/test_exec.js',
        },
      };

      await messageHandler(initMessage);
      await new Promise((resolve) => setTimeout(resolve, 250));
      mockPostMessage.mockClear();
    });

    it('should handle EXECUTE message', async () => {
      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          callback({ output: `Executed: ${cmd}`, error: '' });
        }
      );

      const executeMessage = {
        data: {
          type: 'EXECUTE',
          id: 'exec-1',
          command: 'port list',
        },
      };

      await messageHandler(executeMessage);

      expect(mockExecuteMegaportCommandAsync).toHaveBeenCalledWith(
        'port list',
        expect.any(Function)
      );
    });

    it('should post command results', async () => {
      const mockResult = {
        output: 'Command output',
        error: '',
      };

      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          callback(mockResult);
        }
      );

      const executeMessage = {
        data: {
          type: 'EXECUTE',
          id: 'exec-result',
          command: 'vxc list',
        },
      };

      await messageHandler(executeMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'RESULT',
        id: 'exec-result',
        result: mockResult,
      });
    });

    it('should reset buffers before execution', async () => {
      mockExecuteMegaportCommandAsync.mockImplementation(
        (cmd: string, callback: Function) => {
          callback({ output: 'test', error: '' });
        }
      );

      const executeMessage = {
        data: {
          type: 'EXECUTE',
          id: 'exec-reset',
          command: 'test command',
        },
      };

      await messageHandler(executeMessage);

      expect(mockResetWasmOutput).toHaveBeenCalled();
    });

    it('should handle missing execute function', async () => {
      (global.self as any).executeMegaportCommandAsync = undefined;
      (global.self as any).executeMegaportCommand = undefined;

      const executeMessage = {
        data: {
          type: 'EXECUTE',
          id: 'exec-error',
          command: 'test',
        },
      };

      await messageHandler(executeMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'ERROR',
        id: 'exec-error',
        error: expect.stringContaining('No WASM execute function'),
      });
    });

    it('should fallback to sync execution', async () => {
      (global.self as any).executeMegaportCommandAsync = undefined;
      const mockSyncExecute = vi.fn(() => ({
        output: 'sync result',
        error: '',
      }));
      (global.self as any).executeMegaportCommand = mockSyncExecute;

      const executeMessage = {
        data: {
          type: 'EXECUTE',
          id: 'sync-exec',
          command: 'sync test',
        },
      };

      await messageHandler(executeMessage);

      expect(mockSyncExecute).toHaveBeenCalledWith('sync test');
      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'RESULT',
        id: 'sync-exec',
        result: { output: 'sync result', error: '' },
      });
    });
  });

  describe('Buffer Reset', () => {
    it('should handle RESET message', async () => {
      await import('../workers/megaport-worker');

      const resetMessage = {
        data: {
          type: 'RESET',
          id: 'reset-1',
        },
      };

      await messageHandler(resetMessage);

      expect(mockResetWasmOutput).toHaveBeenCalled();
    });

    it('should post RESET_COMPLETE confirmation', async () => {
      await import('../workers/megaport-worker');

      const resetMessage = {
        data: {
          type: 'RESET',
          id: 'reset-confirm',
        },
      };

      await messageHandler(resetMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'RESET_COMPLETE',
        id: 'reset-confirm',
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle unknown message types', async () => {
      await import('../workers/megaport-worker');

      const unknownMessage = {
        data: {
          type: 'UNKNOWN',
          id: 'unknown-1',
        },
      };

      await messageHandler(unknownMessage);

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'ERROR',
        id: 'unknown-1',
        error: expect.stringContaining('Unknown message type'),
      });
    });

    it('should handle exceptions during message processing', async () => {
      await import('../workers/megaport-worker');

      const badMessage = {
        data: { type: 'INVALID' },
      };

      await messageHandler(badMessage);

      expect(mockPostMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'ERROR',
        })
      );
    });

    it('should handle WASM instantiation errors', async () => {
      // Setup error condition before importing worker
      global.WebAssembly.instantiate = vi.fn(() =>
        Promise.reject(new Error('Invalid WASM module'))
      );

      // Clear mocks and reset modules to get fresh worker state
      vi.clearAllMocks();
      vi.resetModules();

      // Re-setup addEventListener to capture new message handler
      let errorTestMessageHandler: Function = () => {};
      (global.self.addEventListener as any).mockImplementation(
        (event: string, handler: Function) => {
          if (event === 'message') {
            errorTestMessageHandler = handler;
          }
        }
      );

      // Import fresh worker instance
      await import('../workers/megaport-worker');

      const initMessage = {
        data: {
          type: 'INIT',
          id: 'wasm-error',
          wasmPath: '/bad.wasm',
          wasmExecPath: '/exec.js',
        },
      };

      await errorTestMessageHandler(initMessage);

      // Wait for async error handling
      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(mockPostMessage).toHaveBeenCalledWith({
        type: 'ERROR',
        id: 'wasm-error',
        error: expect.stringContaining('Invalid WASM module'),
      });
    });
  });

  describe('Message Protocol', () => {
    it('should include message ID in responses', async () => {
      await import('../workers/megaport-worker');

      const messages = [
        {
          type: 'SET_AUTH',
          id: 'auth-id',
          accessKey: '',
          secretKey: '',
          environment: 'staging',
        },
        { type: 'RESET', id: 'reset-id' },
      ];

      for (const data of messages) {
        mockPostMessage.mockClear();
        await messageHandler({ data });

        expect(mockPostMessage).toHaveBeenCalledWith(
          expect.objectContaining({ id: data.id })
        );
      }
    });

    it('should handle messages with missing IDs gracefully', async () => {
      await import('../workers/megaport-worker');

      const noIdMessage = {
        data: {
          type: 'RESET',
        },
      };

      await messageHandler(noIdMessage);

      // Should still process the message
      expect(mockResetWasmOutput).toHaveBeenCalled();
    });
  });
});
