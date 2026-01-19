import { config } from '@/config';

export interface User {
    id: string;
    name: string;
    email: string;
    avatarUrl: string;
    githubId: number;
    publicKey: string | null; // null until first device setup
    masterKeyVersion: number; // for detecting key rotation on other devices
}

export interface TokenResponse {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: User;
}

export interface ExchangeCodeRequest {
    code: string;
    devicePublicKey?: string;
}

export interface RefreshTokenRequest {
    refreshToken: string;
}

export interface UpdatePublicKeyRequest {
    publicKey: string;
}

export interface RotateMasterKeyRequest {
    newPublicKey: string;
    identityKeys: Record<string, string>; // identityId -> encrypted master key
    teamKeys: Record<string, string>; // teamId -> encrypted team key
}

export interface RotateMasterKeyResponse {
    message: string;
    publicKey: string;
    masterKeyVersion: number;
    identitiesUpdated: number;
    teamsUpdated: number;
}

/**
 * Auth service for authentication endpoints.
 * Note: These endpoints don't use the standard api client because they
 * either don't require auth or handle tokens directly.
 */
export class AuthService {
    static async exchangeLinkingCode(request: ExchangeCodeRequest): Promise<TokenResponse> {
        const response = await fetch(`${config.backendUrl}/auth/exchange`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request),
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: response.statusText }));
            throw new Error(error.error || 'Exchange failed');
        }

        return response.json();
    }

    static async refreshToken(request: RefreshTokenRequest): Promise<TokenResponse> {
        const response = await fetch(`${config.backendUrl}/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request),
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: response.statusText }));
            throw new Error(error.error || 'Refresh failed');
        }

        return response.json();
    }

    static async logout(accessToken: string): Promise<void> {
        await fetch(`${config.backendUrl}/auth/logout`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${accessToken}` },
        });
    }

    static async getCurrentUser(accessToken: string): Promise<User> {
        const response = await fetch(`${config.backendUrl}/me`, {
            headers: { 'Authorization': `Bearer ${accessToken}` },
        });

        if (!response.ok) {
            throw new Error('Failed to fetch user');
        }

        return response.json();
    }

    static async updatePublicKey(accessToken: string, publicKey: string): Promise<{publicKey: string}> {
        const response = await fetch(`${config.backendUrl}/me/public-key`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${accessToken}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ publicKey }),
        });

        if (!response.ok) {
            throw new Error('Failed to update public key');
        }

        return response.json()
    }

    static async rotateMasterKey(
        accessToken: string,
        request: RotateMasterKeyRequest
    ): Promise<RotateMasterKeyResponse> {
        const response = await fetch(`${config.backendUrl}/me/rotate-master-key`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${accessToken}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: response.statusText }));
            throw new Error(error.error || 'Failed to rotate master key');
        }

        return response.json();
    }
}
