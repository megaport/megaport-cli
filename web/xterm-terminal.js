/**
 * XTerm.js Terminal Integration for Megaport CLI
 *
 * This module provides a full-featured terminal experience using xterm.js
 * with support for ANSI colors, proper table rendering, and terminal features.
 *
 * Note: xterm.js and addons are loaded via script tags in HTML as UMD modules
 * They are available as window.Terminal, window.FitAddon, and window.WebLinksAddon
 */

class XTerminalManager {
  constructor() {
    this.terminal = null;
    this.fitAddon = null;
    this.currentLine = '';
    this.commandHistory = [];
    this.historyIndex = -1;
    this.promptText = 'megaport> ';
    this.initialized = false;
  }

  /**
   * Initialize the xterm.js terminal
   */
  init(container) {
    if (this.initialized) {
      console.warn('Terminal already initialized');
      return;
    }

    try {
      // Ensure Terminal is available from global scope
      if (!window.Terminal) {
        throw new Error(
          'XTerm.js Terminal not loaded. Ensure xterm.js script is loaded before this module.'
        );
      }

      // Create terminal with Megaport theme and full ANSI support
      this.terminal = new window.Terminal({
        cursorBlink: true,
        cursorStyle: 'block',
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        fontSize: 14,
        lineHeight: 1.2,
        convertEol: false, // Don't auto-convert EOL, handle ANSI as-is
        drawBoldTextInBrightColors: true, // Enable bright colors for bold text
        theme: {
          background: '#000000',
          foreground: '#ffffff', // White for primary text
          cursor: '#33ff33',
          cursorAccent: '#000000',
          selection: 'rgba(51, 255, 51, 0.3)',
          black: '#4d4d4d', // Medium grey - readable against black
          red: '#ff5555',
          green: '#50fa7b',
          yellow: '#f1fa8c',
          blue: '#66d9ef',
          magenta: '#c30048', // Megaport red
          cyan: '#8be9fd',
          white: '#f8f8f2',
          brightBlack: '#b0b0b0', // Very bright grey for secondary text
          brightRed: '#ff6e6e',
          brightGreen: '#69ff94',
          brightYellow: '#ffffa5',
          brightBlue: '#7ee2ff',
          brightMagenta: '#ff79c6',
          brightCyan: '#a4ffff',
          brightWhite: '#ffffff',
        },
        scrollback: 10000,
        allowProposedApi: true,
        allowTransparency: false,
        windowOptions: {},
      });

      // Load addons (loaded as UMD modules via script tags)
      if (window.FitAddon) {
        this.fitAddon = new window.FitAddon.FitAddon();
        this.terminal.loadAddon(this.fitAddon);
      }
      if (window.WebLinksAddon) {
        this.terminal.loadAddon(new window.WebLinksAddon.WebLinksAddon());
      }

      // Open terminal in container
      this.terminal.open(container);

      // Fit terminal to container
      this.fitAddon.fit();

      // Handle window resize
      window.addEventListener('resize', () => {
        if (this.fitAddon) {
          this.fitAddon.fit();
        }
      });

      // Setup input handling
      this.setupInputHandling();

      this.initialized = true;
      console.log('✅ XTerm.js terminal initialized');

      return this.terminal;
    } catch (error) {
      console.error('❌ Failed to initialize XTerm.js terminal:', error);
      throw error;
    }
  }

  /**
   * Setup input handling for the terminal
   */
  setupInputHandling() {
    this.terminal.onData((data) => {
      const code = data.charCodeAt(0);

      // Handle Enter key
      if (code === 13) {
        this.handleCommand(this.currentLine.trim());
        this.currentLine = '';
        return;
      }

      // Handle Backspace
      if (code === 127 || code === 8) {
        if (this.currentLine.length > 0) {
          this.currentLine = this.currentLine.slice(0, -1);
          this.terminal.write('\b \b');
        }
        return;
      }

      // Handle Ctrl+C
      if (code === 3) {
        this.terminal.write('^C\r\n');
        this.currentLine = '';
        this.writePrompt();
        return;
      }

      // Handle Ctrl+L (clear screen)
      if (code === 12) {
        // Clear scrollback and screen completely
        this.terminal.clear();
        this.terminal.reset();
        this.currentLine = '';
        // Write a single clean prompt
        this.terminal.write(this.promptText);
        return;
      }

      // Handle arrow up (history)
      if (data === '\x1b[A') {
        this.navigateHistory('up');
        return;
      }

      // Handle arrow down (history)
      if (data === '\x1b[B') {
        this.navigateHistory('down');
        return;
      }

      // Handle regular characters
      if (code >= 32 && code <= 126) {
        this.currentLine += data;
        this.terminal.write(data);
      }
    });
  }

  /**
   * Navigate command history
   */
  navigateHistory(direction) {
    if (this.commandHistory.length === 0) return;

    if (direction === 'up') {
      if (this.historyIndex < this.commandHistory.length - 1) {
        this.historyIndex++;
      }
    } else {
      if (this.historyIndex > -1) {
        this.historyIndex--;
      }
    }

    // Clear current line
    this.terminal.write('\r' + this.promptText);
    this.terminal.write(' '.repeat(this.currentLine.length));
    this.terminal.write('\r' + this.promptText);

    // Show history item
    if (this.historyIndex >= 0) {
      const historyCommand =
        this.commandHistory[this.commandHistory.length - 1 - this.historyIndex];
      this.currentLine = historyCommand;
      this.terminal.write(historyCommand);
    } else {
      this.currentLine = '';
    }
  }

  /**
   * Handle command execution
   */
  handleCommand(command) {
    this.terminal.write('\r\n');

    if (!command) {
      this.writePrompt();
      return;
    }

    // Add to history
    this.commandHistory.push(command);
    this.historyIndex = -1;

    // Execute command
    try {
      if (typeof window.executeMegaportCommandAsync === 'function') {
        // Use callback-based async execution (required for WASM)
        window.executeMegaportCommandAsync(command, (result) => {
          if (result && result.output) {
            // XTerm.js will automatically handle ANSI codes
            this.write(result.output);
          } else if (result && result.error) {
            this.writeError('Error: ' + result.error);
          }
          this.writePrompt();
        });
      } else if (typeof window.executeMegaportCommand === 'function') {
        const result = window.executeMegaportCommand(command);

        if (result && result.output) {
          this.write(result.output);
        } else if (result && result.error) {
          this.writeError('Error: ' + result.error);
        }
        this.writePrompt();
      } else {
        this.writeError('Error: Command execution not available');
        this.writePrompt();
      }
    } catch (err) {
      this.writeError('Error: ' + err.message);
      console.error('Command execution error:', err);
      this.writePrompt();
    }
  }

  /**
   * Write text to terminal (supports ANSI codes)
   */
  write(text) {
    if (!this.terminal) return;

    // Debug: Log ANSI codes in the output
    if (text.includes('\x1b[')) {
      console.log('🎨 ANSI codes detected in output:', text.substring(0, 200));
      // Show hex representation of first 100 chars for debugging
      const sample = text.substring(0, 100);
      const hex = Array.from(sample)
        .map((c) => c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join(' ');
      console.log('   Hex:', hex);
    }

    // XTerm.js handles ANSI codes automatically
    // Convert \n to \r\n for proper terminal display
    text = text.replace(/\n/g, '\r\n');

    // Remove any duplicate \r\n\r\n sequences
    text = text.replace(/\r\n\r\n/g, '\r\n');

    this.terminal.write(text);
  }

  /**
   * Write error message
   */
  writeError(message) {
    this.terminal.write('\x1b[31m' + message + '\x1b[0m\r\n');
  }

  /**
   * Write success message
   */
  writeSuccess(message) {
    this.terminal.write('\x1b[32m' + message + '\x1b[0m\r\n');
  }

  /**
   * Write info message
   */
  writeInfo(message) {
    this.terminal.write('\x1b[36m' + message + '\x1b[0m\r\n');
  }

  /**
   * Write warning message
   */
  writeWarning(message) {
    this.terminal.write('\x1b[33m' + message + '\x1b[0m\r\n');
  }

  /**
   * Write the command prompt
   */
  writePrompt() {
    this.terminal.write(this.promptText);
  }

  /**
   * Clear the terminal
   */
  clear() {
    if (this.terminal) {
      this.terminal.clear();
      this.terminal.reset();
      this.currentLine = '';
      this.terminal.write(this.promptText);
    }
  }

  /**
   * Focus the terminal
   */
  focus() {
    if (this.terminal) {
      this.terminal.focus();
    }
  }

  /**
   * Resize terminal to fit container
   */
  fit() {
    if (this.fitAddon) {
      this.fitAddon.fit();
    }
  }

  /**
   * Dispose terminal
   */
  dispose() {
    if (this.terminal) {
      this.terminal.dispose();
      this.terminal = null;
      this.initialized = false;
    }
  }
}

// Create singleton instance
export const xtermManager = new XTerminalManager();

// Export for global access
if (typeof window !== 'undefined') {
  window.xtermManager = xtermManager;
  console.log('✅ XTerm.js manager available globally');
}

export default xtermManager;
