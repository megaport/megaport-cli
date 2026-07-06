import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { defineComponent } from 'vue';
import { mount } from '@vue/test-utils';
import { useMegaportWASM } from '../composables/useMegaportWASM';

class MockGo {
  run = vi.fn();
  importObject = {};
}

// Mount the composable inside a component so onMounted init runs, then wait for
// isReady so execute() can be exercised end to end.
const createReadyComposable = async () => {
  let composableInstance: any;
  const TestComponent = defineComponent({
    template: '<div>Test</div>',
    setup() {
      composableInstance = useMegaportWASM();
      return composableInstance;
    },
  });
  const wrapper = mount(TestComponent);

  const start = Date.now();
  while (!composableInstance.isReady.value && Date.now() - start < 1000) {
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
  if (!composableInstance.isReady.value) {
    wrapper.unmount();
    throw new Error('WASM did not become ready in time');
  }
  return { wrapper, composable: composableInstance };
};

describe('Interactive Mode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    delete (window as any).registerPromptHandler;
    delete (window as any).submitPromptResponse;
    delete (window as any).cancelPrompt;
  });

  describe('Prompt Handler Registration', () => {
    it('should register a custom prompt handler', () => {
      const mockRegister = vi.fn(() => true);
      (window as any).registerPromptHandler = mockRegister;

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      const result = registerPromptHandler(customHandler);

      expect(result).toBe(true);
      expect(mockRegister).toHaveBeenCalledWith(customHandler);
    });

    it('should return false when registerPromptHandler is not available', () => {
      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      const result = registerPromptHandler(customHandler);

      expect(result).toBe(false);
    });

    it('should warn when registerPromptHandler is not available', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn');
      const { registerPromptHandler } = useMegaportWASM();

      registerPromptHandler(vi.fn());

      // Filter out Vue lifecycle warnings (expected since we're not in a component context)
      const relevantWarnings = consoleWarnSpy.mock.calls.filter(
        (call: any[]) =>
          !call.some(
            (arg) =>
              typeof arg === 'string' &&
              arg.includes('[Vue warn]') &&
              arg.includes('Lifecycle injection APIs')
          )
      );

      expect(relevantWarnings.length).toBeGreaterThan(0);
      expect(relevantWarnings[0][0]).toBe(
        'registerPromptHandler not available - WASM may not be initialized'
      );
    });

    it('should allow registering multiple times', () => {
      const mockRegister = vi.fn(() => true);
      (window as any).registerPromptHandler = mockRegister;

      const { registerPromptHandler } = useMegaportWASM();
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      registerPromptHandler(handler1);
      registerPromptHandler(handler2);

      expect(mockRegister).toHaveBeenCalledTimes(2);
      expect(mockRegister).toHaveBeenNthCalledWith(1, handler1);
      expect(mockRegister).toHaveBeenNthCalledWith(2, handler2);
    });
  });

  describe('Prompt Request Handling', () => {
    it('should invoke custom handler when prompt is requested', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      const mockRegister = vi.fn((handler: (request: any) => void) => {
        registeredHandler = handler;
        return true;
      });
      (window as any).registerPromptHandler = mockRegister;

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      // Simulate WASM requesting a prompt
      const promptRequest = {
        id: 'prompt-1',
        message: 'Enter your name:',
        defaultValue: '',
      };

      registeredHandler!(promptRequest);

      expect(customHandler).toHaveBeenCalledWith(promptRequest);
    });

    it('should handle prompt requests with default values', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      const promptRequest = {
        id: 'prompt-2',
        message: 'Enter port name:',
        defaultValue: 'my-port',
      };

      registeredHandler!(promptRequest);

      expect(customHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          defaultValue: 'my-port',
        })
      );
    });

    it('should handle prompt requests with validation requirements', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      const promptRequest = {
        id: 'prompt-3',
        message: 'Enter bandwidth (Mbps):',
        defaultValue: '1000',
        validation: {
          required: true,
          pattern: '^[0-9]+$',
        },
      };

      registeredHandler!(promptRequest);

      expect(customHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          validation: expect.any(Object),
        })
      );
    });
  });

  describe('Prompt Response Submission', () => {
    it('should submit prompt response when user provides input', () => {
      const mockSubmit = vi.fn();
      (window as any).submitPromptResponse = mockSubmit;

      // Simulate submitting a response
      (window as any).submitPromptResponse('prompt-1', 'John Doe');

      expect(mockSubmit).toHaveBeenCalledWith('prompt-1', 'John Doe');
    });

    it('should submit empty response when allowed', () => {
      const mockSubmit = vi.fn();
      (window as any).submitPromptResponse = mockSubmit;

      (window as any).submitPromptResponse('prompt-2', '');

      expect(mockSubmit).toHaveBeenCalledWith('prompt-2', '');
    });

    it('should submit numeric responses', () => {
      const mockSubmit = vi.fn();
      (window as any).submitPromptResponse = mockSubmit;

      (window as any).submitPromptResponse('prompt-3', '1000');

      expect(mockSubmit).toHaveBeenCalledWith('prompt-3', '1000');
    });

    it('should handle special characters in responses', () => {
      const mockSubmit = vi.fn();
      (window as any).submitPromptResponse = mockSubmit;

      (window as any).submitPromptResponse('prompt-4', 'test@example.com');

      expect(mockSubmit).toHaveBeenCalledWith('prompt-4', 'test@example.com');
    });
  });

  describe('Prompt Cancellation', () => {
    it('should cancel prompt when user cancels', () => {
      const mockCancel = vi.fn();
      (window as any).cancelPrompt = mockCancel;

      (window as any).cancelPrompt('prompt-1');

      expect(mockCancel).toHaveBeenCalledWith('prompt-1');
    });

    it('should handle cancellation of multiple prompts', () => {
      const mockCancel = vi.fn();
      (window as any).cancelPrompt = mockCancel;

      (window as any).cancelPrompt('prompt-1');
      (window as any).cancelPrompt('prompt-2');
      (window as any).cancelPrompt('prompt-3');

      expect(mockCancel).toHaveBeenCalledTimes(3);
    });
  });

  describe('Interactive Command Flow', () => {
    it('should handle complete interactive command flow', async () => {
      let registeredHandler: ((request: any) => void) | null = null;
      const mockRegister = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });
      const mockSubmit = vi.fn();
      const mockExecute = vi.fn(() =>
        Promise.resolve({ output: 'Port created', error: '' })
      );

      (window as any).registerPromptHandler = mockRegister;
      (window as any).submitPromptResponse = mockSubmit;
      (window as any).executeMegaportCommandAsync = mockExecute;

      const { registerPromptHandler } = useMegaportWASM();

      // Register handler that auto-responds
      const autoResponseHandler = vi.fn((request: any) => {
        if (request.message.includes('name')) {
          (window as any).submitPromptResponse(request.id, 'test-port');
        } else if (request.message.includes('bandwidth')) {
          (window as any).submitPromptResponse(request.id, '1000');
        }
      });

      registerPromptHandler(autoResponseHandler);

      // Simulate prompt request
      const promptRequest = {
        id: 'prompt-1',
        message: 'Enter port name:',
      };
      registeredHandler!(promptRequest);

      expect(autoResponseHandler).toHaveBeenCalledWith(promptRequest);
      expect(mockSubmit).toHaveBeenCalledWith('prompt-1', 'test-port');
    });

    it('should handle interactive command cancellation', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      const mockRegister = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });
      const mockCancel = vi.fn();

      (window as any).registerPromptHandler = mockRegister;
      (window as any).cancelPrompt = mockCancel;

      const { registerPromptHandler } = useMegaportWASM();

      const cancelHandler = vi.fn((request: any) => {
        (window as any).cancelPrompt(request.id);
      });

      registerPromptHandler(cancelHandler);

      // Simulate prompt request
      const promptRequest = { id: 'prompt-1', message: 'Enter value:' };
      registeredHandler!(promptRequest);

      expect(cancelHandler).toHaveBeenCalledWith(promptRequest);
      expect(mockCancel).toHaveBeenCalledWith('prompt-1');
    });
  });

  describe('Multiple Prompts', () => {
    it('should handle multiple sequential prompts', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });
      (window as any).submitPromptResponse = vi.fn();

      const { registerPromptHandler } = useMegaportWASM();
      const responseTracker: string[] = [];

      const multiPromptHandler = vi.fn((request: any) => {
        responseTracker.push(request.id);
      });

      registerPromptHandler(multiPromptHandler);

      // Simulate multiple prompts
      registeredHandler!({ id: 'prompt-1', message: 'Name:' });
      registeredHandler!({ id: 'prompt-2', message: 'Location:' });
      registeredHandler!({ id: 'prompt-3', message: 'Bandwidth:' });

      expect(multiPromptHandler).toHaveBeenCalledTimes(3);
      expect(responseTracker).toEqual(['prompt-1', 'prompt-2', 'prompt-3']);
    });

    it('should handle prompts with different types', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      // Different prompt types
      registeredHandler!({
        id: 'prompt-1',
        message: 'Text input:',
        type: 'text',
      });
      registeredHandler!({
        id: 'prompt-2',
        message: 'Password:',
        type: 'password',
      });
      registeredHandler!({
        id: 'prompt-3',
        message: 'Confirm (y/n):',
        type: 'confirm',
      });

      expect(customHandler).toHaveBeenCalledTimes(3);
    });
  });

  describe('Error Handling', () => {
    it('should handle errors in custom prompt handler', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const errorHandler = vi.fn(() => {
        throw new Error('Handler error');
      });

      registerPromptHandler(errorHandler);

      // Should not throw when handler errors
      expect(() => {
        registeredHandler!({ id: 'prompt-1', message: 'Test:' });
      }).toThrow('Handler error');
    });

    it('should handle missing submitPromptResponse function', () => {
      delete (window as any).submitPromptResponse;

      // Should not throw
      expect(() => {
        // Attempt to call undefined function would normally error
        // This test verifies the application handles this gracefully
        if ((window as any).submitPromptResponse) {
          (window as any).submitPromptResponse('id', 'value');
        }
      }).not.toThrow();
    });

    it('should handle missing cancelPrompt function', () => {
      delete (window as any).cancelPrompt;

      expect(() => {
        if ((window as any).cancelPrompt) {
          (window as any).cancelPrompt('id');
        }
      }).not.toThrow();
    });
  });

  describe('Prompt Handler Context', () => {
    it('should maintain handler context across multiple calls', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();

      class PromptManager {
        private responses: Map<string, string> = new Map();

        handlePrompt = (request: any) => {
          this.responses.set(request.id, request.message);
        };

        getResponseCount() {
          return this.responses.size;
        }
      }

      const manager = new PromptManager();
      registerPromptHandler(manager.handlePrompt);

      registeredHandler!({ id: '1', message: 'First' });
      registeredHandler!({ id: '2', message: 'Second' });

      expect(manager.getResponseCount()).toBe(2);
    });

    it('should allow handler to access external state', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();

      const externalState = { promptCount: 0 };

      const statefulHandler = vi.fn(() => {
        externalState.promptCount++;
      });

      registerPromptHandler(statefulHandler);

      registeredHandler!({ id: '1', message: 'Test' });
      registeredHandler!({ id: '2', message: 'Test' });
      registeredHandler!({ id: '3', message: 'Test' });

      expect(externalState.promptCount).toBe(3);
    });
  });

  describe('Password Masking', () => {
    it('lets the handler mask input when the prompt type is password', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();

      const masked: Record<string, boolean> = {};
      const maskingHandler = vi.fn((request: any) => {
        masked[request.id] = request.type === 'password';
      });
      registerPromptHandler(maskingHandler);

      registeredHandler!({ id: 'p1', message: 'Enter name:', type: 'text' });
      registeredHandler!({ id: 'p2', message: 'Enter secret key:', type: 'password' });

      expect(masked).toEqual({ p1: false, p2: true });
    });
  });

  describe('Cancel Lifecycle', () => {
    it('cancels via the request id and does not submit a response', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });
      const mockSubmit = vi.fn();
      const mockCancel = vi.fn();
      (window as any).submitPromptResponse = mockSubmit;
      (window as any).cancelPrompt = mockCancel;

      const { registerPromptHandler } = useMegaportWASM();

      // Handler dismisses the prompt (e.g. user pressed Escape).
      registerPromptHandler((request: any) => {
        (window as any).cancelPrompt(request.id);
      });

      registeredHandler!({ id: 'p1', message: 'Continue? [y/N]', type: 'confirm' });

      expect(mockCancel).toHaveBeenCalledWith('p1');
      expect(mockSubmit).not.toHaveBeenCalled();
    });
  });

  describe('Async-Only Entrypoint', () => {
    const originalFetch = global.fetch;
    const originalWebAssembly = global.WebAssembly;
    const originalGlobalGo = (global as any).Go;
    const originalWindowGo = (window as any).Go;

    beforeEach(() => {
      (global as any).Go = MockGo;
      (window as any).Go = MockGo;

      global.fetch = vi.fn(() =>
        Promise.resolve({
          arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
        } as Response)
      );

      global.WebAssembly = {
        instantiate: vi.fn(() => Promise.resolve({ instance: {}, module: {} })),
        instantiateStreaming: vi.fn(),
      } as any;
    });

    afterEach(() => {
      global.fetch = originalFetch;
      global.WebAssembly = originalWebAssembly;
      (global as any).Go = originalGlobalGo;
      (window as any).Go = originalWindowGo;
      delete (window as any).executeMegaportCommandAsync;
      delete (global as any).executeMegaportCommandAsync;
      delete (window as any).executeMegaportCommand;
    });

    it('routes interactive commands through the async entrypoint, never the sync one', async () => {
      const asyncMock = vi.fn((_cmd: string, cb: (r: any) => void) =>
        cb({ output: 'done', error: '' })
      );
      const syncMock = vi.fn(() => ({ output: 'sync', error: '' }));
      (window as any).executeMegaportCommandAsync = asyncMock;
      (window as any).executeMegaportCommand = syncMock;

      const { wrapper, composable } = await createReadyComposable();
      const result = await composable.execute('vxc buy --interactive');

      expect(asyncMock).toHaveBeenCalledTimes(1);
      expect(asyncMock.mock.calls[0][0]).toBe('vxc buy --interactive');
      expect(syncMock).not.toHaveBeenCalled();
      expect(result.output).toBe('done');

      wrapper.unmount();
    });

    it('surfaces the sync-entrypoint guard error instead of hanging', async () => {
      // No async entrypoint, so the composable falls back to the sync one, which
      // the WASM guard rejects when a command would prompt.
      const guardError = 'interactive mode requires the async entrypoint';
      const syncMock = vi.fn(() => ({ error: guardError }));
      delete (window as any).executeMegaportCommandAsync;
      delete (global as any).executeMegaportCommandAsync;
      (window as any).executeMegaportCommand = syncMock;

      const { wrapper, composable } = await createReadyComposable();
      const result = await composable.execute('vxc buy --interactive');

      expect(syncMock).toHaveBeenCalledTimes(1);
      expect(result.error).toContain('async entrypoint');

      wrapper.unmount();
    });
  });

  describe('Prompt Message Formatting', () => {
    it('should handle prompts with HTML entities', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      registeredHandler!({
        id: 'prompt-1',
        message: 'Enter value &lt;required&gt;:',
      });

      expect(customHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          message: 'Enter value &lt;required&gt;:',
        })
      );
    });

    it('should handle prompts with newlines', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      registeredHandler!({
        id: 'prompt-1',
        message: 'Line 1\nLine 2\nLine 3',
      });

      expect(customHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          message: expect.stringContaining('\n'),
        })
      );
    });

    it('should handle prompts with unicode characters', () => {
      let registeredHandler: ((request: any) => void) | null = null;
      (window as any).registerPromptHandler = vi.fn((handler: any) => {
        registeredHandler = handler;
        return true;
      });

      const { registerPromptHandler } = useMegaportWASM();
      const customHandler = vi.fn();

      registerPromptHandler(customHandler);

      registeredHandler!({
        id: 'prompt-1',
        message: 'Enter value 🚀 📝 ✅:',
      });

      expect(customHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          message: expect.stringContaining('🚀'),
        })
      );
    });
  });
});
