/**
 * Terminal Output Handler with Charsm Integration
 *
 * This module handles terminal output rendering, with support for:
 * - Charsm-styled output (when available)
 * - Plain text fallback
 * - Table rendering with Charsm
 * - JSON formatting
 * - ANSI code stripping
 */

import { charsmRenderer } from './charsm-renderer.js';

class TerminalOutputHandler {
  constructor() {
    this.terminal = null;
    this.useCharsm = true;
    this.initialized = false;
  }

  /**
   * Initialize the terminal output handler
   */
  init(terminalElement) {
    this.terminal = terminalElement;
    this.initialized = true;
    console.log('✅ Terminal output handler initialized');
  }

  /**
   * Append text to terminal with optional styling
   */
  appendToTerminal(text, className) {
    if (!this.terminal) {
      console.error('Terminal element not initialized');
      return;
    }

    // Handle empty or whitespace-only text
    if (!text || text.trim() === '') return;

    // Split by lines and process each
    const lines = text.split('\n');

    lines.forEach((line) => {
      if (line.trim() === '') return;

      const lineElement = document.createElement('div');

      // Apply Charsm styling if enabled and available
      if (this.useCharsm && charsmRenderer.initialized) {
        lineElement.innerHTML = this._processWithCharsm(line, className);
      } else {
        lineElement.innerHTML = this._processPlain(line);
      }

      if (className) {
        lineElement.className = className;
      }

      this.terminal.appendChild(lineElement);
    });

    this.terminal.scrollTop = this.terminal.scrollHeight;
  }

  /**
   * Process text with Charsm styling
   */
  _processWithCharsm(text, className) {
    // Detect output type
    if (this._isTableOutput(text)) {
      return this._renderTableOutput(text);
    }

    if (this._isJSONOutput(text)) {
      try {
        const jsonData = JSON.parse(text);
        return charsmRenderer.renderJSON(jsonData);
      } catch (e) {
        // Not valid JSON, continue with regular processing
      }
    }

    // Apply style based on className
    let styledText = text;
    switch (className) {
      case 'error':
        styledText = charsmRenderer.renderStyled(text, 'error');
        break;
      case 'system':
        styledText = charsmRenderer.renderStyled(text, 'info');
        break;
      case 'success':
        styledText = charsmRenderer.renderStyled(text, 'success');
        break;
      default:
        // Use default styling for regular output
        styledText = this._stripAnsiCodes(text);
    }

    return this._escapeHtml(styledText);
  }

  /**
   * Process text without Charsm (plain text)
   */
  _processPlain(text) {
    const cleanText = this._stripAnsiCodes(text);

    // Preserve table formatting
    if (this._isTableOutput(cleanText)) {
      return `<pre style="font-family: monospace; white-space: pre; margin: 0;">${this._escapeHtml(
        cleanText
      )}</pre>`;
    }

    return this._escapeHtml(cleanText);
  }

  /**
   * Check if text contains table output
   */
  _isTableOutput(text) {
    return (
      text.includes('┌') ||
      text.includes('│') ||
      text.includes('└') ||
      text.includes('─')
    );
  }

  /**
   * Check if text is JSON
   */
  _isJSONOutput(text) {
    const trimmed = text.trim();
    return (
      (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))
    );
  }

  /**
   * Render table output with Charsm
   */
  _renderTableOutput(text) {
    if (!charsmRenderer.initialized) {
      // Fallback to plain pre-formatted text
      return `<pre style="font-family: monospace; white-space: pre; margin: 10px 0; color: #33ff33;">${this._escapeHtml(
        text
      )}</pre>`;
    }

    try {
      // Parse the box-drawing table
      const parsed = this._parseBoxDrawingTable(text);
      if (parsed && parsed.headers && parsed.data) {
        console.log('✅ Parsed table data:', {
          headers: parsed.headers,
          rows: parsed.data.length,
        });
        // Use Charsm's table renderer
        const charsmTable = charsmRenderer.renderTable(
          parsed.data,
          parsed.headers
        );
        // Wrap in pre tag to preserve formatting and spacing
        return `<pre style="font-family: monospace; white-space: pre; margin: 10px 0;">${charsmTable}</pre>`;
      }
    } catch (error) {
      console.warn('Failed to parse table for Charsm rendering:', error);
    }

    // Fallback to plain text
    return `<pre style="font-family: monospace; white-space: pre; margin: 10px 0; color: #33ff33;">${this._escapeHtml(
      text
    )}</pre>`;
  }

  /**
   * Parse box-drawing table into structured data
   */
  _parseBoxDrawingTable(text) {
    const lines = text.split('\n').filter((l) => l.trim());
    if (lines.length < 4) return null;

    // Find header separator line (contains ├ or ┼)
    let headerIndex = -1;
    let dataStartIndex = -1;

    for (let i = 0; i < lines.length; i++) {
      if (lines[i].includes('├') || lines[i].includes('┼')) {
        headerIndex = i - 1; // Header is line before separator
        dataStartIndex = i + 1; // Data starts after separator
        break;
      }
    }

    if (headerIndex < 0 || dataStartIndex < 0) return null;

    // Parse header
    const headerLine = lines[headerIndex];
    const headers = headerLine
      .split('│')
      .map((h) => h.trim())
      .filter((h) => h);

    // Parse data rows
    const data = [];
    for (let i = dataStartIndex; i < lines.length; i++) {
      const line = lines[i];
      if (line.includes('└') || line.includes('┘')) break; // End of table

      const cells = line
        .split('│')
        .map((c) => c.trim())
        .filter((c) => c);

      if (cells.length === headers.length) {
        const row = {};
        headers.forEach((header, idx) => {
          row[header] = cells[idx];
        });
        data.push(row);
      }
    }

    return { headers, data };
  }

  /**
   * Strip ANSI escape codes
   */
  _stripAnsiCodes(text) {
    return text
      .replace(/\x1B\[[0-9;]*[mGKHfABCDEFnsuJST]/g, '')
      .replace(/\x1B\[[\?]?[0-9;]*[a-zA-Z]/g, '')
      .replace(/[\u001b]\[[0-9;]*[mGKHfABCDEFnsuJST]/g, '')
      .replace(/[\u001b]\[[\?]?[0-9;]*[a-zA-Z]/g, '');
  }

  /**
   * Escape HTML special characters
   */
  _escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  /**
   * Clear terminal
   */
  clear() {
    if (this.terminal) {
      this.terminal.innerHTML = '';
    }
  }

  /**
   * Toggle Charsm rendering
   */
  toggleCharsm() {
    this.useCharsm = !this.useCharsm;
    console.log(`Charsm rendering: ${this.useCharsm ? 'enabled' : 'disabled'}`);
    return this.useCharsm;
  }

  /**
   * Set Charsm rendering state
   */
  setCharsmEnabled(enabled) {
    this.useCharsm = enabled;
    console.log(`Charsm rendering: ${this.useCharsm ? 'enabled' : 'disabled'}`);
  }
}

// Create singleton instance
export const terminalOutputHandler = new TerminalOutputHandler();

// Export for global access
if (typeof window !== 'undefined') {
  window.terminalOutputHandler = terminalOutputHandler;

  // Create global appendToTerminal function for backward compatibility
  window.appendToTerminal = (text, className) => {
    terminalOutputHandler.appendToTerminal(text, className);
  };

  console.log('✅ Terminal output handler available globally');
}

export default terminalOutputHandler;
