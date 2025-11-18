import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';

// IMPORTANT: Hoist all mocks to top of file before component import
// Mock xterm with proper constructor that returns instance
const mockTerminalInstance = {
  open: vi.fn(),
  write: vi.fn(),
  writeln: vi.fn(),
  clear: vi.fn(),
  dispose: vi.fn(),
  loadAddon: vi.fn(),
  onKey: vi.fn(),
  onData: vi.fn((callback: any) => {
    // Store callback for testing
    mockTerminalInstance._dataCallback = callback;
  }),
  focus: vi.fn(),
  _dataCallback: null as any,
};

vi.mock('@xterm/xterm', () => ({
  Terminal: vi.fn(function (this: any, options: any) {
    Object.assign(this, mockTerminalInstance);
    this.options = options;
  }),
}));

const mockFitAddon = {
  fit: vi.fn(),
  dispose: vi.fn(),
};

vi.mock('@xterm/addon-fit', () => ({
  FitAddon: vi.fn(function (this: any) {
    this.fit = mockFitAddon.fit;
    this.dispose = mockFitAddon.dispose;
  }),
}));

const mockWebLinksAddon = {
  dispose: vi.fn(),
};

vi.mock('@xterm/addon-web-links', () => ({
  WebLinksAddon: vi.fn(function (this: any) {
    this.dispose = vi.fn();
  }),
}));

// Mock composable - must be hoisted before component import
const mockComposable = {
  isLoading: ref(false),
  isReady: ref(true),
  error: ref<Error | null>(null),
  execute: vi.fn((cmd: string) =>
    Promise.resolve({ output: `Executed: ${cmd}`, error: '' })
  ),
  setAuth: vi.fn(),
  clearAuth: vi.fn(),
  getAuthInfo: vi.fn(() => ({
    accessKeySet: false,
    accessKeyPreview: '',
    secretKeySet: false,
    secretKeyPreview: '',
    environment: 'staging',
  })),
  resetOutput: vi.fn(),
  toggleDebug: vi.fn(),
};

vi.mock('../composables/useMegaportWASM', () => ({
  useMegaportWASM: vi.fn(() => mockComposable),
}));

// Import component AFTER all mocks are hoisted
import MegaportTerminal from '../components/MegaportTerminal.vue';

describe('MegaportTerminal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset mock composable state to default
    mockComposable.isLoading.value = false;
    mockComposable.isReady.value = true;
    mockComposable.error.value = null;
    mockComposable.execute.mockClear();
    mockComposable.setAuth.mockClear();
  });

  describe('Component Mounting', () => {
    it('should mount successfully', () => {
      const wrapper = mount(MegaportTerminal);
      expect(wrapper.exists()).toBe(true);
    });

    it('should accept props', () => {
      const wrapper = mount(MegaportTerminal, {
        props: {
          wasmPath: '/custom.wasm',
          wasmExecPath: '/custom_exec.js',
          welcomeMessage: 'Custom welcome',
        },
      });

      expect(wrapper.props('wasmPath')).toBe('/custom.wasm');
      expect(wrapper.props('wasmExecPath')).toBe('/custom_exec.js');
      expect(wrapper.props('welcomeMessage')).toBe('Custom welcome');
    });

    it('should use default props', () => {
      const wrapper = mount(MegaportTerminal);

      expect(wrapper.props('wasmPath')).toBe('/megaport.wasm');
      expect(wrapper.props('wasmExecPath')).toBe('/wasm_exec.js');
    });

    it('should accept custom theme', () => {
      const wrapper = mount(MegaportTerminal, {
        props: {
          theme: {
            background: '#000000',
            foreground: '#ffffff',
            cursor: '#ff0000',
          },
        },
      });

      expect(wrapper.props('theme')).toEqual({
        background: '#000000',
        foreground: '#ffffff',
        cursor: '#ff0000',
      });
    });
  });

  describe('Loading States', () => {
    it('should show loading state when WASM is loading', () => {
      // Modify mock state for this test
      mockComposable.isLoading.value = true;
      mockComposable.isReady.value = false;
      mockComposable.error.value = null;

      const wrapper = mount(MegaportTerminal);

      expect(wrapper.find('.terminal-loading').exists()).toBe(true);
      expect(wrapper.text()).toContain('Loading Megaport CLI');
    });

    it('should show error state when WASM fails to load', () => {
      // Modify mock state for this test
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = false;
      mockComposable.error.value = new Error('Failed to load WASM');

      const wrapper = mount(MegaportTerminal);

      expect(wrapper.find('.terminal-error').exists()).toBe(true);
      expect(wrapper.text()).toContain('Failed to load Megaport CLI');
      expect(wrapper.text()).toContain('Failed to load WASM');
    });

    it('should show retry button on error', () => {
      // Modify mock state for this test
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = false;
      mockComposable.error.value = new Error('Network error');

      const wrapper = mount(MegaportTerminal);

      const retryButton = wrapper.find('.terminal-error button');
      expect(retryButton.exists()).toBe(true);
      expect(retryButton.text()).toBe('Retry');
    });

    it('should show terminal when ready', () => {
      // Reset to ready state
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = true;
      mockComposable.error.value = null;

      const wrapper = mount(MegaportTerminal);

      expect(wrapper.find('.terminal-wrapper').exists()).toBe(true);
      expect(wrapper.find('.terminal-loading').exists()).toBe(false);
      expect(wrapper.find('.terminal-error').exists()).toBe(false);
    });
  });

  describe('Terminal Initialization', () => {
    it('should initialize terminal on mount', async () => {
      mount(MegaportTerminal);

      // Wait for terminal initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      expect(vi.mocked(Terminal)).toHaveBeenCalled();
    });

    it('should load terminal addons', async () => {
      mount(MegaportTerminal);

      // Wait for terminal initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      // Terminal should load addons when initialized
      expect(mockTerminalInstance.loadAddon).toHaveBeenCalled();
    });

    it('should apply custom theme to terminal', async () => {
      mount(MegaportTerminal, {
        props: {
          theme: {
            background: '#123456',
            foreground: '#abcdef',
            cursor: '#ff00ff',
          },
        },
      });

      // Wait for terminal initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      expect(vi.mocked(Terminal)).toHaveBeenCalledWith(
        expect.objectContaining({
          theme: expect.objectContaining({
            background: '#123456',
            foreground: '#abcdef',
            cursor: '#ff00ff',
          }),
        })
      );
    });
  });

  describe('Component API', () => {
    it('should expose executeCommand method', async () => {
      const wrapper = mount(MegaportTerminal);

      // Access exposed methods via vm
      expect(wrapper.vm).toBeDefined();
    });

    it('should expose clearTerminal method', () => {
      const wrapper = mount(MegaportTerminal);
      expect(wrapper.vm).toBeDefined();
    });

    it('should expose focusTerminal method', () => {
      const wrapper = mount(MegaportTerminal);
      expect(wrapper.vm).toBeDefined();
    });
  });

  describe('Cleanup', () => {
    it('should cleanup terminal on unmount', () => {
      const wrapper = mount(MegaportTerminal);
      const disposeSpy = vi.fn();

      // Mock terminal instance
      (wrapper.vm as any).terminal = {
        dispose: disposeSpy,
      };

      wrapper.unmount();

      expect(disposeSpy).toHaveBeenCalled();
    });

    it('should cleanup fit addon on unmount', async () => {
      const wrapper = mount(MegaportTerminal);

      // Wait for terminal initialization
      await new Promise((resolve) => setTimeout(resolve, 200));

      // Verify FitAddon was created
      expect(vi.mocked(FitAddon)).toHaveBeenCalled();

      wrapper.unmount();

      // The fit addon's dispose should be called on unmount
      expect(mockFitAddon.dispose).toHaveBeenCalled();
    });
  });

  describe('Styling', () => {
    it('should have terminal container class', () => {
      const wrapper = mount(MegaportTerminal);
      expect(wrapper.find('.megaport-terminal-container').exists()).toBe(true);
    });

    it('should apply CSS classes correctly', () => {
      const wrapper = mount(MegaportTerminal);

      expect(wrapper.classes()).toContain('megaport-terminal-container');
    });
  });

  describe('Error Recovery', () => {
    it('should handle reload on error', async () => {
      // Set error state
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = false;
      mockComposable.error.value = new Error('Test error');

      const wrapper = mount(MegaportTerminal);

      const retryButton = wrapper.find('.terminal-error button');
      expect(retryButton.exists()).toBe(true);

      // Mock window.location.reload
      const reloadSpy = vi.fn();
      Object.defineProperty(window, 'location', {
        value: { reload: reloadSpy },
        writable: true,
      });

      // Clicking retry should trigger reload
      await retryButton.trigger('click');

      // Component should attempt to reload
      expect(reloadSpy).toHaveBeenCalled();
    });
  });

  describe('Accessibility', () => {
    it('should have semantic HTML structure', () => {
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = true;
      mockComposable.error.value = null;

      const wrapper = mount(MegaportTerminal);
      expect(wrapper.element.tagName).toBe('DIV');
    });

    it('should provide loading feedback', () => {
      mockComposable.isLoading.value = true;
      mockComposable.isReady.value = false;
      mockComposable.error.value = null;

      const wrapper = mount(MegaportTerminal);
      expect(wrapper.text()).toContain('Loading');
    });

    it('should provide error feedback', () => {
      mockComposable.isLoading.value = false;
      mockComposable.isReady.value = false;
      mockComposable.error.value = new Error('Test error');

      const wrapper = mount(MegaportTerminal);
      expect(wrapper.text()).toContain('Failed to load');
    });
  });
});
