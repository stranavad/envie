import { useAuthStore } from '../stores/auth';
import { config } from '@/config';
import { toast } from '@/lib/toast';

export class ApiError extends Error {
    constructor(
        message: string,
        public status: number,
        public body?: unknown
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

async function parseErrorResponse(response: Response): Promise<string> {
    try {
        const body = await response.json();
        return body.error || body.message || response.statusText;
    } catch {
        return response.statusText;
    }
}

async function handleErrorResponse(response: Response): Promise<never> {
    const message = await parseErrorResponse(response);
    toast.error(message);
    throw new ApiError(message, response.status);
}

interface RequestOptions extends Omit<RequestInit, 'headers' | 'body'> {
    headers?: Record<string, string>;
    body?: BodyInit | null;
}

/**
 * API client with automatic token refresh
 */
export const api = {
    /**
     * Make an authenticated request with automatic token refresh
     */
    async fetch(endpoint: string, options: RequestOptions = {}): Promise<Response> {
        const authStore = useAuthStore();

        // Get a valid token (refreshes if needed)
        const token = await authStore.getValidToken();

        if (!token) {
            throw new ApiError('Not authenticated', 401);
        }

        const isFormData = options.body instanceof FormData;

        const headers: Record<string, string> = {
            ...options.headers,
            'Authorization': `Bearer ${token}`
        };

        // Only set Content-Type for non-FormData requests
        // Browser sets the correct Content-Type with boundary for FormData
        if (!isFormData) {
            headers['Content-Type'] = 'application/json';
        }

        const response = await fetch(`${config.backendUrl}${endpoint}`, {
            ...options,
            headers
        });

        // If we get 401, try one more refresh and retry
        if (response.status === 401) {
            const refreshSuccess = await authStore.refreshAccessToken();
            if (refreshSuccess) {
                const newToken = authStore.accessToken;
                headers['Authorization'] = `Bearer ${newToken}`;

                return fetch(`${config.backendUrl}${endpoint}`, {
                    ...options,
                    headers
                });
            } else {
                // Refresh failed, clear auth
                await authStore.clearAuth();
                throw new ApiError('Session expired', 401);
            }
        }

        return response;
    },

    async get<T>(endpoint: string): Promise<T> {
        const response = await this.fetch(endpoint);
        if (!response.ok) {
            await handleErrorResponse(response);
        }
        return response.json();
    },

    async post<T>(endpoint: string, data?: unknown): Promise<T> {
        const response = await this.fetch(endpoint, {
            method: 'POST',
            body: data ? JSON.stringify(data) : undefined
        });
        if (!response.ok) {
            await handleErrorResponse(response);
        }
        return response.json();
    },

    async put<T>(endpoint: string, data?: unknown): Promise<T> {
        const response = await this.fetch(endpoint, {
            method: 'PUT',
            body: data ? JSON.stringify(data) : undefined
        });
        if (!response.ok) {
            await handleErrorResponse(response);
        }
        return response.json();
    },

    async delete(endpoint: string): Promise<void> {
        const response = await this.fetch(endpoint, {
            method: 'DELETE'
        });
        if (!response.ok) {
            await handleErrorResponse(response);
        }
    },

    async postFormData<T>(endpoint: string, formData: FormData): Promise<T> {
        const response = await this.fetch(endpoint, {
            method: 'POST',
            body: formData,
            headers: {} // Let browser set Content-Type with boundary
        });
        if (!response.ok) {
            await handleErrorResponse(response);
        }
        return response.json();
    }
};
