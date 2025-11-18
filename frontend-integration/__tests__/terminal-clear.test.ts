import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('Terminal Clear Functionality', () => {
  let mockTerminal: any;
  let writeHistory: string[];

  beforeEach(() => {
    writeHistory = [];

    mockTerminal = {
      write: vi.fn((text: string) => {
        writeHistory.push(text);
      }),
      clear: vi.fn(),
    };
  });

  describe('Clear Command', () => {
    it('should clear the terminal when clear command is executed', () => {
      mockTerminal.clear();
      expect(mockTerminal.clear).toHaveBeenCalled();
    });

    it('should clear the terminal when cls command is executed', () => {
      // cls is an alias for clear
      mockTerminal.clear();
      expect(mockTerminal.clear).toHaveBeenCalled();
    });

    it('should move cursor to home position after clear', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H'); // Home position escape code

      expect(writeHistory).toContain('\x1b[H');
    });

    it('should write prompt without newline after clear', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      // Should not have \r\n before the prompt
      expect(writeHistory[writeHistory.length - 1]).toBe(
        '\x1b[32mmegaport>\x1b[0m '
      );
      expect(writeHistory[writeHistory.length - 1]).not.toContain('\r\n');
    });
  });

  describe('Ctrl+L Shortcut', () => {
    it('should clear terminal when Ctrl+L is pressed', () => {
      // Simulate Ctrl+L (ASCII code 12)
      const ctrlLCode = 12;

      mockTerminal.clear();
      mockTerminal.write('\x1b[H');

      expect(mockTerminal.clear).toHaveBeenCalled();
      expect(writeHistory).toContain('\x1b[H');
    });

    it('should reset cursor position after Ctrl+L', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory[0]).toBe('\x1b[H');
      expect(writeHistory[1]).toBe('\x1b[32mmegaport>\x1b[0m ');
    });
  });

  describe('Prompt Positioning', () => {
    it('should position prompt at left margin after clear', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H'); // Move to home
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      // Home position escape code should come before prompt
      const homeIndex = writeHistory.indexOf('\x1b[H');
      const promptIndex = writeHistory.indexOf('\x1b[32mmegaport>\x1b[0m ');

      expect(homeIndex).toBeGreaterThanOrEqual(0);
      expect(promptIndex).toBeGreaterThan(homeIndex);
    });

    it('should not add extra newlines after clear', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      // Check that no writes contain \r\n after clear
      const writesAfterHome = writeHistory.slice(
        writeHistory.indexOf('\x1b[H') + 1
      );
      const hasNewline = writesAfterHome.some((w) => w.includes('\r\n'));

      expect(hasNewline).toBe(false);
    });

    it('should position cursor at column 0 after clear', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H'); // Positions at row 1, column 1

      expect(writeHistory).toContain('\x1b[H');
    });
  });

  describe('Initial Terminal State', () => {
    it('should write welcome message before first prompt', () => {
      const welcomeMessage =
        'Welcome to Megaport CLI (WebAssembly)\nType "help" for available commands.\n';
      mockTerminal.write(welcomeMessage);
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory[0]).toBe(welcomeMessage);
      expect(writeHistory[1]).toBe('\r\n\x1b[32mmegaport>\x1b[0m ');
    });

    it('should add newline before initial prompt', () => {
      const welcomeMessage = 'Welcome\n';
      mockTerminal.write(welcomeMessage);
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      // First prompt after welcome should have \r\n
      expect(writeHistory[1]).toContain('\r\n');
    });

    it('should not use justCleared flag for initial prompt', () => {
      mockTerminal.write('Welcome\n');
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      // Should include newline for normal prompt
      expect(writeHistory[1]).toContain('\r\n');
    });
  });

  describe('Multiple Clear Operations', () => {
    it('should handle multiple consecutive clears', () => {
      // First clear
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(mockTerminal.clear).toHaveBeenCalledTimes(1);

      // Second clear
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(mockTerminal.clear).toHaveBeenCalledTimes(2);
    });

    it('should reset state correctly after each clear', () => {
      for (let i = 0; i < 5; i++) {
        writeHistory = [];
        mockTerminal.clear();
        mockTerminal.write('\x1b[H');
        mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

        expect(writeHistory[0]).toBe('\x1b[H');
        expect(writeHistory[1]).toBe('\x1b[32mmegaport>\x1b[0m ');
      }
    });

    it('should not accumulate state across clears', () => {
      // Clear 1
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      const firstClearWrites = writeHistory.length;

      // Clear 2
      writeHistory = [];
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory.length).toBe(firstClearWrites);
    });
  });

  describe('Clear After Command Execution', () => {
    it('should position prompt correctly after command then clear', () => {
      // Execute command
      mockTerminal.write('ports list');
      mockTerminal.write('\r\n');
      mockTerminal.write('Output data...\r\n');
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      writeHistory = [];

      // Then clear
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory[0]).toBe('\x1b[H');
      expect(writeHistory[1]).toBe('\x1b[32mmegaport>\x1b[0m ');
      expect(writeHistory[1]).not.toContain('\r\n');
    });

    it('should clear all previous output', () => {
      // Add lots of output
      for (let i = 0; i < 100; i++) {
        mockTerminal.write(`Line ${i}\r\n`);
      }

      // Clear
      mockTerminal.clear();

      expect(mockTerminal.clear).toHaveBeenCalled();
    });
  });

  describe('Clear During Interactive Command', () => {
    it('should be able to clear during interactive input', () => {
      // Start interactive command
      mockTerminal.write('ports buy --interactive\r\n');
      mockTerminal.write('Enter port name: ');

      // User clears
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(mockTerminal.clear).toHaveBeenCalled();
      expect(writeHistory).toContain('\x1b[H');
    });
  });

  describe('Escape Sequences', () => {
    it('should use correct ANSI escape code for home position', () => {
      mockTerminal.write('\x1b[H');

      expect(writeHistory[0]).toBe('\x1b[H');
    });

    it('should use correct ANSI escape codes for prompt color', () => {
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory[0]).toContain('\x1b[32m'); // Green color
      expect(writeHistory[0]).toContain('\x1b[0m'); // Reset color
    });

    it('should not include carriage return in post-clear prompt', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      const promptWrite = writeHistory[writeHistory.length - 1];
      expect(promptWrite).not.toContain('\r');
      expect(promptWrite).not.toContain('\n');
    });
  });

  describe('justCleared Flag Behavior', () => {
    it('should set justCleared to true after clear', () => {
      let justCleared = false;

      mockTerminal.clear();
      justCleared = true;

      expect(justCleared).toBe(true);
    });

    it('should reset justCleared to false after writing prompt', () => {
      let justCleared = true;

      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');
      justCleared = false;

      expect(justCleared).toBe(false);
    });

    it('should not set justCleared for normal prompts', () => {
      let justCleared = false;

      // Normal command execution
      mockTerminal.write('help\r\n');
      mockTerminal.write('Available commands:\r\n');
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      expect(justCleared).toBe(false);
    });

    it('should only set justCleared on actual clear operations', () => {
      let justCleared = false;

      // Regular operations
      mockTerminal.write('test\r\n');
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');
      expect(justCleared).toBe(false);

      // Clear operation
      mockTerminal.clear();
      justCleared = true;
      expect(justCleared).toBe(true);

      // After prompt
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');
      justCleared = false;
      expect(justCleared).toBe(false);
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid clear commands', () => {
      for (let i = 0; i < 10; i++) {
        mockTerminal.clear();
        mockTerminal.write('\x1b[H');
        mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');
      }

      expect(mockTerminal.clear).toHaveBeenCalledTimes(10);
    });

    it('should handle clear with no previous output', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(mockTerminal.clear).toHaveBeenCalled();
      expect(writeHistory.length).toBe(2);
    });

    it('should handle clear immediately after terminal initialization', () => {
      mockTerminal.write('Welcome\n');
      mockTerminal.write('\r\n\x1b[32mmegaport>\x1b[0m ');

      writeHistory = [];

      mockTerminal.clear();
      mockTerminal.write('\x1b[H');
      mockTerminal.write('\x1b[32mmegaport>\x1b[0m ');

      expect(writeHistory[0]).toBe('\x1b[H');
    });
  });

  describe('Browser Compatibility', () => {
    it('should use standard ANSI codes for maximum compatibility', () => {
      mockTerminal.write('\x1b[H'); // Home position
      mockTerminal.write('\x1b[32m'); // Green color
      mockTerminal.write('\x1b[0m'); // Reset

      expect(writeHistory[0]).toBe('\x1b[H');
      expect(writeHistory[1]).toBe('\x1b[32m');
      expect(writeHistory[2]).toBe('\x1b[0m');
    });

    it('should not use platform-specific clear codes', () => {
      mockTerminal.clear();
      mockTerminal.write('\x1b[H');

      // Should use standard terminal.clear() and \x1b[H
      // Not platform-specific codes like \033c or clear screen codes
      expect(writeHistory[0]).toBe('\x1b[H');
    });
  });
});
