import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    outDir: resolve(__dirname, 'web/static/js'),
    emptyOutDir: true,
    rollupOptions: {
      input: {
        'theme-init': resolve(__dirname, 'web/js/theme-init.js'),
        main: resolve(__dirname, 'web/js/main.js'),
      },
      output: {
        format: 'iife',
        entryFileNames: '[name].js',
        chunkFileNames: '[name].js',
      },
    },
  },
});
