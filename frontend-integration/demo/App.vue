/** * Demo Application: Megaport CLI WASM with Vue 3 */

<template>
  <div id="app" class="app-container">
    <header class="app-header">
      <h1>🚀 Megaport CLI WebAssembly Demo</h1>
      <p>Vue 3 + Vite + WASM Integration</p>
    </header>

    <main class="app-main">
      <!-- Auth Panel -->
      <section class="auth-panel" v-if="!authConfigured">
        <h2>Authentication</h2>
        <p>Configure your Megaport API credentials to get started.</p>

        <form @submit.prevent="handleAuth">
          <div class="form-group">
            <label for="accessKey">Access Key:</label>
            <input
              id="accessKey"
              v-model="authForm.accessKey"
              type="password"
              placeholder="Enter access key"
              autocomplete="off"
              data-1p-ignore
              required
            />
          </div>

          <div class="form-group">
            <label for="secretKey">Secret Key:</label>
            <input
              id="secretKey"
              v-model="authForm.secretKey"
              type="password"
              placeholder="Enter secret key"
              autocomplete="off"
              data-1p-ignore
              required
            />
          </div>

          <div class="form-group">
            <label for="environment">Environment:</label>
            <select id="environment" v-model="authForm.environment">
              <option value="staging">Staging</option>
              <option value="production">Production</option>
            </select>
          </div>

          <button type="submit" class="btn-primary">Set Credentials</button>
        </form>
      </section>

      <!-- Terminal -->
      <section v-else class="terminal-section">
        <div class="terminal-header">
          <h2>Interactive Terminal</h2>
          <button @click="handleClearAuth" class="btn-secondary">
            🔓 Clear Auth
          </button>
        </div>

        <MegaportTerminal
          ref="terminalRef"
          wasm-path="/megaport.wasm"
          wasm-exec-path="/wasm_exec.js"
          :welcome-message="welcomeMessage"
        />
      </section>

      <!-- Quick Actions -->
      <aside class="quick-actions" v-if="authConfigured">
        <h3>Quick Actions</h3>
        <div class="action-buttons">
          <button
            @click="runCommand('ports list --output json')"
            class="btn-action"
          >
            📋 List Ports
          </button>
          <button
            @click="runCommand('mcr list --output json')"
            class="btn-action"
          >
            🌐 List MCRs
          </button>
          <button
            @click="runCommand('mve list --output json')"
            class="btn-action"
          >
            🖥️ List MVEs
          </button>
          <button
            @click="runCommand('locations list --output json')"
            class="btn-action"
          >
            📍 List Locations
          </button>
          <button @click="runCommand('help')" class="btn-action">
            ❓ Help
          </button>
          <button @click="runCommand('clear')" class="btn-action">
            🧹 Clear
          </button>
        </div>
      </aside>

      <!-- Status Info -->
      <aside class="status-info">
        <h3>WASM Status</h3>
        <div class="status-item">
          <span class="label">Loading:</span>
          <span :class="['value', isLoading ? 'loading' : 'ready']">
            {{ isLoading ? '⏳ Yes' : '✅ No' }}
          </span>
        </div>
        <div class="status-item">
          <span class="label">Ready:</span>
          <span :class="['value', isReady ? 'ready' : 'not-ready']">
            {{ isReady ? '✅ Yes' : '❌ No' }}
          </span>
        </div>
        <div class="status-item" v-if="error">
          <span class="label">Error:</span>
          <span class="value error">{{ error.message }}</span>
        </div>
        <div class="status-item" v-if="authInfo">
          <span class="label">Environment:</span>
          <span class="value">{{ authInfo.environment }}</span>
        </div>
      </aside>
    </main>

    <footer class="app-footer">
      <p>
        Megaport CLI WASM Integration Demo |
        <a href="https://github.com/megaport" target="_blank">GitHub</a> |
        <a href="https://docs.megaport.com" target="_blank">Documentation</a>
      </p>
    </footer>
  </div>
</template>

<script setup lang="ts">
/// <reference path="../vite-env.d.ts" />
import { ref, computed, onMounted } from 'vue';
// @ts-ignore - Vue SFC default export is handled by the Vue compiler
import MegaportTerminal from '../components/MegaportTerminal.vue';
import { useMegaportWASM } from '../composables/useMegaportWASM';

// WASM integration
const { isLoading, isReady, error, setAuth, clearAuth, getAuthInfo } =
  useMegaportWASM({
    debug: true,
    useWorker: false,
  });

// Auth state
const authForm = ref({
  accessKey: '',
  secretKey: '',
  environment: 'staging',
});

const authConfigured = ref(false);
const authInfo = ref<any>(null);

// Terminal reference
const terminalRef = ref<InstanceType<typeof MegaportTerminal> | null>(null);

// Welcome message
const welcomeMessage = computed(() => {
  return `Welcome to Megaport CLI (WebAssembly)
Environment: ${authForm.value.environment}
Type "help" for available commands.

`;
});

/**
 * Handle auth form submission
 */
const handleAuth = () => {
  setAuth(
    authForm.value.accessKey,
    authForm.value.secretKey,
    authForm.value.environment
  );

  authConfigured.value = true;
  authInfo.value = getAuthInfo();

  // Also set auth in terminal
  if (terminalRef.value?.setAuth) {
    terminalRef.value.setAuth(
      authForm.value.accessKey,
      authForm.value.secretKey,
      authForm.value.environment
    );
  }
};

/**
 * Clear authentication
 */
const handleClearAuth = () => {
  clearAuth();
  authConfigured.value = false;
  authInfo.value = null;
  authForm.value = {
    accessKey: '',
    secretKey: '',
    environment: 'staging',
  };
};

/**
 * Run a command in the terminal
 */
const runCommand = (command: string) => {
  if (terminalRef.value?.execute) {
    terminalRef.value.execute(command);
  }
};

// Check for existing auth on mount
onMounted(() => {
  authInfo.value = getAuthInfo();
  if (authInfo.value?.accessKeySet && authInfo.value?.secretKeySet) {
    authConfigured.value = true;
  }
});
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, sans-serif;
}

.app-header {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 2rem;
  text-align: center;
  color: white;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
}

.app-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2rem;
  font-weight: 700;
}

.app-header p {
  margin: 0;
  opacity: 0.9;
  font-size: 1.1rem;
}

.app-main {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 300px;
  grid-template-rows: auto 1fr;
  gap: 1.5rem;
  padding: 1.5rem;
  max-width: 1800px;
  margin: 0 auto;
  width: 100%;
}

.auth-panel {
  grid-column: 1 / -1;
  background: white;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  max-width: 500px;
  margin: 2rem auto;
}

.auth-panel h2 {
  margin: 0 0 0.5rem 0;
  color: #333;
}

.auth-panel p {
  margin: 0 0 1.5rem 0;
  color: #666;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

.terminal-section {
  grid-column: 1;
  grid-row: 1 / -1;
  background: white;
  border-radius: 8px;
  padding: 1rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  min-height: 600px;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #e0e0e0;
}

.terminal-header h2 {
  margin: 0;
  color: #333;
  font-size: 1.3rem;
}

.quick-actions,
.status-info {
  background: white;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.quick-actions {
  grid-column: 2;
  grid-row: 1;
}

.status-info {
  grid-column: 2;
  grid-row: 2;
}

.quick-actions h3,
.status-info h3 {
  margin: 0 0 1rem 0;
  color: #333;
  font-size: 1.1rem;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.btn-primary,
.btn-secondary,
.btn-action {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
  width: 100%;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(102, 126, 234, 0.3);
}

.btn-secondary {
  background: #f0f0f0;
  color: #333;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn-action {
  background: #f8f9fa;
  color: #333;
  text-align: left;
  border: 1px solid #e0e0e0;
}

.btn-action:hover {
  background: #667eea;
  color: white;
  border-color: #667eea;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  border-bottom: 1px solid #f0f0f0;
}

.status-item:last-child {
  border-bottom: none;
}

.status-item .label {
  font-weight: 600;
  color: #666;
}

.status-item .value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
}

.status-item .value.loading {
  color: #f39c12;
}

.status-item .value.ready {
  color: #27ae60;
}

.status-item .value.not-ready {
  color: #e74c3c;
}

.status-item .value.error {
  color: #e74c3c;
  font-size: 0.85rem;
}

.app-footer {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  padding: 1.5rem;
  text-align: center;
  color: white;
  border-top: 1px solid rgba(255, 255, 255, 0.2);
}

.app-footer p {
  margin: 0;
  font-size: 0.9rem;
}

.app-footer a {
  color: white;
  text-decoration: none;
  margin: 0 0.5rem;
  font-weight: 600;
}

.app-footer a:hover {
  text-decoration: underline;
}

@media (max-width: 1200px) {
  .app-main {
    grid-template-columns: 1fr;
    grid-template-rows: auto;
  }

  .terminal-section {
    grid-column: 1;
    grid-row: auto;
  }

  .quick-actions,
  .status-info {
    grid-column: 1;
    grid-row: auto;
  }
}
</style>
