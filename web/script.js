// Add this at the VERY TOP of script.js before any other code
(function initGlobalFunctions() {
  // Direct browser fetch implementation that bypasses worker issues
  window.directBrowserFetch = function (url, token, onSuccess, onError) {
    console.log(`🔥 DIRECT FETCH: Starting direct browser fetch to ${url}`);

    // Use XMLHttpRequest for maximum compatibility and direct control
    const xhr = new XMLHttpRequest();
    xhr.open('GET', url, true);
    xhr.setRequestHeader('Authorization', `Bearer ${token}`);
    xhr.setRequestHeader('Content-Type', 'application/json');

    // Set up response handler
    xhr.onreadystatechange = function () {
      if (xhr.readyState === 4) {
        console.log(`🔥 DIRECT FETCH: Completed with status ${xhr.status}`);

        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            console.log(
              `🔥 DIRECT FETCH: Success, data length: ${xhr.responseText.length} bytes`
            );
            onSuccess(xhr.responseText);
          } catch (e) {
            console.error('Error in success callback:', e);
            onError(`Callback error: ${e.message}`);
          }
        } else {
          onError(`HTTP Error: ${xhr.status} ${xhr.statusText}`);
        }
      }
    };

    // Handle network errors
    xhr.onerror = function () {
      console.error('XHR Network error');
      onError('Network error occurred');
    };

    // Send the request
    try {
      xhr.send();
      console.log(`🔥 DIRECT FETCH: Request sent`);
    } catch (e) {
      console.error('XHR Send error:', e);
      onError(`Send error: ${e.message}`);
    }
  };

  console.log('✅ Global API helper functions registered');
})();

// Terminal helpers
function appendToTerminal(text, className) {
  const terminal = document.getElementById('terminal');
  const line = document.createElement('div');
  if (className) line.className = className;
  line.textContent = text;
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

// Initialize web worker for API requests
let apiWorker;
try {
  apiWorker = new Worker('worker.js');
  console.log('API Worker initialized');

  // Enhanced worker message handler with better debugging and synchronization
  apiWorker.onmessage = function (e) {
    const { requestId, status, result, error, headers, processingTime, type } =
      e.data;

    // Handle heartbeat messages
    if (type === 'heartbeat') {
      console.log(`Worker heartbeat: ${e.data.message}`);
      return;
    }

    // Log ALL worker messages for debugging
    console.log(
      `[${performance
        .now()
        .toFixed(
          1
        )}ms] Worker message received for #${requestId}: ${status} (${processingTime}ms)`
    );

    // CRITICAL: Find the request in our tracker
    const req = window.requestTracker.pendingRequests[requestId];
    if (!req) {
      console.error(
        `Worker message for unknown request #${requestId} - available requests:`,
        Object.keys(window.requestTracker.pendingRequests)
      );
      return;
    }

    let resultData = null;
    let errorMsg = null;

    // Handle different status types
    if (
      status === 'headers_received' ||
      status.startsWith('headers_received')
    ) {
      console.log(`Processing headers_received for #${requestId}`);
      resultData = { headers, processingTime };
      // Store in responseLookup for direct access
      if (req.url) {
        window.responseLookup = window.responseLookup || {};
        window.responseLookup[req.url] = {
          timestamp: Date.now(),
          headers: headers,
          status: headers.status,
          partial: true,
        };
      }
    } else if (status === 'completed' || status.startsWith('completed')) {
      console.log(`Processing completed for #${requestId}`);
      resultData = { result, processingTime };
      // Update responseLookup
      if (req.url) {
        window.responseLookup = window.responseLookup || {};
        window.responseLookup[req.url] = {
          timestamp: Date.now(),
          data: result,
          status: 200,
          partial: false,
        };
      }
    } else if (status === 'error' || status.startsWith('error')) {
      console.log(`Processing error for #${requestId}: ${error}`);
      errorMsg = error;
    } else if (status === 'processing' || status === 'partial_result') {
      console.log(
        `Processing intermediate status for #${requestId}: ${status}`
      );
      resultData = e.data; // Store the entire data object for these intermediate states
    }

    // Update request status
    const updateSuccess = window.requestTracker.updateRequest(
      requestId,
      status,
      resultData,
      errorMsg
    );

    if (!updateSuccess) {
      console.error(`Failed to update request #${requestId} in tracker`);
      return;
    }

    // HIGH PRIORITY: Notify Go IMMEDIATELY about the status change
    if (typeof window.notifyRequestComplete === 'function') {
      try {
        console.log(`Calling notifyRequestComplete for #${requestId}`);
        window.notifyRequestComplete(requestId);
      } catch (e) {
        console.error(
          `Error notifying Go about request #${requestId} update:`,
          e
        );
      }
    } else {
      console.error('window.notifyRequestComplete is not available!');
    }
  };

  apiWorker.onerror = function (e) {
    console.error('Worker error:', e);
  };

  // Add debugging wrapper for postMessage
  const originalPostMessage = apiWorker.postMessage;
  apiWorker.postMessage = function (data) {
    console.log(`📤 Sending to worker:`, data);
    return originalPostMessage.call(this, data);
  };
} catch (e) {
  console.error('Failed to initialize Web Worker:', e);
  apiWorker = null;
}

// Auth helper
window.fetchAuthToken = async function (clientId, clientSecret, tokenUrl) {
  console.log('Starting browser-based authentication…');
  const env = tokenUrl.includes('staging')
    ? 'staging'
    : tokenUrl.includes('dev')
    ? 'development'
    : 'production';

  const cached = tokenManager.getToken(env);
  if (cached) {
    return JSON.stringify({ access_token: cached, cached: true });
  }

  const timeoutPromise = new Promise((_, rej) =>
    setTimeout(
      () => rej(new Error('Auth request timed out after 30 seconds')),
      30000
    )
  );
  const params = new URLSearchParams();
  params.append('grant_type', 'client_credentials');
  const authHeader = btoa(`${clientId}:${clientSecret}`);

  try {
    console.log('Fetching new auth token…');
    const start = performance.now();
    const resp = await Promise.race([
      fetch(tokenUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          Authorization: `Basic ${authHeader}`,
        },
        body: params,
      }),
      timeoutPromise,
    ]);
    console.log(
      `Auth fetch call took ${(performance.now() - start).toFixed(2)}ms`
    );

    if (!resp.ok) {
      const txt = await resp.text();
      throw new Error(`Authentication failed - HTTP ${resp.status}: ${txt}`);
    }
    const data = await resp.json();
    console.log('Auth successful, new token obtained.');
    tokenManager.storeToken(env, data.access_token, data.expires_in);
    return JSON.stringify(data);
  } catch (e) {
    console.error('Auth error:', e.message);
    throw new Error(`Auth error: ${e.message}`);
  }
};

// API helpers
window.fetchApiEndpoint = async function (url, token, options = {}) {
  console.log(`Starting browser API request to: ${url}`);

  // If we have a worker available, use it
  if (apiWorker) {
    console.log(`Using Web Worker for API request to ${url}`);
    return new Promise((resolve, reject) => {
      const workerRequestId = `worker_${Date.now()}_${Math.random()
        .toString(36)
        .substr(2, 9)}`;

      const timeoutId = setTimeout(() => {
        reject(new Error('API request timed out after 210 seconds'));
      }, 210000);

      const messageHandler = function (e) {
        if (e.data && e.data.requestId === workerRequestId) {
          clearTimeout(timeoutId);
          apiWorker.removeEventListener('message', messageHandler);

          if (e.data.status === 'completed') {
            resolve(e.data.result);
          } else {
            reject(new Error(e.data.error || 'Unknown worker error'));
          }
        }
      };

      apiWorker.addEventListener('message', messageHandler);
      apiWorker.postMessage({
        requestId: workerRequestId,
        url,
        token,
        options,
      });
    });
  }

  // Fallback to the original implementation if worker is not available
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
    console.warn(
      `fetchApiEndpoint: Aborting request to ${url} due to JS timeout (210s)`
    );
    controller.abort();
  }, 210000);

  const start = performance.now();
  try {
    console.log(`API request: ${fetchOpts.method} ${url}`);
    const resp = await fetch(url, { ...fetchOpts, signal: controller.signal });

    console.log(
      `Response received, yielding to event loop before processing body`
    );
    await new Promise((resolve) => setTimeout(resolve, 0));

    const text = await resp.text();
    clearTimeout(timeoutId);

    console.log(
      `Fetched ${text.length} chars in ${(performance.now() - start).toFixed(
        2
      )}ms`
    );

    for (let i = 0; i < 5; i++) {
      await new Promise((resolve) => setTimeout(resolve, 0));
    }

    return text;
  } catch (e) {
    clearTimeout(timeoutId);
    if (e.name === 'AbortError') {
      console.error(
        `fetchApiEndpoint error for ${url}: Request aborted (likely timeout).`
      );
    } else {
      console.error(`fetchApiEndpoint error for ${url}:`, e.message);
    }
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

// Request tracking
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
      console.error(`updateRequest: Request #${id} not found in tracker.`);
      return false;
    }
    const prevStatus = req.status;
    req.status = status;

    if (resultData !== undefined) req.result = resultData;
    if (errorMsg !== undefined) req.error = errorMsg;

    req.completedAt = performance.now();
    const elapsed = ((req.completedAt - req.startTime) / 1000).toFixed(2);
    console.log(
      `Request #${id} updated: ${prevStatus} → ${status} (took ${elapsed}s)`
    );
    if (errorMsg) console.error(`Request #${id} error: ${errorMsg}`);
    return true;
  },

  cleanup() {
    const now = performance.now();
    for (const id in this.pendingRequests) {
      const req = this.pendingRequests[id];
      if (req.completedAt && now - req.completedAt > 300000) {
        console.log(`Cleaning up completed request #${id}`);
        delete this.pendingRequests[id];
      }
    }
  },
};

setInterval(() => window.requestTracker.cleanup(), 60000);

// Enhanced startApiRequest function with better error handling
window.startApiRequest = function (url, token, options = {}) {
  const id = window.requestTracker.startRequest(url);
  console.log(`Starting non-blocking API request #${id} to ${url}`);

  // Special handling for worker-based requests
  if (apiWorker) {
    try {
      // Send directly to worker - CRITICAL: use the same ID
      apiWorker.postMessage({
        requestId: id, // Make sure this matches what onmessage expects
        url,
        token,
        options,
      });
      console.log(`Request #${id} sent to worker thread`);
      return id;
    } catch (err) {
      console.error(
        `Failed to use worker for request #${id}, falling back to main thread:`,
        err
      );
      // Fall through to normal method
    }
  }

  // Normal method if worker is unavailable
  window
    .invokeApiRequest(url, token, options)
    .then((responseObject) => {
      console.log(
        `invokeApiRequest for #${id} .then() called. Processing time: ${responseObject.processingTime}ms`
      );
      if (responseObject.error) {
        console.error(
          `Request #${id} completed with error: ${responseObject.error}`
        );
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
        console.log(`Notifying Go about request #${id} completion/error.`);
        try {
          window.notifyRequestComplete(id);
        } catch (e) {
          console.error(`Error calling notifyRequestComplete for #${id}:`, e);
        }
      }
    })
    .catch((err) => {
      console.error(
        `Critical error in startApiRequest promise chain for #${id}:`,
        err
      );
      window.requestTracker.updateRequest(
        id,
        'error',
        null,
        err.message || 'Unknown critical error'
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

  // Normalize any redundant JS statuses back to the four canonical states
  let status = req.status;
  if (status.startsWith('headers_received')) {
    status = 'headers_received';
  } else if (status.startsWith('processing')) {
    status = 'processing';
  } else if (status.startsWith('partial_result')) {
    status = 'partial_result';
  } else if (status.startsWith('completed')) {
    status = 'completed';
  } else if (status.startsWith('error')) {
    status = 'error';
  }
  req.status = status;

  const elapsed = performance.now() - req.startTime;

  // Attempt a direct lookup if still pending
  if (status === 'pending' && req.url && window.responseLookup) {
    const cached = window.responseLookup[req.url];
    if (cached) {
      if (cached.partial && cached.headers) {
        status = 'headers_received';
        req.status = status;
        req.result = {
          headers: cached.headers,
          processingTime: elapsed.toFixed(2),
        };
      } else if (!cached.partial && cached.data) {
        status = 'completed';
        req.status = status;
        req.result = {
          result: cached.data,
          processingTime: elapsed.toFixed(2),
        };
        delete window.responseLookup[req.url];
      }
    }
  }

  // Log at most once every 5 seconds
  if (!req._lastLogTime || elapsed - req._lastLogTime > 5000) {
    console.log(
      `Status check #${id}: ${status} (${(elapsed / 1000).toFixed(1)}s elapsed)`
    );
    req._lastLogTime = elapsed;
  }

  // Build the return object
  const out = { status, url: req.url, elapsedMs: elapsed };

  if (status === 'headers_received' && req.result && req.result.headers) {
    out.headers = req.result.headers;
    out.headerProcessingTime = req.result.processingTime;
    out.partial = true;
  } else if (status === 'completed' && req.result && req.result.result) {
    out.result = req.result.result;
  }

  if (req.error) {
    out.error = req.error;
  }

  return out;
};

// Force a check for a specific request, useful when notifications might be missed
window.forceRequestStatusCheck = function (id) {
  const req = window.requestTracker.pendingRequests[id];
  if (!req) {
    console.warn(`forceRequestStatusCheck: Request #${id} not found`);
    return false;
  }

  // Check if the URL for this request is in the responseLookup
  if (req.url && window.responseLookup && window.responseLookup[req.url]) {
    const cachedResponse = window.responseLookup[req.url];

    // Check if this cached response is already processed
    if (cachedResponse.processed) {
      return false;
    }

    // Mark as processed to avoid redundant notifications
    cachedResponse.processed = true;

    // If we have a full response but the status isn't completed, fix it
    if (
      !cachedResponse.partial &&
      cachedResponse.data &&
      req.status !== 'completed'
    ) {
      console.warn(
        `Force status check found completed response for request #${id} - updating status`
      );

      window.requestTracker.updateRequest(
        id,
        'completed',
        {
          result: cachedResponse.data,
          processingTime: (performance.now() - req.startTime).toFixed(2),
        },
        null
      );

      // Notify Go about this update
      if (typeof window.notifyRequestComplete === 'function') {
        try {
          window.notifyRequestComplete(id);
          return true;
        } catch (e) {
          console.error(`Error in force notification for #${id}:`, e);
        }
      }
    }
  }

  return false;
};

// Enhanced WASM status listener with aggressive polling
window.registerWasmStatusListener = function (requestId) {
  const req = window.requestTracker.pendingRequests[requestId];
  if (!req) return false;

  console.log(
    `Registering enhanced WASM status listener for request #${requestId}`
  );

  // CRITICAL: Use polling within JavaScript to force notification to Go
  const statusInterval = setInterval(() => {
    // Always check responseLookup first as it might have data that hasn't been processed
    if (req.url && window.responseLookup && window.responseLookup[req.url]) {
      const cachedResponse = window.responseLookup[req.url];

      if (!cachedResponse.notifiedWasm) {
        cachedResponse.notifiedWasm = true;
        console.log(
          `Status listener found response for #${requestId}, forcing notification`
        );

        // Ensure we notify Go immediately with high priority
        if (typeof window.notifyRequestComplete === 'function') {
          try {
            // Use both immediate and scheduled notifications for redundancy
            window.notifyRequestComplete(requestId);
            setTimeout(() => window.notifyRequestComplete(requestId), 50);
            setTimeout(() => window.notifyRequestComplete(requestId), 100);
          } catch (e) {
            console.error(
              `Error in status listener notification for #${requestId}:`,
              e
            );
          }
        }
      }
    }

    // Stop polling once request is completed or 60 seconds have passed
    const elapsed = (performance.now() - req.startTime) / 1000;
    if (req.status === 'completed' || req.status === 'error' || elapsed > 60) {
      console.log(
        `Status listener for #${requestId} stopping (status: ${
          req.status
        }, elapsed: ${elapsed.toFixed(1)}s)`
      );
      clearInterval(statusInterval);
    }
  }, 200); // Check every 200ms (aggressive polling)

  return true;
};

window.checkResponseLookup = function (url) {
  if (!url || !window.responseLookup) return null;

  const cachedResponse = window.responseLookup[url];
  if (!cachedResponse) return null;

  console.log(`checkResponseLookup found cached response for ${url}`);
  return {
    exists: true,
    partial: !!cachedResponse.partial,
    complete: !cachedResponse.partial && !!cachedResponse.data,
    timestamp: cachedResponse.timestamp || Date.now(),
  };
};

// Add response cache to track and correlate network responses with pending requests
if (!window.responseLookup) {
  window.responseLookup = {};

  // Monitor fetch responses globally
  const originalFetch = window.fetch;
  window.fetch = async function (url, options) {
    const response = await originalFetch(url, options);

    if (response.ok) {
      // Clone the response to avoid consuming it
      const clonedResponse = response.clone();

      // Process the response asynchronously to avoid blocking
      clonedResponse
        .text()
        .then((text) => {
          // Store in our lookup table
          const urlObj = new URL(url, window.location.href);
          const urlKey = urlObj.toString();

          window.responseLookup[urlKey] = {
            timestamp: Date.now(),
            data: text,
            status: response.status,
          };

          console.log(
            `🔄 Response for ${urlKey} cached (${text.length} bytes)`
          );
        })
        .catch((e) => console.error('Error caching response:', e));
    }

    return response;
  };
}

// Enhanced debugging function
window.debugWorkerState = function () {
  console.group('🔧 Worker Debug State');
  console.log('Worker available:', !!apiWorker);
  console.log(
    'Pending requests:',
    Object.keys(window.requestTracker.pendingRequests)
  );
  console.log(
    'Response lookup keys:',
    window.responseLookup ? Object.keys(window.responseLookup) : 'none'
  );
  console.log(
    'notifyRequestComplete available:',
    typeof window.notifyRequestComplete
  );
  console.groupEnd();
};

const originalFetchApiEndpoint = window.fetchApiEndpoint;
window.fetchApiEndpoint = async function (url, token, options = {}) {
  const result = await originalFetchApiEndpoint(url, token, options);

  // Find which request this might belong to
  for (const id in window.requestTracker.pendingRequests) {
    if (window.requestTracker.pendingRequests[id].url === url) {
      window._lastFetchCompletedId = parseInt(id);
      window._lastFetchResult = { result };
      console.log(`Marked request #${id} as potentially completed`);
      break;
    }
  }

  return result;
};

window.wasmDebug = (msg) => {
  console.log('[WASM Debug]', msg);
  appendToTerminal('[Debug] ' + msg, 'system');
};

// Add the waitUntilRequestComplete function to help with WASM-JS synchronization
window.waitUntilRequestComplete = function (requestId, timeoutMs = 30000) {
  return new Promise((resolve, reject) => {
    const startTime = Date.now();
    const checkInterval = 100; // milliseconds

    // Check function
    function checkStatus() {
      if (!window.requestTracker.pendingRequests[requestId]) {
        return reject(new Error(`Request #${requestId} not found in tracker`));
      }

      const req = window.requestTracker.pendingRequests[requestId];

      // Check if request is complete
      if (req.status === 'completed') {
        console.log(`Request #${requestId} is complete, resolving promise`);
        return resolve(req.result);
      } else if (req.status === 'error') {
        console.error(`Request #${requestId} failed with error: ${req.error}`);
        return reject(new Error(req.error || 'Unknown request error'));
      }

      // Check for timeout
      if (Date.now() - startTime > timeoutMs) {
        console.error(`Request #${requestId} timed out after ${timeoutMs}ms`);
        return reject(new Error(`Request timed out after ${timeoutMs}ms`));
      }

      // Continue checking
      setTimeout(checkStatus, checkInterval);
    }

    // Start checking
    checkStatus();
  });
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
    try {
      if (typeof window.executeMegaportCommand !== 'function') {
        appendToTerminal(
          'Error: executeMegaportCommand is not available. WASM module may not be ready or failed to load.',
          'error'
        );
        return;
      }
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
    e.target.value = '';
  }
});
