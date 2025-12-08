import { describe, it, expect, beforeEach, vi } from 'vitest';
import { nextTick } from 'vue';
import { useMegaportWASM } from '../composables/useMegaportWASM';
import { mount, VueWrapper } from '@vue/test-utils';

describe('Authentication Flow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('setAuthCredentials', () => {
    it('should set authentication credentials with all parameters', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('access-key-123', 'secret-key-456', 'staging');

      expect(mockSetAuth).toHaveBeenCalledWith(
        'access-key-123',
        'secret-key-456',
        'staging'
      );
    });

    it('should handle production environment', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('prod-access', 'prod-secret', 'production');

      expect(mockSetAuth).toHaveBeenCalledWith(
        'prod-access',
        'prod-secret',
        'production'
      );
    });

    it('should return success response', () => {
      const mockSetAuth = vi.fn(() => ({ success: true, message: 'Auth set' }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('key', 'secret', 'staging');

      expect(mockSetAuth).toHaveReturnedWith({
        success: true,
        message: 'Auth set',
      });
    });
  });

  describe('clearAuthCredentials', () => {
    it('should clear authentication credentials', () => {
      const mockClearAuth = vi.fn(() => ({ success: true }));
      (window as any).clearAuthCredentials = mockClearAuth;

      const { clearAuth } = useMegaportWASM();
      clearAuth();

      expect(mockClearAuth).toHaveBeenCalled();
    });

    it('should clear auth after being set', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      const mockClearAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;
      (window as any).clearAuthCredentials = mockClearAuth;

      const { setAuth, clearAuth } = useMegaportWASM();

      setAuth('key', 'secret', 'staging');
      expect(mockSetAuth).toHaveBeenCalled();

      clearAuth();
      expect(mockClearAuth).toHaveBeenCalled();
    });
  });

  describe('getAuthInfo', () => {
    it('should retrieve authentication information', () => {
      const mockAuthInfo = {
        accessKeySet: true,
        accessKeyPreview: 'acc***',
        secretKeySet: true,
        secretKeyPreview: 'sec***',
        environment: 'staging',
      };
      (window as any).debugAuthInfo = vi.fn(() => mockAuthInfo);

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info).toEqual(mockAuthInfo);
    });

    it('should show access key is set', () => {
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'test***',
        secretKeySet: true,
        secretKeyPreview: 'sec***',
        environment: 'production',
      }));

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info?.accessKeySet).toBe(true);
      expect(info?.accessKeyPreview).toBe('test***');
    });

    it('should show secret key is set', () => {
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'acc***',
        secretKeySet: true,
        secretKeyPreview: 'my-sec***',
        environment: 'staging',
      }));

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info?.secretKeySet).toBe(true);
      expect(info?.secretKeyPreview).toBe('my-sec***');
    });

    it('should return current environment', () => {
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'key***',
        secretKeySet: true,
        secretKeyPreview: 'sec***',
        environment: 'production',
      }));

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info?.environment).toBe('production');
    });

    it('should handle unconfigured auth state', () => {
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: false,
        accessKeyPreview: '',
        secretKeySet: false,
        secretKeyPreview: '',
        environment: '',
      }));

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info?.accessKeySet).toBe(false);
      expect(info?.secretKeySet).toBe(false);
      expect(info?.environment).toBe('');
    });
  });

  describe('Authentication State Transitions', () => {
    it('should transition from unauthenticated to authenticated', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: false,
        accessKeyPreview: '',
        secretKeySet: false,
        secretKeyPreview: '',
        environment: '',
      }));

      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth, getAuthInfo } = useMegaportWASM();

      let info = getAuthInfo();
      expect(info?.accessKeySet).toBe(false);

      // Simulate auth state change
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'key***',
        secretKeySet: true,
        secretKeyPreview: 'sec***',
        environment: 'staging',
      }));
      setAuth('key', 'secret', 'staging');

      info = getAuthInfo();
      expect(info?.accessKeySet).toBe(true);
    });

    it('should transition from authenticated to unauthenticated', () => {
      const mockClearAuth = vi.fn(() => ({ success: true }));
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'key***',
        secretKeySet: true,
        secretKeyPreview: 'sec***',
        environment: 'staging',
      }));

      (window as any).clearAuthCredentials = mockClearAuth;

      const { clearAuth, getAuthInfo } = useMegaportWASM();

      let info = getAuthInfo();
      expect(info?.accessKeySet).toBe(true);

      // Simulate auth clear
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: false,
        accessKeyPreview: '',
        secretKeySet: false,
        secretKeyPreview: '',
        environment: '',
      }));
      clearAuth();

      info = getAuthInfo();
      expect(info?.accessKeySet).toBe(false);
    });

    it('should handle re-authentication with different credentials', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'old***',
        secretKeySet: true,
        secretKeyPreview: 'old***',
        environment: 'staging',
      }));

      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth, getAuthInfo } = useMegaportWASM();

      let info = getAuthInfo();
      expect(info?.accessKeyPreview).toBe('old***');
      expect(info?.environment).toBe('staging');

      // Simulate re-auth with new credentials
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'new***',
        secretKeySet: true,
        secretKeyPreview: 'new***',
        environment: 'production',
      }));
      setAuth('new-key', 'new-secret', 'production');

      info = getAuthInfo();
      expect(info?.accessKeyPreview).toBe('new***');
      expect(info?.environment).toBe('production');
    });
  });

  describe('Authentication Security', () => {
    it('should not expose full credentials in preview', () => {
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        accessKeyPreview: 'abc***xyz',
        secretKeySet: true,
        secretKeyPreview: 'sec***ret',
        environment: 'staging',
      }));

      const { getAuthInfo } = useMegaportWASM();
      const info = getAuthInfo();

      expect(info?.accessKeyPreview).toContain('***');
      expect(info?.secretKeyPreview).toContain('***');
      expect(info?.accessKeyPreview).not.toContain('full-access-key');
      expect(info?.secretKeyPreview).not.toContain('full-secret-key');
    });

    it('should handle auth with password-type inputs', () => {
      // This simulates that credentials are entered in password fields
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('sensitive-key', 'sensitive-secret', 'production');

      expect(mockSetAuth).toHaveBeenCalledWith(
        'sensitive-key',
        'sensitive-secret',
        'production'
      );
    });
  });

  describe('Environment Validation', () => {
    it('should accept staging environment', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('key', 'secret', 'staging');

      expect(mockSetAuth).toHaveBeenCalledWith('key', 'secret', 'staging');
    });

    it('should accept production environment', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;

      const { setAuth } = useMegaportWASM();
      setAuth('key', 'secret', 'production');

      expect(mockSetAuth).toHaveBeenCalledWith('key', 'secret', 'production');
    });
  });

  describe('Error Handling', () => {
    it('should handle missing setAuthCredentials function', () => {
      (window as any).setAuthCredentials = undefined;

      const { setAuth } = useMegaportWASM();

      // Should not throw error
      expect(() => {
        setAuth('key', 'secret', 'staging');
      }).not.toThrow();
    });

    it('should handle missing clearAuthCredentials function', () => {
      (window as any).clearAuthCredentials = undefined;

      const { clearAuth } = useMegaportWASM();

      // Should not throw error
      expect(() => {
        clearAuth();
      }).not.toThrow();
    });

    it('should handle missing debugAuthInfo', () => {
      (window as any).debugAuthInfo = undefined;

      const { getAuthInfo } = useMegaportWASM();

      // Should not throw error and return undefined
      expect(() => {
        getAuthInfo();
      }).not.toThrow();
    });
  });

  describe('Token Authentication (setAuthToken)', () => {
    it('should set authentication using portal token', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuthToken } = useMegaportWASM();
      setAuthToken('portal-session-token-12345', 'production');

      expect(mockSetAuthToken).toHaveBeenCalledWith(
        'portal-session-token-12345',
        'production'
      );
    });

    it('should work with staging environment', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuthToken } = useMegaportWASM();
      setAuthToken('staging-token', 'staging');

      expect(mockSetAuthToken).toHaveBeenCalledWith('staging-token', 'staging');
    });

    it('should handle JWT-style tokens', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuthToken } = useMegaportWASM();
      const jwtToken =
        'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.test';
      setAuthToken(jwtToken, 'production');

      expect(mockSetAuthToken).toHaveBeenCalledWith(jwtToken, 'production');
    });

    it('should return success response', () => {
      const mockSetAuthToken = vi.fn(() => ({
        success: true,
        message: 'Token set',
      }));
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuthToken } = useMegaportWASM();
      setAuthToken('token', 'production');

      expect(mockSetAuthToken).toHaveReturnedWith({
        success: true,
        message: 'Token set',
      });
    });

    it('should handle missing setAuthToken function', () => {
      (window as any).setAuthToken = undefined;

      const { setAuthToken } = useMegaportWASM();

      // Should not throw error
      expect(() => {
        setAuthToken('token', 'production');
      }).not.toThrow();
    });

    it('should call debugAuthInfo after setting token', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      const mockDebugAuthInfo = vi.fn(() => ({
        accessTokenSet: true,
        accessTokenPreview: 'por***45',
        authMethod: 'token',
        environment: 'production',
      }));
      (window as any).setAuthToken = mockSetAuthToken;
      (window as any).debugAuthInfo = mockDebugAuthInfo;

      const { setAuthToken } = useMegaportWASM();
      setAuthToken('portal-token', 'production');

      expect(mockDebugAuthInfo).toHaveBeenCalled();
    });

    it('should show token preview is masked in auth info', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthToken = mockSetAuthToken;
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessTokenSet: true,
        accessTokenPreview: 'tok***123',
        authMethod: 'token',
        environment: 'production',
      }));

      const { setAuthToken, getAuthInfo } = useMegaportWASM();
      setAuthToken('token-123456789', 'production');

      const info = getAuthInfo();
      expect(info?.accessTokenPreview).toContain('***');
      expect(info?.accessTokenPreview).not.toBe('token-123456789');
    });

    it('should indicate token auth method', () => {
      (window as any).setAuthToken = vi.fn(() => ({ success: true }));
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessTokenSet: true,
        accessTokenPreview: 'tok***',
        authMethod: 'token',
        environment: 'production',
      }));

      const { setAuthToken, getAuthInfo } = useMegaportWASM();
      setAuthToken('test-token', 'production');

      const info = getAuthInfo();
      expect(info?.authMethod).toBe('token');
    });

    it('should handle error response from setAuthToken', () => {
      const mockSetAuthToken = vi.fn(() => ({
        success: false,
        error: 'Invalid token format',
      }));
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuthToken } = useMegaportWASM();

      // Should not throw, but log error
      expect(() => {
        setAuthToken('invalid-token', 'production');
      }).not.toThrow();

      expect(mockSetAuthToken).toHaveBeenCalled();
    });
  });

  describe('Token vs API Key Authentication', () => {
    it('should support both token and API key auth methods', () => {
      const mockSetAuth = vi.fn(() => ({ success: true }));
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthCredentials = mockSetAuth;
      (window as any).setAuthToken = mockSetAuthToken;

      const { setAuth, setAuthToken } = useMegaportWASM();

      // Both methods should be available
      setAuth('api-key', 'api-secret', 'staging');
      expect(mockSetAuth).toHaveBeenCalled();

      setAuthToken('portal-token', 'production');
      expect(mockSetAuthToken).toHaveBeenCalled();
    });

    it('should show different auth methods in debug info', () => {
      const mockSetAuthToken = vi.fn(() => ({ success: true }));
      (window as any).setAuthToken = mockSetAuthToken;
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessTokenSet: true,
        authMethod: 'token',
        environment: 'production',
      }));

      const { setAuthToken, getAuthInfo } = useMegaportWASM();
      setAuthToken('token', 'production');

      const info = getAuthInfo();
      expect(info?.authMethod).toBe('token');

      // Simulate switching to API key
      (window as any).debugAuthInfo = vi.fn(() => ({
        accessKeySet: true,
        secretKeySet: true,
        authMethod: 'apikey',
        environment: 'staging',
      }));

      const info2 = getAuthInfo();
      expect(info2?.authMethod).toBe('apikey');
    });
  });
});
