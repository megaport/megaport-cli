import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';

export default defineConfig({
  plugins: [vue()],

  build: {
    lib: {
      entry: resolve(__dirname, 'index.ts'),
      name: 'MegaportCLIWASM',
      fileName: (format: string) => `megaport-cli-wasm.${format}.js`,
    },
    rollupOptions: {
      external: [
        'vue',
        '@xterm/xterm',
        '@xterm/addon-fit',
        '@xterm/addon-web-links',
      ],
      output: {
        globals: {
          vue: 'Vue',
          '@xterm/xterm': 'Terminal',
          '@xterm/addon-fit': 'FitAddon',
          '@xterm/addon-web-links': 'WebLinksAddon',
        },
      },
    },
  },

  // For development/demo
  server: {
    port: 3000,
    open: true,
  },

  // Optimization for WASM
  optimizeDeps: {
    exclude: ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
  },
});
