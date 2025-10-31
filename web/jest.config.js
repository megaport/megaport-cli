export default {
  testEnvironment: 'jsdom',
  testMatch: ['**/*.test.js'],
  moduleFileExtensions: ['js', 'mjs'],
  transform: {},
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
    '^.*node_modules/charsm/dist/index\\.mjs$': '<rootDir>/__mocks__/charsm.js',
  },
  collectCoverageFrom: [
    '*.js',
    '!*.test.js',
    '!wasm_exec.js',
    '!node_modules/**',
  ],
  coveragePathIgnorePatterns: ['/node_modules/', 'wasm_exec.js'],
  testTimeout: 30000, // 30 seconds for WASM initialization
  globals: {
    'ts-jest': {
      useESM: true,
    },
  },
  setupFilesAfterEnv: ['./jest.setup.js'],
};
