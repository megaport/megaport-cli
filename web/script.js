// Direct browser fetch implementation
(function initGlobalFunctions() {
  window.directBrowserFetch = function (
    url,
    token,
    onSuccess,
    onError,
    options
  ) {
    const method = options?.method || 'GET';
    const body = options?.body || null;
    const headers = options?.headers || {};

    console.log(`Starting ${method} request to ${url}`);

    const fetchOptions = {
      method: method,
      headers: {},
    };

    // Set headers from options
    if (headers && typeof headers === 'object') {
      Object.entries(headers).forEach(([key, value]) => {
        fetchOptions.headers[key] = value;
      });
    }

    // Set token if provided
    if (token) {
      fetchOptions.headers['Authorization'] = `Bearer ${token}`;
    }

    // Set default Content-Type if not already set
    if (
      !fetchOptions.headers['Content-Type'] &&
      !fetchOptions.headers['content-type']
    ) {
      fetchOptions.headers['Content-Type'] = 'application/json';
    }

    // Add body if provided
    if (body) {
      fetchOptions.body = body;
    }

    // Use native fetch
    fetch(url, fetchOptions)
      .then((response) => {
        console.log(`Response received with status ${response.status}`);

        if (response.ok) {
          return response.text();
        } else {
          throw new Error(
            `HTTP Error: ${response.status} ${response.statusText}`
          );
        }
      })
      .then((text) => {
        console.log(`Success, data length: ${text.length} bytes`);
        try {
          onSuccess(text);
        } catch (e) {
          console.error('Error in success callback:', e);
          onError(`Callback error: ${e.message}`);
        }
      })
      .catch((error) => {
        console.error('Fetch error:', error);
        try {
          onError(error.message || 'Network error occurred');
        } catch (e) {
          console.error('Error in error callback:', e);
        }
      });

    console.log(`${method} request initiated`);
  };

  console.log('✅ Direct fetch initialized');
})();

// Terminal helpers
// Strip ANSI escape codes from text
function stripAnsiCodes(text) {
  // Remove all ANSI escape sequences
  // This handles: ESC[...m (colors), ESC[...K (clear), ESC[...H (cursor), etc.
  // Using both \x1B and \u001b to catch all escape sequences
  return text
    .replace(/\x1B\[[0-9;]*[mGKHfABCDEFnsuJST]/g, '') // Standard ANSI codes
    .replace(/\x1B\[[\?]?[0-9;]*[a-zA-Z]/g, '') // Extended ANSI codes
    .replace(/[\u001b]\[[0-9;]*[mGKHfABCDEFnsuJST]/g, '') // Unicode escape
    .replace(/[\u001b]\[[\?]?[0-9;]*[a-zA-Z]/g, ''); // Unicode extended
}

// Convert ANSI codes to HTML (basic implementation)
function ansiToHtml(text) {
  // For now, just strip them - we can enhance this later with color support
  return stripAnsiCodes(text);
}

function appendToTerminal(text, className) {
  const terminal = document.getElementById('terminal');
  const line = document.createElement('div');
  if (className) line.className = className;

  // For table output or any output with ANSI codes, clean it up
  const cleanText = ansiToHtml(text);

  // Use <pre> for table output to preserve formatting
  if (
    cleanText.includes('┌') ||
    cleanText.includes('│') ||
    cleanText.includes('└')
  ) {
    const pre = document.createElement('pre');
    pre.style.fontFamily = 'monospace';
    pre.style.whiteSpace = 'pre';
    pre.style.margin = '0';
    pre.textContent = cleanText;
    line.appendChild(pre);
  } else {
    line.textContent = cleanText;
  }

  terminal.appendChild(line);
  terminal.scrollTop = terminal.scrollHeight;
}

// Token management
const tokenManager = {
  storeToken(environment, token, expiresIn) {
    const data = {
      accessToken: token,
      environment,
      expiresAt: Date.now() + expiresIn * 1000,
      createdAt: Date.now(),
    };
    localStorage.setItem('megaport_auth_token', JSON.stringify(data));
    console.log(`Token stored for ${environment}, expires in ${expiresIn}s`);
  },

  getToken(environment) {
    try {
      const str = localStorage.getItem('megaport_auth_token');
      if (!str) return null;
      const data = JSON.parse(str);
      if (data.environment !== environment) return null;
      // Check if token expires in the next 5 minutes
      if (Date.now() > data.expiresAt - 5 * 60 * 1000) {
        console.log('Cached token is expiring soon or has expired.');
        localStorage.removeItem('megaport_auth_token'); // Remove stale token
        return null;
      }
      const mins = Math.floor((data.expiresAt - Date.now()) / 60000);
      console.log(`Using cached token (expires in ${mins}m)`);
      return data.accessToken;
    } catch (e) {
      console.error('Error reading token:', e);
      localStorage.removeItem('megaport_auth_token'); // Clear corrupted token
      return null;
    }
  },

  clearToken() {
    localStorage.removeItem('megaport_auth_token');
    console.log('Token cleared');
  },
};
window.tokenManager = tokenManager;

// Auth helper
window.fetchAuthToken = async function (clientId, clientSecret, tokenUrl) {
  console.log('Starting authentication...');
  const env = tokenUrl.includes('staging')
    ? 'staging'
    : tokenUrl.includes('dev')
    ? 'development'
    : 'production';

  const cached = tokenManager.getToken(env);
  if (cached) {
    return JSON.stringify({ access_token: cached, cached: true });
  }

  const params = new URLSearchParams();
  params.append('grant_type', 'client_credentials');
  const authHeader = btoa(`${clientId}:${clientSecret}`);

  try {
    console.log('Fetching new auth token...');
    const start = performance.now();
    const resp = await fetch(tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        Authorization: `Basic ${authHeader}`,
      },
      body: params,
    });

    const elapsed = (performance.now() - start).toFixed(2);
    console.log(`Auth request: ${elapsed}ms`);

    if (!resp.ok) {
      const txt = await resp.text();
      throw new Error(`Authentication failed - HTTP ${resp.status}: ${txt}`);
    }
    const data = await resp.json();
    console.log('Auth successful');
    tokenManager.storeToken(env, data.access_token, data.expires_in);
    return JSON.stringify(data);
  } catch (e) {
    console.error('Auth error:', e.message);
    throw new Error(`Auth error: ${e.message}`);
  }
};

// API helpers
window.fetchApiEndpoint = async function (url, token, options = {}) {
  console.log(`API request: ${url}`);

  const fetchOpts = {
    method: options.method || 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    ...(options.body && { body: options.body }),
    ...options,
  };

  const controller = new AbortController();
  const timeoutId = setTimeout(() => {
    console.warn(`Request timeout (30s): ${url}`);
    controller.abort();
  }, 30000);

  const start = performance.now();
  try {
    const resp = await fetch(url, { ...fetchOpts, signal: controller.signal });
    const text = await resp.text();
    clearTimeout(timeoutId);

    const elapsed = (performance.now() - start).toFixed(2);
    console.log(`Response: ${text.length} chars in ${elapsed}ms`);

    return text;
  } catch (e) {
    clearTimeout(timeoutId);
    console.error(`Fetch error for ${url}:`, e.message);
    throw e;
  }
};

window.invokeApiRequest = function (url, token, options) {
  console.log(`Invoking API request to ${url}`);
  return new Promise((resolve) => {
    const t0 = performance.now();
    window
      .fetchApiEndpoint(url, token, options)
      .then((result) => {
        const dt = (performance.now() - t0).toFixed(2);
        console.log(
          `fetchApiEndpoint promise resolved successfully in ${dt}ms`
        );
        resolve({ result, error: null, processingTime: dt });
      })
      .catch((err) => {
        const dt = (performance.now() - t0).toFixed(2);
        console.error(
          'invokeApiRequest: fetchApiEndpoint promise rejected:',
          err.message
        );
        resolve({
          result: null,
          error: err.message || 'Unknown API error',
          processingTime: dt,
        });
      });
  });
};

// Simple request tracking
window.requestTracker = {
  pendingRequests: {},
  nextRequestId: 1,

  startRequest(url) {
    const id = this.nextRequestId++;
    this.pendingRequests[id] = {
      url,
      status: 'pending',
      startTime: performance.now(),
      result: null,
      error: null,
    };
    console.log(`Request #${id} started for ${url}`);
    return id;
  },

  updateRequest(id, status, resultData, errorMsg) {
    const req = this.pendingRequests[id];
    if (!req) {
      console.error(`Request #${id} not found`);
      return false;
    }
    req.status = status;
    if (resultData !== undefined) req.result = resultData;
    if (errorMsg !== undefined) req.error = errorMsg;
    req.completedAt = performance.now();
    const elapsed = ((req.completedAt - req.startTime) / 1000).toFixed(2);
    console.log(`Request #${id}: ${status} (${elapsed}s)`);
    return true;
  },

  cleanup() {
    const now = performance.now();
    for (const id in this.pendingRequests) {
      const req = this.pendingRequests[id];
      if (req.completedAt && now - req.completedAt > 300000) {
        delete this.pendingRequests[id];
      }
    }
  },
};

setInterval(() => window.requestTracker.cleanup(), 60000);

// Start API request
window.startApiRequest = function (url, token, options = {}) {
  const id = window.requestTracker.startRequest(url);
  console.log(`Starting API request #${id} to ${url}`);

  window
    .invokeApiRequest(url, token, options)
    .then((responseObject) => {
      console.log(
        `Request #${id} processing time: ${responseObject.processingTime}ms`
      );
      if (responseObject.error) {
        window.requestTracker.updateRequest(
          id,
          'error',
          null,
          responseObject.error
        );
      } else {
        window.requestTracker.updateRequest(
          id,
          'completed',
          responseObject,
          null
        );
      }

      if (typeof window.notifyRequestComplete === 'function') {
        try {
          window.notifyRequestComplete(id);
        } catch (e) {
          console.error(`Error notifying completion for #${id}:`, e);
        }
      }
    })
    .catch((err) => {
      console.error(`Request #${id} error:`, err);
      window.requestTracker.updateRequest(
        id,
        'error',
        null,
        err.message || 'Unknown error'
      );
      if (typeof window.notifyRequestComplete === 'function') {
        window.notifyRequestComplete(id);
      }
    });
  return id;
};

window.checkRequestStatus = function (id) {
  const req = window.requestTracker.pendingRequests[id];
  if (!req) {
    return { status: 'not_found' };
  }

  const status = req.status;
  const elapsed = performance.now() - req.startTime;

  const out = { status, url: req.url, elapsedMs: elapsed };

  if (status === 'completed' && req.result && req.result.result) {
    out.result = req.result.result;
  }

  if (req.error) {
    out.error = req.error;
  }

  return out;
};

// Check fetch request status
window.checkFetchRequestStatus = function (requestId) {
  const id = typeof requestId === 'string' ? requestId : String(requestId);
  const status = window.checkRequestStatus(id);

  return {
    completed: status.status === 'completed' || status.status === 'error',
    success: status.status === 'completed',
    data: status.result || null,
    error: status.error || null,
  };
};

// Debugging function
window.debugRequestState = function () {
  console.group('🔧 Request State');
  console.log(
    'Pending requests:',
    Object.keys(window.requestTracker.pendingRequests)
  );
  console.log(
    'notifyRequestComplete available:',
    typeof window.notifyRequestComplete
  );
  console.groupEnd();
};

window.wasmDebug = (msg) => {
  console.log('[WASM Debug]', msg);
  appendToTerminal('[Debug] ' + msg, 'system');
};

// WASM init & input handling
const go = new Go();
WebAssembly.instantiateStreaming(fetch('megaport.wasm'), go.importObject)
  .then((res) => {
    console.log('WASM module loaded successfully');
    go.run(res.instance);

    console.log('Go WASM program has started.');

    if (typeof window.registerAuthFunction === 'function') {
      console.log('Auth fn registration check (JS side): function exists.');
    } else {
      console.warn(
        'Auth fn registration check (JS side): window.registerAuthFunction not found.'
      );
    }
    if (typeof window.browserApiRequest === 'function') {
      console.log(
        'API helper (browserApiRequest) check (JS side): function exists.'
      );
    }
    if (typeof window.executeMegaportCommand === 'function') {
      console.log('executeMegaportCommand check (JS side): function exists.');
    } else {
      console.warn(
        'executeMegaportCommand check (JS side): window.executeMegaportCommand not found.'
      );
    }
  })
  .catch((err) => {
    appendToTerminal('WASM load error: ' + err, 'error');
    console.error('WASM instantiation or run error:', err);
  });

document.getElementById('input').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') {
    const cmd = e.target.value.trim();
    if (!cmd) return;
    appendToTerminal('megaport> ' + cmd);

    // Clear input immediately
    e.target.value = '';

    try {
      // Try async version first (preferred for commands that need auth/API calls)
      console.log(
        'Checking for async function:',
        typeof window.executeMegaportCommandAsync
      );
      if (typeof window.executeMegaportCommandAsync === 'function') {
        console.log('🚀 Using async command execution for:', cmd);
        appendToTerminal('⏳ Processing...', 'system');

        window.executeMegaportCommandAsync(cmd, function (result) {
          console.log('✅ Async command completed', result);

          // Remove the "Processing..." message
          const terminalDiv = document.getElementById('terminal');
          const lastLine = terminalDiv.lastElementChild;
          if (lastLine && lastLine.textContent.includes('Processing...')) {
            terminalDiv.removeChild(lastLine);
          }

          if (result && result.output) {
            appendToTerminal(result.output);
          } else if (result && result.error) {
            appendToTerminal('Error: ' + result.error, 'error');
          } else {
            appendToTerminal('Command completed (no output)', 'system');
          }
        });

        return; // Exit early - callback will handle the rest
      }

      // Fallback to sync version (may not work for async operations)
      if (typeof window.executeMegaportCommand !== 'function') {
        appendToTerminal(
          'Error: No command execution function available. WASM module may not be ready.',
          'error'
        );
        return;
      }

      console.log(
        '⚠️ Using sync command execution (may block on async operations)'
      );
      const out = window.executeMegaportCommand(cmd);
      if (out && out.output) appendToTerminal(out.output);
      else if (out && out.error)
        appendToTerminal('Error: ' + out.error, 'error');
      else if (out === undefined && cmd.startsWith('exit')) {
        /* Graceful exit */
      } else
        appendToTerminal(
          'No output or unexpected return from command.',
          'system'
        );
    } catch (err) {
      appendToTerminal('Execution error: ' + err.message, 'error');
      console.error('Error executing command via WASM:', err);
    }
  }
});
