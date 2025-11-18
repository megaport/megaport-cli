/** * Vue 3 Component: Megaport CLI Terminal * Provides an interactive terminal
interface using xterm.js */

<template>
  <div class="megaport-terminal-container">
    <!-- Loading State -->
    <div v-if="isLoading" class="terminal-loading">
      <div class="spinner"></div>
      <p>Loading Megaport CLI...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="hasError" class="terminal-error">
      <h3>❌ Failed to load Megaport CLI</h3>
      <p>{{ displayError?.message }}</p>
      <button @click="reload">Retry</button>
    </div>

    <!-- Terminal with Spinner Overlay -->
    <div v-else class="terminal-container">
      <!-- Active Spinner Overlay -->
      <div
        v-if="activeSpinners && activeSpinners.size > 0"
        class="spinner-overlay"
      >
        <div class="spinner-content">
          <div class="spinner"></div>
          <p v-for="[id, message] in activeSpinners" :key="id">
            {{ message }}
          </p>
        </div>
      </div>

      <!-- Terminal -->
      <div ref="terminalRef" class="terminal-wrapper"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
/// <reference path="../vite-env.d.ts" />
import {
  ref,
  computed,
  onMounted,
  onBeforeUnmount,
  onErrorCaptured,
} from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { useMegaportWASM } from '../composables/useMegaportWASM';
import { TERMINAL_CONFIG, WASM_CONFIG } from '../constants/megaportWASM';
import type { MegaportPromptRequest } from '../types/megaport-wasm';

// Type augmentation for window methods
declare global {
  interface Window {
    submitPromptResponse?: (id: string, response: string) => void;
    cancelPrompt?: (id: string) => void;
  }
}

export interface MegaportTerminalProps {
  wasmPath?: string;
  wasmExecPath?: string;
  welcomeMessage?: string;
  theme?: {
    background?: string;
    foreground?: string;
    cursor?: string;
  };
}

const props = withDefaults(defineProps<MegaportTerminalProps>(), {
  wasmPath: '/megaport.wasm',
  wasmExecPath: '/wasm_exec.js',
  welcomeMessage:
    'Welcome to Megaport CLI (WebAssembly)\nType "help" for available commands.\n',
  theme: () => ({
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    cursor: '#aeafad',
  }),
});

// Terminal setup
const terminalRef = ref<HTMLElement | null>(null);
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;
let currentLine = '';
let cursorPosition = 0;
let justCleared = false; // Track if terminal was just cleared

// Command history
const commandHistory = ref<string[]>([]);
let historyIndex = -1; // Current position in history (-1 = not browsing)

// WASM integration
const {
  isLoading,
  isReady,
  error,
  execute,
  setAuth,
  registerPromptHandler,
  activeSpinners,
} = useMegaportWASM({
  wasmPath: props.wasmPath,
  wasmExecPath: props.wasmExecPath,
  debug: true,
});

// Local error state for component-level errors
const componentError = ref<Error | null>(null);

// Computed to combine WASM and component errors
const hasError = computed(() => error.value || componentError.value);
const displayError = computed(() => componentError.value || error.value);

// Prompt handling state
let activePrompt: { id: string; resolve: (value: string) => void } | null =
  null;
let promptInputBuffer = '';
let isInInteractiveCommand = false; // Track if we're in an interactive command session
let resizeTimeoutId: NodeJS.Timeout | null = null; // For debouncing resize

/**
 * Debounce utility function
 */
const debounce = (fn: Function, ms: number) => {
  return (...args: any[]) => {
    if (resizeTimeoutId) {
      clearTimeout(resizeTimeoutId);
    }
    resizeTimeoutId = setTimeout(() => fn(...args), ms);
  };
};

/**
 * Error boundary handler - captures component-level errors
 */
onErrorCaptured((err, instance, info) => {
  console.error('❌ Terminal component error:', err);
  console.error('Error context:', info);

  // Set local error state to display error UI
  componentError.value = err instanceof Error ? err : new Error(String(err));

  // Prevent error from propagating to parent components
  return false;
});

/**
 * Register inline terminal prompt handler
 */
const setupPromptHandler = () => {
  if (!registerPromptHandler || typeof registerPromptHandler !== 'function') {
    console.warn(
      'registerPromptHandler not available - WASM may not be initialized'
    );
    return;
  }
  registerPromptHandler((promptRequest: MegaportPromptRequest) => {
    if (!terminal) return;

    console.log(
      '🔔 Prompt handler called:',
      promptRequest.id,
      promptRequest.message
    );

    // Display the prompt message in terminal style
    terminal.write(`\r\n\x1b[36m${promptRequest.message}\x1b[0m `);

    // Track this prompt
    activePrompt = {
      id: promptRequest.id,
      resolve: (response: string) => {
        if (window.submitPromptResponse) {
          window.submitPromptResponse(promptRequest.id, response);
        }
      },
    };

    console.log('✅ activePrompt set:', activePrompt.id);

    promptInputBuffer = '';
  });
};

/**
 * Lazy load xterm CSS only when needed
 */
const loadXtermCSS = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    // Check if already loaded
    if (document.querySelector('link[href*="xterm.css"]')) {
      resolve();
      return;
    }

    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://cdn.jsdelivr.net/npm/@xterm/xterm@5.5.0/css/xterm.css';
    link.onload = () => resolve();
    link.onerror = () => reject(new Error('Failed to load xterm CSS'));
    document.head.appendChild(link);
  });
};

/**
 * Initialize xterm.js terminal
 */
const initTerminal = async () => {
  if (!terminalRef.value) return;

  // Lazy load xterm CSS first
  try {
    await loadXtermCSS();
  } catch (err) {
    console.error('Failed to load xterm CSS:', err);
    componentError.value = err instanceof Error ? err : new Error(String(err));
    return;
  }

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: TERMINAL_CONFIG.FONT_SIZE,
    fontFamily: TERMINAL_CONFIG.FONT_FAMILY,
    theme: {
      background: props.theme.background,
      foreground: props.theme.foreground,
      cursor: props.theme.cursor,
    },
  });

  // Add addons
  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.loadAddon(new WebLinksAddon());

  // Check if terminalRef still exists (component might have unmounted)
  if (!terminalRef.value) {
    terminal?.dispose();
    fitAddon?.dispose();
    return;
  }

  // Open terminal
  terminal.open(terminalRef.value);
  fitAddon.fit();

  // Display welcome message
  terminal.write(props.welcomeMessage);
  // Don't set justCleared here - welcome message already has proper newlines
  writePrompt();

  // Handle input
  terminal.onData((data: string) => {
    handleInput(data);
  });

  // Handle resize with debounce to prevent excessive re-calculations
  const handleResize = debounce(() => {
    fitAddon?.fit();
  }, TERMINAL_CONFIG.RESIZE_DEBOUNCE_DELAY);

  window.addEventListener('resize', handleResize);
};

/**
 * Write command prompt
 */
const writePrompt = () => {
  // After clear, just write the prompt (cursor is already at home)
  // Otherwise, add newline and carriage return to start fresh line
  if (justCleared) {
    terminal?.write('\x1b[32mmegaport>\x1b[0m ');
    justCleared = false;
  } else {
    terminal?.write('\r\n\x1b[32mmegaport>\x1b[0m ');
  }
  currentLine = '';
  cursorPosition = 0;
  historyIndex = -1; // Reset history index when showing new prompt
};

/**
 * Handle prompt input when in interactive mode
 */
const handlePromptInput = (data: string, code: number): boolean => {
  if (!terminal || !activePrompt) return false;

  // Enter key - submit prompt response
  if (code === 13) {
    terminal.write('\r\n');
    const response = promptInputBuffer;

    // Submit the response
    activePrompt.resolve(response);

    // Clear the prompt buffer but DON'T clear activePrompt yet
    // The next prompt will overwrite it, or command completion will clear it
    promptInputBuffer = '';
    return true;
  }

  // Backspace
  if (code === 127) {
    if (promptInputBuffer.length > 0) {
      promptInputBuffer = promptInputBuffer.slice(0, -1);
      terminal.write('\b \b');
    }
    return true;
  }

  // Ctrl+C - cancel prompt
  if (code === 3) {
    terminal.write('^C\r\n');
    if (window.cancelPrompt && activePrompt) {
      window.cancelPrompt(activePrompt.id);
    }
    activePrompt = null;
    promptInputBuffer = '';
    writePrompt();
    return true;
  }

  // Regular character for prompt
  if (code >= 32 && code <= 126) {
    promptInputBuffer += data;
    terminal.write(data);
    return true;
  }

  return true;
};

/**
 * Handle arrow key navigation
 */
const handleArrowKeys = (data: string): boolean => {
  if (!terminal) return false;

  // Left arrow
  if (data === '\x1b[D') {
    if (cursorPosition > 0) {
      cursorPosition--;
      terminal.write('\x1b[D');
    }
    return true;
  }

  // Right arrow
  if (data === '\x1b[C') {
    if (cursorPosition < currentLine.length) {
      cursorPosition++;
      terminal.write('\x1b[C');
    }
    return true;
  }

  // Up arrow - navigate history backwards
  if (data === '\x1b[A') {
    if (historyIndex < commandHistory.value.length - 1) {
      historyIndex++;
      const historicalCommand =
        commandHistory.value[commandHistory.value.length - 1 - historyIndex];

      // Clear current line
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[32mmegaport>\x1b[0m ');

      // Write historical command
      terminal.write(historicalCommand);
      currentLine = historicalCommand;
      cursorPosition = currentLine.length;
    }
    return true;
  }

  // Down arrow - navigate history forwards
  if (data === '\x1b[B') {
    if (historyIndex > 0) {
      historyIndex--;
      const historicalCommand =
        commandHistory.value[commandHistory.value.length - 1 - historyIndex];

      // Clear current line
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[32mmegaport>\x1b[0m ');

      // Write historical command
      terminal.write(historicalCommand);
      currentLine = historicalCommand;
      cursorPosition = currentLine.length;
    } else if (historyIndex === 0) {
      // Return to empty line
      historyIndex = -1;
      terminal.write('\r\x1b[K');
      terminal.write('\x1b[32mmegaport>\x1b[0m ');
      currentLine = '';
      cursorPosition = 0;
    }
    return true;
  }

  return false;
};

/**
 * Handle control keys (Ctrl+C, Ctrl+L, etc.)
 */
const handleControlKeys = (code: number): boolean => {
  if (!terminal) return false;

  // Ctrl+C
  if (code === 3) {
    terminal.write('^C');
    writePrompt();
    return true;
  }

  // Ctrl+L (clear)
  if (code === 12) {
    terminal.clear();
    terminal.write('\x1b[H'); // Move cursor to home position
    justCleared = true;
    writePrompt();
    return true;
  }

  return false;
};

/**
 * Handle terminal input - delegates to specialized helper functions
 */
const handleInput = (data: string) => {
  if (!terminal) return;

  const code = data.charCodeAt(0);

  console.log(
    '⌨️ Input:',
    data,
    'code:',
    code,
    'activePrompt:',
    !!activePrompt,
    'isInInteractiveCommand:',
    isInInteractiveCommand
  );

  // Handle prompt mode input
  if (activePrompt) {
    handlePromptInput(data, code);
    return;
  }

  // If we're in an interactive command but not in an active prompt,
  // ignore input (waiting for next prompt)
  if (isInInteractiveCommand && !activePrompt) {
    return;
  }

  // Handle control keys first
  if (handleControlKeys(code)) {
    return;
  }

  // Handle arrow keys for navigation and history
  if (handleArrowKeys(data)) {
    return;
  }

  // Enter key - execute command
  if (code === 13) {
    terminal.write('\r\n');
    if (currentLine.trim()) {
      const commandToExecute = currentLine.trim();
      currentLine = '';
      cursorPosition = 0;
      executeCommand(commandToExecute);
    } else {
      writePrompt();
    }
    return;
  }

  // Backspace
  if (code === 127) {
    if (cursorPosition > 0) {
      currentLine =
        currentLine.slice(0, cursorPosition - 1) +
        currentLine.slice(cursorPosition);
      cursorPosition--;
      terminal.write('\b \b');
    }
    return;
  }

  // Regular character input
  if (code >= 32 && code <= 126) {
    currentLine =
      currentLine.slice(0, cursorPosition) +
      data +
      currentLine.slice(cursorPosition);
    cursorPosition++;
    terminal.write(data);
  }
};
/**
 * Execute a CLI command
 */
const executeCommand = async (command: string) => {
  if (!terminal || !isReady.value) {
    terminal?.write('\x1b[31mCLI not ready\x1b[0m');
    writePrompt();
    return;
  }

  // Add to command history (skip empty commands and duplicates)
  if (command.trim()) {
    // Don't add if same as last command
    if (
      commandHistory.value.length === 0 ||
      commandHistory.value[commandHistory.value.length - 1] !== command
    ) {
      commandHistory.value.push(command);

      // Limit history size
      if (commandHistory.value.length > TERMINAL_CONFIG.MAX_HISTORY_SIZE) {
        commandHistory.value.shift();
      }
    }
  }

  try {
    // Handle built-in commands
    if (command === 'clear' || command === 'cls') {
      terminal.clear();
      terminal.write('\x1b[H'); // Move cursor to home position
      justCleared = true;
      writePrompt();
      return;
    }

    if (command === 'help') {
      terminal.write('Available commands:\r\n');
      terminal.write('  port list           - List all ports\r\n');
      terminal.write('  vxc list            - List all VXCs\r\n');
      terminal.write('  mcr list            - List all MCRs\r\n');
      terminal.write('  mve list            - List all MVEs\r\n');
      terminal.write('  location list       - List all locations\r\n');
      terminal.write('  servicekey list     - List all service keys\r\n');
      terminal.write('  partner list        - List partner configurations\r\n');
      terminal.write('  clear               - Clear the terminal\r\n');
      writePrompt();
      return;
    }

    // Execute WASM command
    // Don't show "Executing..." for interactive commands as prompts will appear immediately
    const isInteractive =
      command.includes('--interactive') || command.includes('-i');

    if (isInteractive) {
      isInInteractiveCommand = true;
    } else {
      terminal.write('\x1b[90mExecuting...\x1b[0m\r\n');
    }

    const result = await execute(command);

    if (result.error) {
      terminal.write(`\x1b[31mError: ${result.error}\x1b[0m\r\n`);
    } else if (result.output) {
      // For interactive commands, filter out ONLY prompt messages from output
      // Keep all other output like progress indicators, success messages, etc.
      let outputToDisplay = result.output;

      if (isInteractive) {
        // Remove ONLY the prompt text patterns that were already displayed via prompt handler
        // Keep everything else (progress messages, results, etc.)
        const promptPatterns = [
          /^Enter port name \(required\): ?\n?/gm,
          /^Enter term \([^)]+\) \(required\): ?\n?/gm,
          /^Enter port speed \([^)]+\) \(required\): ?\n?/gm,
          /^Enter location ID \(required\): ?\n?/gm,
          /^Enter marketplace visibility \([^)]+\) \(required\): ?\n?/gm,
          /^Enter diversity zone \(optional\): ?\n?/gm,
          /^Enter cost centre \(optional\): ?\n?/gm,
          /^Enter promo code \(optional\): ?\n?/gm,
          /^Would you like to add resource tags\? \[y\/N\] ?\n?/gm,
          /^Tag key \([^)]+\): ?\n?/gm,
          /^Tag value for '[^']+': ?\n?/gm,
        ];

        promptPatterns.forEach((pattern) => {
          outputToDisplay = outputToDisplay.replace(pattern, '');
        });

        // Clean up multiple consecutive newlines
        outputToDisplay = outputToDisplay.replace(/\n{3,}/g, '\n\n');
      }

      // Format and display output
      if (outputToDisplay.trim()) {
        const lines = outputToDisplay.split('\n');
        lines.forEach((line) => {
          if (terminal) {
            terminal.write(line + '\r\n');
          }
        });
      }
    } else if (!isInteractive) {
      // Only show "no output" message for non-interactive commands
      terminal.write('\x1b[90mCommand completed with no output\x1b[0m\r\n');
    }
  } catch (err) {
    if (terminal) {
      terminal.write(`\x1b[31mError: ${(err as Error).message}\x1b[0m\r\n`);
    }
  }

  // Reset interactive command flag when command completes
  isInInteractiveCommand = false;

  // Clear any lingering prompt state
  activePrompt = null;
  promptInputBuffer = '';

  // Only write a new prompt if we're not in an active prompt session
  if (!activePrompt) {
    writePrompt();
  }
};

/**
 * Reload the page
 */
const reload = () => {
  window.location.reload();
};

// Lifecycle
onMounted(() => {
  // Wait for WASM to be ready
  const checkReady = setInterval(() => {
    if (isReady.value) {
      clearInterval(checkReady);
      initTerminal(); // Now async but we don't need to await
      setupPromptHandler(); // Register inline prompt handler
    }
  }, 100);

  // Cleanup after 30 seconds if not ready
  setTimeout(() => clearInterval(checkReady), 30000);
});

onBeforeUnmount(() => {
  // Clear resize timeout if pending
  if (resizeTimeoutId) {
    clearTimeout(resizeTimeoutId);
  }

  fitAddon?.dispose();
  terminal?.dispose();
});

// Expose terminal instance for parent components
defineExpose({
  terminal,
  fitAddon,
  execute: executeCommand,
  setAuth,
});
</script>

<style scoped>
.megaport-terminal-container {
  width: 100%;
  height: 100%;
  background-color: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.terminal-loading,
.terminal-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #d4d4d4;
  padding: 2rem;
}

.spinner {
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-top: 4px solid #4a9eff;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.terminal-error h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.2rem;
}

.terminal-error p {
  margin: 0 0 1rem 0;
  color: #ff6b6b;
}

.terminal-error button {
  padding: 0.5rem 1rem;
  background-color: #4a9eff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.terminal-error button:hover {
  background-color: #3a8eef;
}

.terminal-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.spinner-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.spinner-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  background-color: #2a2a2a;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
}

.spinner-content p {
  margin: 0.5rem 0 0 0;
  color: #4a9eff;
  font-size: 0.9rem;
  text-align: center;
}

.terminal-wrapper {
  width: 100%;
  height: 100%;
  padding: 0.5rem;
}

:deep(.xterm) {
  height: 100% !important;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
