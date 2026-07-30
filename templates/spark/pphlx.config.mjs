import { defineConfig } from 'vite';

// Single Source of Truth PPHLX Configuration (pphlx.config.mjs)
export default defineConfig({
  srcDir: 'src',
  outDir: 'dist',
  output: {
    target: 'php'
  }
});
