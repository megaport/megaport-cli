import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';

export default defineConfig({
  plugins: [vue()],

  root: './',

  build: {
    outDir: '../web/vue-demo',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
      },
    },
    // Ensure WASM files are copied correctly
    assetsDir: 'assets',
    copyPublicDir: false,
  },

  // For development
  server: {
    port: 3000,
    open: true,
  },

  // Public directory for WASM files
  publicDir: false,

  // Optimization for WASM and xterm
  optimizeDeps: {
    exclude: ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
  },

  resolve: {
    alias: {
      '@': resolve(__dirname, './'),
    },
  },
});
