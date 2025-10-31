/**
 * Jest Setup File
 *
 * This file runs before each test suite to set up the testing environment.
 */

import { jest } from '@jest/globals';

// Mock console methods to reduce noise in test output (optional)
global.console = {
  ...console,
  log: jest.fn(), // Comment this out if you want to see console.log in tests
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  // Keep error and assert for debugging
  error: console.error,
  assert: console.assert,
};

// Set up localStorage mock for JSDOM
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Add performance.now() if not available
if (typeof performance === 'undefined') {
  global.performance = {
    now: () => Date.now(),
  };
}

// Reset mocks before each test
beforeEach(() => {
  localStorageMock.getItem.mockClear();
  localStorageMock.setItem.mockClear();
  localStorageMock.removeItem.mockClear();
  localStorageMock.clear.mockClear();
});
