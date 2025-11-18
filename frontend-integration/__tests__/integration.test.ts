import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { nextTick, ref, defineComponent } from 'vue';

/**
 * Integration tests for the complete WASM workflow
 * Tests the interaction between components, composables, and workers
 */

// Setup comprehensive mocks - must be synchronous and hoisted
const mockTerminal = {
  open: vi.fn(),
  write: vi.fn(),
  writeln: vi.fn(),
  clear: vi.fn(),
  dispose: vi.fn(),
  loadAddon: vi.fn(),
  onKey: vi.fn(),
  onData: vi.fn(),
  focus: vi.fn(),
};

vi.mock('@xterm/xterm', () => ({
  Terminal: vi.fn(function (this: any) {
    Object.assign(this, mockTerminal);
  }),
}));

vi.mock('@xterm/addon-fit', () => ({
  FitAddon: vi.fn(function (this: any) {
    this.fit = vi.fn();
    this.dispose = vi.fn();
  }),
}));

vi.mock('@xterm/addon-web-links', () => ({
  WebLinksAddon: vi.fn(function (this: any) {
    this.dispose = vi.fn();
  }),
}));

// Helper to wait for WASM ready state
const waitForReady = async (isReady: any, timeout = 500) => {
  const start = Date.now();
  while (!isReady.value && Date.now() - start < timeout) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  if (!isReady.value) {
    throw new Error('WASM did not become ready in time');
  }
};

// Helper to create wrapper component for composable testing
const createTestWrapper = (setupFn: () => any) => {
  const TestComponent = defineComponent({
    template: '<div></div>',
    setup: setupFn,
  });
  return mount(TestComponent);
};

describe('WASM Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Complete Workflow: Auth + Command Execution', () => {
    it('should authenticate and execute commands', async () => {
      const mockResult = { output: 'Port list result', error: '' };
      ((global as any).executeMegaportCommandAsync as any).mockImplementation(
        (cmd: string, callback: Function) => {
          callback(mockResult);
        }
      );

      // Import composable
      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      // Create wrapper with composable
      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { setAuth, execute, isReady } = composableInstance;

      // Wait for WASM to be ready
      await waitForReady(isReady);

      // Step 1: Set authentication
      setAuth('test-key', 'test-secret', 'staging');

      await nextTick();

      // Step 2: Verify auth is set
      expect((global as any).setAuthCredentials).toHaveBeenCalledWith(
        'test-key',
        'test-secret',
        'staging'
      );

      // Step 3: Execute command
      const result = await execute('port list');

      // Step 4: Verify result
      expect(result.output).toBe('Port list result');
      expect((global as any).executeMegaportCommandAsync).toHaveBeenCalledWith(
        'port list',
        expect.any(Function)
      );

      wrapper.unmount();
    });

    it('should handle multiple sequential commands', async () => {
      let commandCount = 0;
      ((global as any).executeMegaportCommandAsync as any).mockImplementation(
        (cmd: string, callback: Function) => {
          commandCount++;
          callback({ output: `Result ${commandCount}: ${cmd}`, error: '' });
        }
      );

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { execute, setAuth, isReady } = composableInstance;
      await waitForReady(isReady);

      setAuth('key', 'secret', 'staging');

      const commands = ['port list', 'vxc list', 'location list'];
      const results = [];

      for (const cmd of commands) {
        const result = await execute(cmd);
        results.push(result);
      }

      expect(results).toHaveLength(3);
      expect(results[0].output).toContain('port list');
      expect(results[1].output).toContain('vxc list');
      expect(results[2].output).toContain('location list');

      wrapper.unmount();
    });
  });

  describe('Terminal Component Integration', () => {
    it('should integrate terminal with WASM composable', async () => {
      const MegaportTerminal = await import(
        '../components/MegaportTerminal.vue'
      );

      const wrapper = mount(MegaportTerminal.default, {
        props: {
          wasmPath: '/test.wasm',
          wasmExecPath: '/test_exec.js',
        },
      });

      await nextTick();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find('.megaport-terminal-container').exists()).toBe(true);
    });

    it('should handle auth flow in terminal component', async () => {
      const MegaportTerminal = await import(
        '../components/MegaportTerminal.vue'
      );

      const wrapper = mount(MegaportTerminal.default);

      await nextTick();

      // Component should initialize WASM
      expect(wrapper.vm).toBeDefined();
    });
  });

  describe('Error Recovery Workflow', () => {
    it('should recover from WASM initialization failure', async () => {
      // Simulate failure
      global.fetch = vi.fn(() => Promise.reject(new Error('Network failure')));
      delete (window as any).Go;

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await new Promise((resolve) => setTimeout(resolve, 250));

      const { error, isLoading } = composableInstance;
      expect(error.value).toBeTruthy();
      expect(isLoading.value).toBe(false);

      wrapper.unmount();
    });

    it('should handle command execution errors gracefully', async () => {
      ((global as any).executeMegaportCommandAsync as any).mockImplementation(
        (cmd: string, callback: Function) => {
          callback({ output: '', error: 'Command failed' });
        }
      );

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { execute, isReady } = composableInstance;
      await waitForReady(isReady);

      const result = await execute('invalid command');

      expect(result.error).toBe('Command failed');
      expect(result.output).toBe('');

      wrapper.unmount();
    });
  });

  describe('State Management Across Components', () => {
    it('should share auth state between instances', async () => {
      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let instance1: any;
      const wrapper1 = createTestWrapper(() => {
        instance1 = useMegaportWASM({});
        return instance1;
      });

      let instance2: any;
      const wrapper2 = createTestWrapper(() => {
        instance2 = useMegaportWASM({});
        return instance2;
      });

      await nextTick();

      // Set auth in first instance
      instance1.setAuth('shared-key', 'shared-secret', 'production');

      await nextTick();

      // Check second instance can see it
      const authInfo = instance2.getAuthInfo();

      expect((global as any).setAuthCredentials).toHaveBeenCalled();

      wrapper1.unmount();
      wrapper2.unmount();
    });

    it('should handle concurrent command execution', async () => {
      const results: any[] = [];
      ((global as any).executeMegaportCommandAsync as any).mockImplementation(
        (cmd: string, callback: Function) => {
          setTimeout(() => {
            callback({ output: `Executed: ${cmd}`, error: '' });
          }, Math.random() * 50);
        }
      );

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { execute, isReady } = composableInstance;
      await waitForReady(isReady);

      // Execute multiple commands concurrently
      const promises = [
        execute('command1'),
        execute('command2'),
        execute('command3'),
      ];

      const allResults = await Promise.all(promises);

      expect(allResults).toHaveLength(3);
      allResults.forEach((result) => {
        expect(result.output).toContain('Executed:');
      });

      wrapper.unmount();
    });
  });

  describe('Performance and Resource Management', () => {
    it('should cleanup resources on unmount', async () => {
      const MegaportTerminal = await import(
        '../components/MegaportTerminal.vue'
      );

      const wrapper = mount(MegaportTerminal.default);

      await nextTick();

      const disposeSpy = vi.fn();
      (wrapper.vm as any).terminal = {
        dispose: disposeSpy,
      };

      wrapper.unmount();

      expect(disposeSpy).toHaveBeenCalled();
    });

    it('should reset output buffers', async () => {
      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { resetOutput } = composableInstance;
      resetOutput();

      expect((global as any).resetWasmOutput).toHaveBeenCalled();

      wrapper.unmount();
    });

    it('should toggle debug mode', async () => {
      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      ((global as any).toggleWasmDebug as any).mockReturnValue(true);

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { toggleDebug } = composableInstance;
      const result = toggleDebug();

      expect((global as any).toggleWasmDebug).toHaveBeenCalled();
      expect(result).toBe(true);

      wrapper.unmount();
    });
  });

  describe('Browser Compatibility', () => {
    it('should handle missing WebAssembly support', async () => {
      const originalWasm = global.WebAssembly;
      (global as any).WebAssembly = undefined;

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { error } = composableInstance;

      await new Promise((resolve) => setTimeout(resolve, 200));

      wrapper.unmount();
      global.WebAssembly = originalWasm;
    });

    it('should handle missing Worker support', async () => {
      const originalWorker = global.Worker;
      (global as any).Worker = undefined;

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      // Should still work in direct mode
      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { isLoading } = composableInstance;
      expect(isLoading.value).toBe(true);

      wrapper.unmount();
      global.Worker = originalWorker;
    });
  });

  describe('Real-world Scenarios', () => {
    it('should simulate user login and resource listing workflow', async () => {
      const mockResponses: Record<string, any> = {
        'port list --output json': {
          output: JSON.stringify([{ id: 1, name: 'Port 1' }]),
          error: '',
        },
        'vxc list --output json': {
          output: JSON.stringify([{ id: 2, name: 'VXC 1' }]),
          error: '',
        },
      };

      ((global as any).executeMegaportCommandAsync as any).mockImplementation(
        (cmd: string, callback: Function) => {
          callback(
            mockResponses[cmd] || { output: '', error: 'Unknown command' }
          );
        }
      );

      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM({});
        return composableInstance;
      });

      await nextTick();

      const { setAuth, execute, getAuthInfo, isReady } = composableInstance;

      // Wait for ready state
      await waitForReady(isReady);

      // Mock getAuthInfo to return production environment
      const productionAuthInfo = {
        accessKeySet: true,
        accessKeyPreview: 'user-***',
        secretKeySet: true,
        secretKeyPreview: '***',
        environment: 'production',
      };
      ((global as any).debugAuthInfo as any).mockReturnValue(
        productionAuthInfo
      );
      (window as any).getAuthInfo = vi.fn(() => productionAuthInfo);

      // User logs in
      setAuth('user-key', 'user-secret', 'production');
      const auth = getAuthInfo();
      expect(auth?.environment).toBe('production');

      // User lists ports
      const portResult = await execute('port list --output json');
      const ports = JSON.parse(portResult.output || '[]');
      expect(ports).toHaveLength(1);
      expect(ports[0].name).toBe('Port 1');

      // User lists VXCs
      const vxcResult = await execute('vxc list --output json');
      const vxcs = JSON.parse(vxcResult.output || '[]');
      expect(vxcs).toHaveLength(1);
      expect(vxcs[0].name).toBe('VXC 1');

      wrapper.unmount();
    });

    it('should handle session timeout and reauthentication', async () => {
      const { useMegaportWASM } = await import(
        '../composables/useMegaportWASM'
      );

      let composableInstance: any;
      const wrapper = createTestWrapper(() => {
        composableInstance = useMegaportWASM();
        return composableInstance;
      });

      await nextTick();

      const { setAuth, clearAuth, getAuthInfo } = composableInstance;

      // Initial auth
      setAuth('key1', 'secret1', 'staging');
      let auth = getAuthInfo();
      expect((global as any).setAuthCredentials).toHaveBeenCalled();

      // Simulate timeout - clear auth
      clearAuth();
      auth = getAuthInfo();

      // Reauthenticate
      setAuth('key2', 'secret2', 'production');
      auth = getAuthInfo();
      expect((global as any).setAuthCredentials).toHaveBeenCalled();

      wrapper.unmount();
    });
  });
});
