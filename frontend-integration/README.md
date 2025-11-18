# Megaport CLI WebAssembly - Vue 3 Integration Guide

## 🎯 Overview

This package provides Vue 3 + Vite integration for the Megaport CLI WebAssembly module. It's designed specifically for integration into the **Megaport Portal** (Vue 3 + Nuxt 3 + Vite stack).

## 📦 Package Contents

```
frontend-integration/
├── types/
│   └── megaport-wasm.d.ts          # TypeScript definitions
├── composables/
│   └── useMegaportWASM.ts          # Vue composable for WASM
├── components/
│   └── MegaportTerminal.vue        # Terminal component with xterm.js
├── workers/
│   └── megaport-worker.ts          # Web Worker (optional, for heavy workloads)
├── demo/
│   ├── App.vue                     # Demo application
│   └── main.ts                     # Demo entry point
├── package.json
├── vite.config.ts
├── tsconfig.json
└── README.md (this file)
```

## 🚀 Quick Start

### 1. Installation

```bash
npm install xterm xterm-addon-fit xterm-addon-web-links
```

### 2. Copy WASM Files

Copy these files to your `public/` directory:

```bash
# From the CLI build output
cp dist/megaport.wasm public/
cp dist/wasm_exec.js public/
```

### 3. Basic Usage in Vue 3

```vue
<template>
  <div>
    <p v-if="isLoading">Loading Megaport CLI...</p>
    <p v-else-if="error">Error: {{ error.message }}</p>
    <div v-else>
      <button @click="listPorts">List Ports</button>
      <pre>{{ output }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useMegaportWASM } from './composables/useMegaportWASM';

const { isLoading, isReady, error, execute, setAuth } = useMegaportWASM();
const output = ref('');

// Set credentials (typically from your auth system)
setAuth('your-access-key', 'your-secret-key', 'staging');

const listPorts = async () => {
  const result = await execute('port list --output json');
  output.value = result.output || result.error;
};
</script>
```

### 4. Using the Terminal Component

```vue
<template>
  <MegaportTerminal
    wasm-path="/megaport.wasm"
    wasm-exec-path="/wasm_exec.js"
    :theme="{
      background: '#1e1e1e',
      foreground: '#d4d4d4',
    }"
  />
</template>

<script setup>
import MegaportTerminal from './components/MegaportTerminal.vue';
</script>
```

## 🏗️ Architecture

### Direct Mode (Recommended for Portal)

```
┌─────────────────────┐
│   Vue 3 Component   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  useMegaportWASM()  │  ← Composable
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   wasm_exec.js      │  ← Go WASM runtime
│   megaport.wasm     │
└─────────────────────┘
```

### Worker Mode (Optional, for better performance)

```
┌─────────────────────┐
│   Vue 3 Component   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  useMegaportWASM()  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Web Worker        │
│  (dedicated thread) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   WASM Module       │
└─────────────────────┘
```

## 📚 API Reference

### `useMegaportWASM(config?)`

Vue composable for WASM integration.

**Parameters:**

- `config.wasmPath` (string): Path to megaport.wasm (default: '/megaport.wasm')
- `config.wasmExecPath` (string): Path to wasm_exec.js (default: '/wasm_exec.js')
- `config.debug` (boolean): Enable debug logging (default: false)
- `config.useWorker` (boolean): Use Web Worker (default: false)

**Returns:**

```typescript
{
  isLoading: Ref<boolean>,       // WASM is loading
  isReady: Ref<boolean>,         // WASM is ready
  error: Ref<Error | null>,      // Initialization error
  execute: (cmd: string) => Promise<Result>,  // Execute command
  setAuth: (key, secret, env) => void,        // Set credentials
  clearAuth: () => void,                       // Clear credentials
  getAuthInfo: () => AuthInfo,                // Get auth status
  resetOutput: () => void,                    // Reset output buffers
  toggleDebug: () => boolean                  // Toggle debug mode
}
```

### Available Commands

All standard Megaport CLI commands are supported:

```bash
# Resource Management
port list [--output json|table|csv]
vxc list [--output json|table|csv]
mcr list [--output json|table|csv]
mve list [--output json|table|csv]

# Information
location list
partner list
servicekey list

# Terminal Commands
help        # Show help
clear       # Clear terminal
```

## 🔐 Authentication

### Browser-Based Auth (Recommended)

Since WASM runs in the browser, use **localStorage** for credentials:

```typescript
const { setAuth } = useMegaportWASM();

// After user logs in via your auth system
setAuth(accessKey, secretKey, 'staging');
```

### Security Best Practices

**Important**: Both the Access Key and Secret Key should be treated as sensitive credentials:

- Use `type="password"` for both Access Key and Secret Key input fields
- Never expose credentials in client-side code or logs
- Clear credentials when users log out using `clearAuth()`
- Consider implementing session timeouts
- Use HTTPS in production to protect credentials in transit

### Environment Variables

WASM reads from localStorage keys:

- `MEGAPORT_ACCESS_KEY`
- `MEGAPORT_SECRET_KEY`
- `MEGAPORT_ENVIRONMENT`

These are automatically set by `setAuth()`.

## 🎨 Vite Configuration

### For Nuxt 3

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  vite: {
    optimizeDeps: {
      exclude: ['xterm', 'xterm-addon-fit', 'xterm-addon-web-links'],
    },
    server: {
      fs: {
        allow: ['..'], // If WASM files are outside public/
      },
    },
  },
});
```

### For Vite

```typescript
// vite.config.ts
export default defineConfig({
  optimizeDeps: {
    exclude: ['xterm', 'xterm-addon-fit', 'xterm-addon-web-links'],
  },
  server: {
    headers: {
      'Cross-Origin-Embedder-Policy': 'require-corp',
      'Cross-Origin-Opener-Policy': 'same-origin',
    },
  },
});
```

## 🧪 Testing the Integration

### 1. Run Demo Application

```bash
cd frontend-integration
npm install
npm run dev
```

### 2. Test Commands

Try these commands in the terminal:

```bash
location list
help
port list --output json
```

### 3. Verify Output

- Check browser console for debug logs
- Verify WASM initialization messages
- Test authentication flow
- Confirm API responses

## ⚡ Performance Considerations

### WASM File Size

- `megaport.wasm`: ~2-5 MB
- `wasm_exec.js`: ~15 KB
- First load: 2-5 seconds (includes compilation)
- Subsequent calls: Near-native speed

### Optimization Tips

1. **Lazy Load**: Load WASM only when needed

   ```typescript
   const showTerminal = ref(false);
   // WASM loads when showTerminal becomes true
   ```

2. **Cache WASM**: Vite/Nuxt will cache WASM files

   ```typescript
   // Service Worker caching
   workbox.precaching.precacheAndRoute([
     { url: '/megaport.wasm', revision: '1.0.0' },
   ]);
   ```

3. **Use Worker**: For heavy workloads
   ```typescript
   useMegaportWASM({ useWorker: true });
   ```

## 🐛 Troubleshooting

### WASM Fails to Load

```javascript
// Check browser console
console.log(window.Go); // Should be defined
console.log(window.executeMegaportCommandAsync); // Should be function
```

**Solutions:**

- Verify WASM files are in `public/`
- Check MIME types: `application/wasm`
- Ensure CORS headers are correct
- Clear browser cache

### Commands Return No Output

```javascript
// Check WASM debug info
window.toggleWasmDebug(); // Enable debug
window.dumpBuffers(); // Check buffer contents
```

**Solutions:**

- Verify authentication is set
- Check command syntax
- Enable debug mode
- Review browser console logs

### TypeScript Errors

Ensure types are properly configured:

```json
// tsconfig.json
{
  "compilerOptions": {
    "types": ["vite/client"],
    "moduleResolution": "bundler"
  },
  "include": ["types/megaport-wasm.d.ts"]
}
```

## 🌐 Browser Compatibility

| Browser | Minimum Version | Status          |
| ------- | --------------- | --------------- |
| Chrome  | 57+             | ✅ Full Support |
| Firefox | 52+             | ✅ Full Support |
| Safari  | 11+             | ✅ Full Support |
| Edge    | 16+             | ✅ Full Support |

## 📝 Example Integration Scenarios

### Scenario 1: Portal Dashboard

```vue
<template>
  <DashboardCard title="Quick Actions">
    <button @click="createVXC">Create VXC</button>
    <button @click="listResources">View Resources</button>
  </DashboardCard>
</template>

<script setup>
import { useMegaportWASM } from '@/composables/useMegaportWASM';

const { execute } = useMegaportWASM();

const createVXC = async () => {
  // Show wizard dialog, collect params
  const result = await execute('vxc create ...');
  // Update dashboard
};
</script>
```

### Scenario 2: Admin Console

```vue
<template>
  <AdminPanel>
    <MegaportTerminal ref="terminal" @command-executed="logActivity" />
  </AdminPanel>
</template>

<script setup>
const logActivity = (command, result) => {
  // Send to analytics
  analytics.track('cli_command', { command, result });
};
</script>
```

## 🔗 Integration with Existing Portal Features

### Authentication

Use your existing auth system:

```typescript
// After user logs in
onUserLogin((credentials) => {
  const { setAuth } = useMegaportWASM();
  setAuth(
    credentials.accessKey,
    credentials.secretKey,
    credentials.environment
  );
});
```

### State Management (Pinia)

```typescript
// stores/megaport.ts
import { defineStore } from 'pinia';
import { useMegaportWASM } from '@/composables/useMegaportWASM';

export const useMegaportStore = defineStore('megaport', () => {
  const wasm = useMegaportWASM();

  const listPorts = async () => {
    const result = await wasm.execute('port list --output json');
    return JSON.parse(result.output);
  };

  return { listPorts };
});
```

### Router Integration

```typescript
// Lazy load for specific routes
{
  path: '/admin/cli',
  component: () => import('@/views/CLITerminal.vue'),
  meta: { requiresWASM: true }
}
```
