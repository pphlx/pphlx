
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import solidPlugin from 'vite-plugin-solid';

export default defineConfig({
  plugins: [vue(), svelte(), solidPlugin()],
  build: {
    lib: {
      entry: 'src/.pphlx_entry.js',
      formats: ['iife'],
      name: 'PphlxViteComponents',
      fileName: () => 'pphlx_vite.js'
    },
    rollupOptions: {
      external: ['vue'],
      output: {
        globals: {
          vue: 'Vue'
        }
      }
    },
    outDir: 'dist/assets/js',
    emptyOutDir: false,
    minify: true
  }
});
