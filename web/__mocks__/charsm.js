/**
 * Mock implementation of Charsm for testing
 *
 * This mock provides a simple implementation that mimics the Charsm API
 * without requiring the actual WASM module.
 */

import { jest } from '@jest/globals';

// Mock Lipgloss class
export class Lipgloss {
  constructor() {
    this.styles = new Map();
  }

  createStyle(options) {
    if (options.id) {
      this.styles.set(options.id, options);
    }
    return options.id;
  }

  apply({ value, id }) {
    const style = this.styles.get(id);
    if (!style) return value;

    // Add simple ANSI-like markers to indicate styling was applied
    let styled = value;
    if (style.bold) styled = `\x1b[1m${styled}\x1b[0m`;
    if (style.canvasColor?.color) styled = `\x1b[32m${styled}\x1b[0m`; // Green for testing

    return styled;
  }

  newTable({ data, table, header, rows }) {
    if (!data || !data.headers || !data.rows) {
      return '';
    }

    const { headers, rows: tableRows } = data;

    // Simple table rendering for tests
    const colWidths = headers.map((h, i) => {
      const maxContentWidth = Math.max(
        h.length,
        ...tableRows.map((row) => String(row[i] || '').length)
      );
      return maxContentWidth + 2;
    });

    // Build table
    const lines = [];

    // Top border
    lines.push('┌' + colWidths.map((w) => '─'.repeat(w)).join('┬') + '┐');

    // Headers
    const headerRow = headers.map((h, i) => h.padEnd(colWidths[i])).join('│');
    lines.push('│' + headerRow + '│');

    // Header separator
    lines.push('├' + colWidths.map((w) => '─'.repeat(w)).join('┼') + '┤');

    // Rows
    tableRows.forEach((row) => {
      const rowStr = row
        .map((cell, i) => String(cell).padEnd(colWidths[i]))
        .join('│');
      lines.push('│' + rowStr + '│');
    });

    // Bottom border
    lines.push('└' + colWidths.map((w) => '─'.repeat(w)).join('┴') + '┘');

    return lines.join('\n');
  }

  join({ direction, elements, position }) {
    if (direction === 'vertical') {
      return elements.join('\n');
    } else {
      return elements.join(' ');
    }
  }

  RenderMD(content, theme) {
    // Simple markdown rendering mock
    return content
      .replace(/^# (.+)$/gm, '\x1b[1m$1\x1b[0m')
      .replace(/\*\*(.+?)\*\*/g, '\x1b[1m$1\x1b[0m')
      .replace(/\*(.+?)\*/g, '\x1b[3m$1\x1b[0m');
  }
}

// Mock initLip function
export const initLip = jest.fn(async () => {
  // Simulate async WASM initialization
  await new Promise((resolve) => setTimeout(resolve, 10));
  return true;
});
