#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');
const net = require('net');

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

const args = process.argv.slice(2);
const isDev = args.includes('dev') || args.includes('watch');
const isPreview = args.includes('preview');

const timePrefix = () => {
  const d = new Date();
  return `\x1b[90m${d.toTimeString().split(' ')[0]}\x1b[0m`;
};

function getFreePort(startPort, callback) {
  const server = net.createServer();
  server.listen(startPort, 'localhost', () => {
    server.close(() => {
      callback(null, startPort);
    });
  });
  server.on('error', (err) => {
    if (err.code === 'EADDRINUSE') {
      console.log(`${timePrefix()} \x1b[33m[pphlx] Port ${startPort} is in use, trying another one...\x1b[0m`);
      getFreePort(startPort + 1, callback);
    } else {
      callback(err);
    }
  });
}

function runCompiler() {
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
}

if (isPreview || isDev) {
  getFreePort(6321, (err, port) => {
    if (err) {
      console.error('Failed to find open port:', err);
      process.exit(1);
    }

    const pkg = require('../package.json');
    const version = pkg.version || '1.0.5';

    if (isPreview) {
      console.log(`\n  \x1b[30m\x1b[42m pphlx \x1b[0m \x1b[32mv${version} preview mode ready\x1b[0m\n`);
      console.log(`  \x1b[90m┃\x1b[0m \x1b[1mLocal\x1b[0m    \x1b[36mhttp://localhost:${port}/\x1b[0m`);
      console.log(`  \x1b[90m┃\x1b[0m \x1b[90mServing directory: ${outDir}\x1b[0m\n`);
      
      const phpServer = spawn('php', ['-S', `localhost:${port}`, '-t', outDir], {
        stdio: 'inherit'
      });
      phpServer.on('close', (code) => {
        process.exit(code);
      });
    } else if (isDev) {
      const phpServer = spawn('php', ['-S', `localhost:${port}`, '-t', outDir], {
        stdio: 'ignore',
        detached: true
      });
      phpServer.unref();

      console.log(`\n  \x1b[30m\x1b[42m pphlx \x1b[0m \x1b[32mv${version} ready\x1b[0m\n`);
      console.log(`  \x1b[90m┃\x1b[0m \x1b[1mLocal\x1b[0m    \x1b[36mhttp://localhost:${port}/\x1b[0m`);
      console.log(`  \x1b[90m┃\x1b[0m \x1b[90mNetwork  use --host to expose\x1b[0m\n`);

      runCompiler();
    }
  });
} else {
  runCompiler();
}
