import { api } from './api';

export interface PendingRotation {
    id: string;
    projectId: string;
    initiatedBy: string;
    newVersion: number;
    status: string;
    requiredApprovals: number;
    expiresAt: string;
    createdAt: string;
    initiator: {
        id: string;
        name: string;
        email: string;
    };
    approvals: RotationApproval[];
    // Snapshot data for verification
    encryptedConfigsSnapshot: string;
    teamEncryptedKeys: string;
}

export interface RotationApproval {
    id: string;
    rotationId: string;
    userId: string;
    approved: boolean;
    verifiedDecryption: boolean;
    comment?: string;
    createdAt: string;
    user: {
        id: string;
        name: string;
        email: string;
    };
}

export interface InitiateRotationRequest {
    teamEncryptedKeys: {
        teamId: string;
        encryptedProjectKey: string;
    }[];
    reEncryptedConfigItems: {
        id: string;
        value: string;
    }[];
    reEncryptedFileFEKs?: {
        id: string;
        encryptedFek: string;
    }[];
}

export interface InitiateRotationResponse {
    message: string;
    rotationId?: string;
    newVersion: number;
    requiredApprovals?: number;
    expiresAt?: string;
    committed: boolean;
}

export interface PendingRotationWithProject extends PendingRotation {
    project: {
        id: string;
        name: string;
        organizationId: string;
    };
}

export class KeyRotationService {
    static async getPendingRotation(projectId: string): Promise<{ pending: PendingRotation | null; staleRotationExists?: boolean }> {
        const data = await api.get<{ pending?: PendingRotation; staleRotationExists?: boolean }>(`/projects/${projectId}/rotation`);
        return {
            pending: data.pending || null,
            staleRotationExists: data.staleRotationExists
        };
    }

    static async initiateRotation(projectId: string, request: InitiateRotationRequest): Promise<InitiateRotationResponse> {
        const response = await api.fetch(`/projects/${projectId}/rotation`, {
            method: 'POST',
            body: JSON.stringify(request)
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to initiate rotation');
        }
        return response.json();
    }

    static async approveRotation(
        projectId: string,
        rotationId: string,
        verifiedDecryption: boolean = false
    ): Promise<{ message: string; committed: boolean; newVersion?: number }> {
        const response = await api.fetch(`/projects/${projectId}/rotation/${rotationId}/approve`, {
            method: 'POST',
            body: JSON.stringify({ verifiedDecryption })
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to approve rotation');
        }
        return response.json();
    }

    static async rejectRotation(projectId: string, rotationId: string, comment?: string): Promise<void> {
        const response = await api.fetch(`/projects/${projectId}/rotation/${rotationId}/reject`, {
            method: 'POST',
            body: JSON.stringify({ comment })
        });
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to reject rotation');
        }
    }

    static async cancelRotation(projectId: string, rotationId: string): Promise<void> {
        await api.delete(`/projects/${projectId}/rotation/${rotationId}`);
    }

    /**
     * Get all pending rotations for the current user that need their attention
     * Used for dashboard notifications
     */
    static async getUserPendingRotations(): Promise<PendingRotationWithProject[]> {
        const data = await api.get<{ pendingRotations?: PendingRotationWithProject[] }>('/pending-rotations');
        return data.pendingRotations || [];
    }
}
