import {defineStore} from 'pinia';
import {computed, ref} from 'vue';
import {openUrl} from '@tauri-apps/plugin-opener';
import {config} from '../config';
import {useVaultStore} from './vault';
import { AuthService, RotateMasterKeyRequest } from '@/services/auth.service';

interface User {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    githubId: number;
    googleId: string;
    publicKey: string | null;
    masterKeyVersion: number;
}

interface TokenResponse {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: User;
}

export const useAuthStore = defineStore('auth', () => {
    const accessToken = ref<string | null>(null);
    const tokenExpiresAt = ref<number | null>(null);
    const user = ref<User | null>(null);
    const isRefreshing = ref(false);
    const knownMasterKeyVersion = ref<number | null>(null);
    const masterKeyVersionMismatch = ref(false);

    // Pending refresh token to be saved to vault once vault is unlocked
    let pendingRefreshToken: string | null = null;

    // Queue of pending requests waiting for token refresh
    let refreshPromise: Promise<boolean> | null = null;

    // Callback for when master key version changes (key was rotated on another device)
    let onMasterKeyVersionChange: (() => void) | null = null;

    const isAuthenticated = computed(() => !!accessToken.value && !!user.value);

    // Check if token is expired or about to expire (within 5 minutes)
    const isTokenExpired = computed(() => {
        if (!tokenExpiresAt.value) return true;
        return Date.now() >= tokenExpiresAt.value - 5 * 60 * 1000;
    });

    // Legacy alias for backwards compatibility
    const token = computed(() => accessToken.value);

    async function login(provider: 'github' | 'google' = 'github') {
        const path = provider === 'google' ? '/auth/login/google' : '/auth/login';
        await openUrl(`${config.backendUrl}${path}?app=envie`);
    }

    /**
     * Exchange a linking code for access and refresh tokens
     * devicePublicKey is optional - will be registered after vault setup
     */
    // Store the last exchange error for the UI to display
    const lastExchangeError = ref<string | null>(null);

    async function exchangeLinkingCode(code: string, devicePublicKey?: string): Promise<boolean> {
        lastExchangeError.value = null;

        try {
            const body: Record<string, string> = { code: code.trim() };
            if (devicePublicKey) {
                body.devicePublicKey = devicePublicKey;
            }

            const response = await fetch(`${config.backendUrl}/auth/exchange`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(body)
            });

            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: `HTTP ${response.status}` }));
                console.error('Exchange failed:', error);
                lastExchangeError.value = error.error || `Server error: ${response.status}`;
                return false;
            }

            const data: TokenResponse = await response.json();

            // Store access token in memory
            accessToken.value = data.accessToken;
            tokenExpiresAt.value = Date.now() + data.expiresIn * 1000;
            user.value = data.user;

            // Try to store refresh token in Stronghold, or keep in memory if vault not ready
            const vaultStore = useVaultStore();
            if (vaultStore.status === 'unlocked') {
                await vaultStore.saveRefreshToken(data.refreshToken);
                pendingRefreshToken = null;
            } else {
                // Vault not ready yet - store in memory, will be persisted when vault unlocks
                pendingRefreshToken = data.refreshToken;
            }

            return true;
        } catch (e: any) {
            console.error("Failed to exchange linking code", e);
            lastExchangeError.value = e.message || 'Network error - could not reach server';
            return false;
        }
    }

    /**
     * Refresh the access token using the stored refresh token
     */
    async function refreshAccessToken(): Promise<boolean> {
        // If already refreshing, wait for that to complete
        if (refreshPromise) {
            return refreshPromise;
        }

        isRefreshing.value = true;

        refreshPromise = (async () => {
            try {
                const vaultStore = useVaultStore();

                // Try to get refresh token from vault, or use pending token
                let refreshToken = pendingRefreshToken;
                if (vaultStore.status === 'unlocked') {
                    refreshToken = await vaultStore.getRefreshToken() || pendingRefreshToken;
                }

                if (!refreshToken) {
                    console.error('No refresh token available');
                    return false;
                }

                const response = await fetch(`${config.backendUrl}/auth/refresh`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        refreshToken: refreshToken
                    })
                });

                if (!response.ok) {
                    // Refresh token invalid/expired - need to re-login
                    console.error('Refresh token invalid');
                    await clearAuth();
                    return false;
                }

                const data: TokenResponse = await response.json();

                // Update access token
                accessToken.value = data.accessToken;
                tokenExpiresAt.value = Date.now() + data.expiresIn * 1000;

                // Update user if returned
                if (data.user) {
                    user.value = data.user;
                }

                // Store new refresh token (token rotation)
                if (vaultStore.status === 'unlocked') {
                    await vaultStore.saveRefreshToken(data.refreshToken);
                    pendingRefreshToken = null;
                } else {
                    // Vault not ready - keep in memory
                    pendingRefreshToken = data.refreshToken;
                }

                return true;
            } catch (e) {
                console.error("Failed to refresh token", e);
                return false;
            } finally {
                isRefreshing.value = false;
                refreshPromise = null;
            }
        })();

        return refreshPromise;
    }

    /**
     * Get a valid access token, refreshing if necessary
     */
    async function getValidToken(): Promise<string | null> {
        if (!accessToken.value) return null;

        if (isTokenExpired.value) {
            const success = await refreshAccessToken();
            if (!success) return null;
        }

        return accessToken.value;
    }

    /**
     * Clear all auth state (used on session expiry/refresh failure)
     */
    async function clearAuth() {
        accessToken.value = null;
        tokenExpiresAt.value = null;
        // Don't clear user - keep it for "Welcome back" experience

        const vaultStore = useVaultStore();
        await vaultStore.clearRefreshToken();
    }

    /**
     * Logout - revoke tokens on backend and clear local state
     * Keeps vault intact so user doesn't have to re-approve device
     */
    async function logout() {
        try {
            if (accessToken.value) {
                // Try to revoke tokens on backend
                await fetch(`${config.backendUrl}/auth/logout`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${accessToken.value}`
                    }
                });
            }
        } catch (e) {
            console.error("Failed to logout on backend", e);
        }

        // Clear tokens but keep user info and vault
        accessToken.value = null;
        tokenExpiresAt.value = null;
        // Keep user.value - for "Welcome back" experience

        const vaultStore = useVaultStore();
        await vaultStore.clearRefreshToken();
        // Don't call vault.reset() - keep the vault intact
    }

    /**
     * Try to restore session using stored refresh token
     */
    async function tryRestoreSession(): Promise<boolean> {
        const vaultStore = useVaultStore();
        const refreshToken = await vaultStore.getRefreshToken();

        if (!refreshToken) {
            return false;
        }

        return await refreshAccessToken();
    }

    async function fetchUser() {
        const validToken = await getValidToken();
        if (!validToken) return;

        try {
            const response = await fetch(`${config.backendUrl}/me`, {
                headers: {
                    'Authorization': `Bearer ${validToken}`
                }
            });

            if (response.ok) {
                // Check master key version from header
                checkMasterKeyVersion(response);

                user.value = await response.json();
            } else if (response.status === 401) {
                // Token invalid even after refresh - logout
                await clearAuth();
            }
        } catch (e) {
            console.error("Failed to fetch user", e);
        }
    }

    /**
     * Check master key version from response header and trigger callback if changed
     */
    function checkMasterKeyVersion(response: Response) {
      const versionHeader = response.headers.get('X-Master-Key-Version');
      if (!versionHeader) {
        return
      }

      const serverVersion = parseInt(versionHeader, 10);

      if (isNaN(serverVersion)) {
        return
      }

      if (knownMasterKeyVersion.value !== null && serverVersion > knownMasterKeyVersion.value) {
          // Key was rotated on another device
          masterKeyVersionMismatch.value = true;
          if (onMasterKeyVersionChange) {
              onMasterKeyVersionChange();
          }
      }

      knownMasterKeyVersion.value = serverVersion;
    }

    /**
     * Set callback for when master key version changes (key rotated on another device)
     */
    function setOnMasterKeyVersionChange(callback: (() => void) | null) {
        onMasterKeyVersionChange = callback;
    }

    /**
     * Clear the master key version mismatch flag (after user has handled it)
     */
    function clearMasterKeyVersionMismatch() {
        masterKeyVersionMismatch.value = false;
    }

    /**
     * Persist any pending refresh token to vault (call after vault is unlocked)
     */
    async function persistPendingRefreshToken() {
        if (!pendingRefreshToken) return;

        const vaultStore = useVaultStore();
        if (vaultStore.status !== 'unlocked') return;

        try {
            await vaultStore.saveRefreshToken(pendingRefreshToken);
            pendingRefreshToken = null;
        } catch (e) {
            console.error("Failed to persist refresh token to vault", e);
        }
    }


    async function setMasterPublicKey(masterPublicKey: string) {
        const validToken = await getValidToken();
        if (!validToken) return;

        try {
            const responseData = await AuthService.updatePublicKey(validToken, masterPublicKey)

            if (user.value) {
                user.value.publicKey = responseData.publicKey;
            }
        } catch (e) {
            console.error("Failed to set master public key", e);
        }
    }

    /**
     * Rotate the master key - updates public key, identity keys, and team keys atomically
     */
    async function rotateMasterKey(request: RotateMasterKeyRequest): Promise<boolean> {
        const validToken = await getValidToken();
        if (!validToken) {
            throw new Error('Not authenticated');
        }

        try {
            const response = await AuthService.rotateMasterKey(validToken, request);

            // Update local state
            if (user.value) {
                user.value.publicKey = response.publicKey;
                user.value.masterKeyVersion = response.masterKeyVersion;
            }
            knownMasterKeyVersion.value = response.masterKeyVersion;
            masterKeyVersionMismatch.value = false;

            return true;
        } catch (e) {
            console.error("Failed to rotate master key", e);
            throw e;
        }
    }


    // Legacy method - now exchanges linking code instead of verifying JWT directly
    async function verifyToken(linkingCode: string, devicePublicKey?: string) {
        if (!linkingCode) return false;

        // If it looks like a linking code (XXXX-XXXX-XXXX format), exchange it
        if (linkingCode.includes('-') && devicePublicKey) {
            return await exchangeLinkingCode(linkingCode, devicePublicKey);
        }

        // Legacy: try as direct JWT token (for backwards compatibility during transition)
        try {
            const response = await fetch(`${config.backendUrl}/me`, {
                headers: {
                    'Authorization': `Bearer ${linkingCode}`
                }
            });

            if (response.ok) {
                const userData = await response.json();
                accessToken.value = linkingCode;
                tokenExpiresAt.value = Date.now() + 60 * 60 * 1000; // Assume 1 hour
                user.value = userData;
                return true;
            }
        } catch (e) {
            console.error("Token verification failed", e);
        }
        return false;
    }

    return {
        // State
        token, // Legacy alias
        accessToken,
        tokenExpiresAt,
        user,
        isRefreshing,
        knownMasterKeyVersion,
        masterKeyVersionMismatch,
        lastExchangeError,

        // Computed
        isAuthenticated,
        isTokenExpired,

        // Actions
        login,
        logout,
        fetchUser,
        verifyToken,
        setMasterPublicKey,
        rotateMasterKey,
        exchangeLinkingCode,
        refreshAccessToken,
        getValidToken,
        tryRestoreSession,
        clearAuth,
        persistPendingRefreshToken,
        setOnMasterKeyVersionChange,
        clearMasterKeyVersionMismatch,
        checkMasterKeyVersion
    };
}, {
    persist: {
        pick: ['user'] // Only persist user info, not tokens (tokens in Stronghold)
    }
});
