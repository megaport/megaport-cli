// Session-based authentication for Megaport CLI WASM
(function () {
  'use strict';

  // Session management
  const SESSION_STORAGE_KEY = 'megaport_session';
  const SESSION_EXPIRY_KEY = 'megaport_session_expiry';

  class SessionManager {
    constructor() {
      this.sessionToken = null;
      this.expiresAt = null;
      this.loadSession();
    }

    // Load session from localStorage
    loadSession() {
      const token = localStorage.getItem(SESSION_STORAGE_KEY);
      const expiry = localStorage.getItem(SESSION_EXPIRY_KEY);

      if (token && expiry) {
        const expiryTime = parseInt(expiry, 10);
        if (Date.now() < expiryTime) {
          this.sessionToken = token;
          this.expiresAt = expiryTime;
          console.log('Session restored from storage');
          return true;
        } else {
          console.log('Stored session expired, clearing...');
          this.clearSession();
        }
      }
      return false;
    }

    // Save session to localStorage
    saveSession(token, expiresIn, environment, accessKey, secretKey) {
      this.sessionToken = token;
      this.expiresAt = Date.now() + expiresIn * 1000;

      localStorage.setItem(SESSION_STORAGE_KEY, token);
      localStorage.setItem(SESSION_EXPIRY_KEY, this.expiresAt.toString());
      
      // Store credentials for WASM (needed after page reload)
      if (accessKey && secretKey && environment) {
        localStorage.setItem('megaport_access_key', accessKey);
        localStorage.setItem('megaport_secret_key', secretKey);
        localStorage.setItem('megaport_environment', environment);
      }

      console.log(`Session saved, expires in ${expiresIn}s`);
    }
    
    // Restore credentials from localStorage
    restoreCredentials() {
      const accessKey = localStorage.getItem('megaport_access_key');
      const secretKey = localStorage.getItem('megaport_secret_key');
      const environment = localStorage.getItem('megaport_environment');
      
      if (accessKey && secretKey && environment) {
        window.megaportCredentials = {
          accessKey,
          secretKey,
          environment,
        };
        console.log('✅ Credentials restored for WASM');
        return true;
      }
      return false;
    }

    // Clear session
    clearSession() {
      this.sessionToken = null;
      this.expiresAt = null;
      localStorage.removeItem(SESSION_STORAGE_KEY);
      localStorage.removeItem(SESSION_EXPIRY_KEY);
      localStorage.removeItem('megaport_access_key');
      localStorage.removeItem('megaport_secret_key');
      localStorage.removeItem('megaport_environment');
      console.log('Session cleared');
    }

    // Check if session is valid
    isValid() {
      return this.sessionToken && this.expiresAt && Date.now() < this.expiresAt;
    }

    // Get session token
    getToken() {
      return this.isValid() ? this.sessionToken : null;
    }

    // Get time remaining (in seconds)
    getTimeRemaining() {
      if (!this.isValid()) return 0;
      return Math.floor((this.expiresAt - Date.now()) / 1000);
    }
  }

  // Create global session manager
  window.sessionManager = new SessionManager();

  // Login function
  window.loginToMegaport = async function (
    accessKey,
    secretKey,
    environment = 'production'
  ) {
    console.log(`Logging in to ${environment}...`);

    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          accessKey,
          secretKey,
          environment,
        }),
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Login failed: ${error}`);
      }

      const data = await response.json();

      // Save session
      window.sessionManager.saveSession(data.sessionToken, data.expiresIn);

      console.log('✅ Login successful');
      return {
        success: true,
        sessionToken: data.sessionToken,
        expiresIn: data.expiresIn,
        environment: data.environment,
      };
    } catch (error) {
      console.error('❌ Login error:', error);
      throw error;
    }
  };

  // Logout function
  window.logoutFromMegaport = async function () {
    const token = window.sessionManager.getToken();
    if (!token) {
      console.log('No active session to logout');
      return;
    }

    try {
      await fetch('/auth/logout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-Token': token,
        },
      });

      window.sessionManager.clearSession();
      console.log('✅ Logged out successfully');
    } catch (error) {
      console.error('❌ Logout error:', error);
      // Clear session anyway
      window.sessionManager.clearSession();
    }
  };

  // Check session validity
  window.checkSession = async function () {
    const token = window.sessionManager.getToken();
    if (!token) {
      return { valid: false, error: 'No session token' };
    }

    try {
      const response = await fetch('/auth/check', {
        headers: {
          'X-Session-Token': token,
        },
      });

      if (!response.ok) {
        window.sessionManager.clearSession();
        return { valid: false, error: 'Session expired or invalid' };
      }

      const data = await response.json();
      return {
        valid: true,
        expiresAt: data.expiresAt,
        timeRemaining: window.sessionManager.getTimeRemaining(),
      };
    } catch (error) {
      console.error('Session check error:', error);
      return { valid: false, error: error.message };
    }
  };

  // Authenticated API request wrapper
  window.authenticatedFetch = async function (endpoint, options = {}) {
    const token = window.sessionManager.getToken();
    if (!token) {
      throw new Error('Not authenticated. Please login first.');
    }

    // Prepend /api/ if not already there
    const url = endpoint.startsWith('/api/') ? endpoint : `/api/${endpoint}`;

    // Add session token to headers
    const headers = {
      ...options.headers,
      'X-Session-Token': token,
    };

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      // Handle 401 (session expired)
      if (response.status === 401) {
        window.sessionManager.clearSession();
        // Trigger session expired event
        if (window.onSessionExpired) {
          window.onSessionExpired();
        }
        throw new Error('Session expired. Please login again.');
      }

      return response;
    } catch (error) {
      console.error('Authenticated fetch error:', error);
      throw error;
    }
  };

  // Login function to authenticate with the server and set WASM environment variables
  window.loginToMegaport = async function (accessKey, secretKey, environment) {
    console.log('🔐 Starting login process...', { environment });

    try {
      // Call the server's login endpoint
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          accessKey,
          secretKey,
          environment,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Login failed: ${errorText}`);
      }

      const data = await response.json();
      console.log('✅ Login successful, session token received');

      // Save session with credentials
      window.sessionManager.saveSession(
        data.sessionToken,
        data.expiresIn,
        environment,
        accessKey,
        secretKey
      );

      // CRITICAL: Store credentials in a global that the WASM can read
      // The WASM binary will check window.megaportCredentials first before checking environment variables
      console.log('Setting credentials for WASM binary...');
      window.megaportCredentials = {
        accessKey,
        secretKey,
        environment,
      };
      console.log('✅ Credentials stored for WASM');

      return data;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  // Logout function
  window.logoutFromMegaport = async function () {
    console.log('🔓 Logging out...');

    const token = window.sessionManager.getToken();
    if (token) {
      try {
        await fetch('/auth/logout', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Session-Token': token,
          },
        });
      } catch (error) {
        console.error('Logout error:', error);
      }
    }

    // Clear session
    window.sessionManager.clearSession();

    // Clear WASM credentials
    if (window.megaportCredentials) {
      delete window.megaportCredentials;
    }

    console.log('✅ Logged out successfully');
  };

  // Helper to get Megaport locations (example usage)
  window.getMegaportLocations = async function () {
    try {
      const response = await window.authenticatedFetch('v2/locations');
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Failed to fetch locations:', error);
      throw error;
    }
  };

  // Verify session with server on page load
  window.verifySessionOnLoad = async function () {
    if (window.sessionManager.isValid()) {
      console.log('✅ Active session found in localStorage');
      console.log(
        `⏱️  Session expires in ${window.sessionManager.getTimeRemaining()}s`
      );

      // CRITICAL: Restore credentials for WASM first
      const credentialsRestored = window.sessionManager.restoreCredentials();
      if (!credentialsRestored) {
        console.log('⚠️  Could not restore credentials, forcing re-login...');
        window.sessionManager.clearSession();
        return false;
      }

      // CRITICAL: Verify session with server (container might have restarted)
      const result = await window.checkSession();
      if (!result.valid) {
        console.log(
          '⚠️  Session invalid on server (container may have restarted), clearing...'
        );
        window.sessionManager.clearSession();
        if (window.megaportCredentials) {
          delete window.megaportCredentials;
        }
        return false;
      } else {
        console.log('✅ Session verified with server');
        console.log('✅ Credentials restored for WASM');
        return true;
      }
    } else {
      console.log('ℹ️  No active session. Please login.');
      return false;
    }
  };

  console.log('✅ Session management initialized');
})();
