#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { WASI } = require('wasi');

// Initialize WASI with host environment and working directory mapped to guest root
const wasi = new WASI({
  version: 'preview1',
  args: process.argv,
  env: process.env,
  preopens: {
    '/': process.cwd()
  }
});

const wasmPath = path.join(__dirname, '../lib/pphlx.wasm');
const wasmBuffer = fs.readFileSync(wasmPath);

WebAssembly.instantiate(wasmBuffer, {
  wasi_snapshot_preview1: wasi.wasiImport
}).then((result) => {
  wasi.start(result.instance);
}).catch((err) => {
  console.error('Failed to run PPHLX compiler:', err);
  process.exit(1);
});
