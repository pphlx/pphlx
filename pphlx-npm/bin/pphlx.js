#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

// Determine output directory from config
let outDir = 'dist';
try {
  const configJsonPath = path.join(process.cwd(), 'pphlx.config.json');
  const manifestJsonPath = path.join(process.cwd(), 'pphlx.json');
  if (fs.existsSync(configJsonPath)) {
    const config = JSON.parse(fs.readFileSync(configJsonPath, 'utf8'));
    if (config.outDir) outDir = config.outDir;
  } else if (fs.existsSync(manifestJsonPath)) {
    const config = JSON.parse(fs.readFileSync(manifestJsonPath, 'utf8'));
    if (config.outDir) outDir = config.outDir;
  } else {
    const configMjsPath = path.join(process.cwd(), 'pphlx.config.mjs');
    if (fs.existsSync(configMjsPath)) {
      const content = fs.readFileSync(configMjsPath, 'utf8');
      const match = content.match(/outDir\s*:\s*["']([^"']+)["']/);
      if (match) outDir = match[1];
    }
  }
} catch (e) {}

// Launch background PHP server on dev or watch commands
const args = process.argv.slice(2);
const isDev = args.includes('dev') || args.includes('watch');
const isPreview = args.includes('preview');

if (isPreview) {
  console.log(`\n  pphlx  v${require('../package.json').version || '1.0.4'} preview mode ready\n`);
  console.log(`  ┃ Local    http://localhost:8000/`);
  console.log(`  ┃ Serving directory: ${outDir}\n`);
  const phpServer = spawn('php', ['-S', 'localhost:8000', '-t', outDir], {
    stdio: 'inherit'
  });
  phpServer.on('close', (code) => {
    process.exit(code);
  });
  return;
}

if (isDev) {
  const phpServer = spawn('php', ['-S', 'localhost:8000', '-t', outDir], {
    stdio: 'ignore',
    detached: true
  });
  phpServer.unref();
  
  const pkg = require('../package.json');
  const version = pkg.version || '1.0.3';
  console.log(`\n  pphlx  v${version} ready\n`);
  console.log(`  ┃ Local    http://localhost:8000/`);
  console.log(`  ┃ Network  use --host to expose\n`);
}

// On Windows, run the native Go compiler binary directly to bypass Node.js WASI filesystem bugs
if (process.platform === 'win32') {
  const binaryPath = path.join(__dirname, 'pphlx-win.exe');
  if (fs.existsSync(binaryPath)) {
    const proc = spawn(binaryPath, args, { stdio: 'inherit' });
    proc.on('close', (code) => {
      process.exit(code);
    });
    return;
  }
}

// Load the Go WebAssembly execution helper
require('./wasm_exec.js');

const go = new Go();
go.argv = process.argv;
go.env = process.env;

const wasmPath = path.join(__dirname, '../lib/pphlx.wasm');
const wasmBuffer = fs.readFileSync(wasmPath);

WebAssembly.instantiate(wasmBuffer, go.importObject).then((result) => {
  go.run(result.instance);
}).catch((err) => {
  console.error('Failed to run PPHLX compiler:', err);
  process.exit(1);
});
