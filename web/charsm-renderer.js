/**
 * Charsm Renderer Module
 *
 * This module provides styling and rendering capabilities using Charsm,
 * a WebAssembly port of lipgloss (Charm's Go styling library).
 *
 * Charsm allows us to apply CSS-like styling to CLI output in the browser,
 * including borders, padding, colors, and layout.
 */

import { initLip, Lipgloss } from './node_modules/charsm/dist/index.mjs';

/**
 * Initialize Charsm styles for the Megaport CLI
 */
export class CharsmRenderer {
  constructor() {
    this.initialized = false;
    this.lip = null;
    this.styleIds = {
      header: 'header',
      error: 'error',
      success: 'success',
      info: 'info',
      warning: 'warning',
      table: 'table',
      tableHeader: 'tableHeader',
      tableCell: 'tableCell',
      tableCellAlternate: 'tableCellAlternate',
      json: 'json',
      prompt: 'prompt',
    };
  }

  /**
   * Initialize the renderer and create predefined styles
   */
  async init() {
    if (this.initialized) return true;

    console.log('🎨 Initializing Charsm renderer...');

    try {
      // Initialize lipgloss WASM
      const isInit = await initLip();
      if (!isInit) {
        console.error('Failed to initialize Charsm WASM');
        return false;
      }

      // Create Lipgloss instance
      this.lip = new Lipgloss();

      // Create predefined styles using Charsm API
      // Header style
      this.lip.createStyle({
        id: this.styleIds.header,
        canvasColor: { color: '#33ff33', background: '#000' },
        bold: true,
        padding: [1, 2],
        margin: [1, 0],
      });

      // Error style
      this.lip.createStyle({
        id: this.styleIds.error,
        canvasColor: { color: '#ff3333' },
        bold: true,
      });

      // Success style
      this.lip.createStyle({
        id: this.styleIds.success,
        canvasColor: { color: '#33ff33' },
        bold: true,
      });

      // Info style
      this.lip.createStyle({
        id: this.styleIds.info,
        canvasColor: { color: '#3388ff' },
      });

      // Warning style
      this.lip.createStyle({
        id: this.styleIds.warning,
        canvasColor: { color: '#ffaa33' },
        bold: true,
      });

      // Table container style
      this.lip.createStyle({
        id: this.styleIds.table,
        border: { type: 'rounded', foreground: '#33ff33', sides: [true] },
        padding: [1, 2],
        margin: [1, 0, 1, 0],
      });

      // Table header style
      this.lip.createStyle({
        id: this.styleIds.tableHeader,
        canvasColor: { color: '#fff', background: '#c30048' }, // Megaport red
        bold: true,
        padding: [0, 1],
        alignV: 'center',
      });

      // Table cell style
      this.lip.createStyle({
        id: this.styleIds.tableCell,
        canvasColor: { color: '#33ff33' },
        padding: [0, 1],
      });

      // Alternate table cell style
      this.lip.createStyle({
        id: this.styleIds.tableCellAlternate,
        canvasColor: { color: '#00cc00' },
        padding: [0, 1],
      });

      // JSON style
      this.lip.createStyle({
        id: this.styleIds.json,
        canvasColor: { color: '#3388ff' },
        padding: [1],
      });

      // Prompt style
      this.lip.createStyle({
        id: this.styleIds.prompt,
        canvasColor: { color: '#33ff33' },
        bold: true,
      });

      this.initialized = true;
      console.log('✅ Charsm renderer initialized');
      return true;
    } catch (error) {
      console.error('❌ Failed to initialize Charsm:', error);
      return false;
    }
  }

  /**
   * Render text with a specific style
   */
  renderStyled(text, styleName) {
    if (!this.initialized) {
      console.warn('Charsm not initialized, returning plain text');
      return text;
    }

    const styleId = this.styleIds[styleName];
    if (!styleId) {
      console.warn(`Style "${styleName}" not found, returning plain text`);
      return text;
    }

    try {
      return this.lip.apply({ value: text, id: styleId });
    } catch (error) {
      console.error(`Error rendering with style "${styleName}":`, error);
      return text;
    }
  }

  /**
   * Render a table with Charsm styling using the newTable API
   *
   * @param {Array<Object>} data - Array of objects representing rows
   * @param {Array<string>} headers - Column headers
   */
  renderTable(data, headers) {
    if (!this.initialized) {
      console.warn('Charsm not initialized, returning plain table');
      return this._renderPlainTable(data, headers);
    }

    try {
      // Convert data objects to rows array
      const rows = data.map((row) =>
        headers.map((header) => String(row[header] || ''))
      );

      // Use Charsm's newTable API
      const table = this.lip.newTable({
        data: { headers: headers, rows: rows },
        table: { border: 'rounded', color: '99', width: 100 },
        header: { color: '212', bold: true },
        rows: { even: { color: '246' } },
      });

      return table;
    } catch (error) {
      console.error('Error rendering table with Charsm:', error);
      return this._renderPlainTable(data, headers);
    }
  }

  /**
   * Render JSON with syntax highlighting
   */
  renderJSON(jsonData) {
    if (!this.initialized) {
      return JSON.stringify(jsonData, null, 2);
    }

    try {
      const formatted = JSON.stringify(jsonData, null, 2);
      return this.lip.apply({ value: formatted, id: this.styleIds.json });
    } catch (error) {
      console.error('Error rendering JSON with Charsm:', error);
      return JSON.stringify(jsonData, null, 2);
    }
  }

  /**
   * Create a custom style and return the rendered text
   *
   * @param {Object} options - Style options
   * @param {string} options.id - Unique ID for the style
   * @param {Object} options.canvasColor - {color, background}
   * @param {Object} options.border - {type, foreground, background, sides}
   * @param {Array<number>} options.padding - Padding values
   * @param {Array<number>} options.margin - Margin values
   * @param {boolean} options.bold - Bold text
   * @param {string} options.alignV - Vertical alignment
   * @param {number} options.width - Width
   * @param {number} options.height - Height
   */
  createStyle(options) {
    if (!this.initialized) {
      console.warn('Charsm not initialized');
      return null;
    }

    try {
      // Generate a unique ID if not provided
      const styleId = options.id || `custom-${Date.now()}`;

      // Create style using Charsm API
      this.lip.createStyle({
        id: styleId,
        ...options,
      });

      // Return a function that can be used to apply the style
      return {
        id: styleId,
        render: (text) => this.lip.apply({ value: text, id: styleId }),
      };
    } catch (error) {
      console.error('Error creating custom style:', error);
      return null;
    }
  }

  /**
   * Join elements horizontally or vertically
   */
  join(direction, elements, position = 'left') {
    if (!this.initialized) {
      console.warn('Charsm not initialized, returning joined elements');
      return elements.join('\n');
    }

    try {
      return this.lip.join({
        direction: direction,
        elements: elements,
        position: position,
      });
    } catch (error) {
      console.error('Error joining elements:', error);
      return elements.join(direction === 'vertical' ? '\n' : ' ');
    }
  }

  /**
   * Render markdown with Charsm
   */
  renderMarkdown(content, theme = 'tokyo-night') {
    if (!this.initialized || !this.lip.RenderMD) {
      console.warn('Charsm not initialized or RenderMD not available');
      return content;
    }

    try {
      return this.lip.RenderMD(content, theme);
    } catch (error) {
      console.error('Error rendering markdown:', error);
      return content;
    }
  }

  /**
   * Calculate column widths for table
   */
  _calculateColumnWidths(data, headers) {
    const widths = headers.map((header) => header.length);

    data.forEach((row) => {
      headers.forEach((header, i) => {
        const value = String(row[header] || '');
        widths[i] = Math.max(widths[i], value.length);
      });
    });

    // Add padding
    return widths.map((w) => w + 4);
  }

  /**
   * Fallback plain table renderer
   */
  _renderPlainTable(data, headers) {
    const widths = this._calculateColumnWidths(data, headers);

    // Header
    const headerRow = headers
      .map((h, i) => h.toUpperCase().padEnd(widths[i]))
      .join(' | ');

    const separator = widths.map((w) => '─'.repeat(w)).join('─┼─');

    // Rows
    const rows = data
      .map((row) =>
        headers
          .map((h, i) => String(row[h] || '').padEnd(widths[i]))
          .join(' | ')
      )
      .join('\n');

    return `┌─${separator}─┐\n│ ${headerRow} │\n├─${separator}─┤\n│ ${rows
      .split('\n')
      .join(' │\n│ ')} │\n└─${separator}─┘`;
  }
}

// Create singleton instance
export const charsmRenderer = new CharsmRenderer();

// Initialize on module load
charsmRenderer
  .init()
  .then((success) => {
    if (success) {
      console.log(
        '✅ Charsm renderer available globally as window.charsmRenderer'
      );
    } else {
      console.warn(
        '⚠️  Charsm initialization failed, fallback to plain text rendering'
      );
    }
  })
  .catch((error) => {
    console.error('Failed to initialize Charsm renderer:', error);
  });

// Export for global access
if (typeof window !== 'undefined') {
  window.charsmRenderer = charsmRenderer;
}
