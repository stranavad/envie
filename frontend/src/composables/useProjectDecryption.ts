import { ref } from 'vue';
import { EncryptionService } from '@/services/encryption.service';
import { IdentityService } from '@/services/identity.service';
import { TeamService, type TeamListItem } from '@/services/team.service';
import { useOrganizationStore } from '@/stores/organization';

export interface ProjectKeyInfo {
    teamId?: string;
    organizationId: string;
    encryptedTeamKey?: string; // User's asymmetric-encrypted team key
    encryptedProjectKey: string; // Team key symmetric-encrypted project key
}

/**
 * Composable for decrypting project and team keys.
 * Centralizes the common decryption patterns used across the app.
 *
 * Decryption chain:
 * 1. Team Key (via user's encryptedTeamKey OR via org key → team.encryptedKey)
 * 2. Project Key (via team key → encryptedProjectKey)
 */
export function useProjectDecryption() {
    const orgStore = useOrganizationStore();

    const isDecrypting = ref(false);
    const decryptionError = ref('');

    /**
     * Decrypt team key using one of two strategies:
     * 1. User is a team member: decrypt encryptedTeamKey using Master Identity private key (asymmetric)
     * 2. User is org owner/admin: decrypt team.encryptedKey using org master key (symmetric)
     */
    async function decryptTeamKey(
        encryptedTeamKey: string | undefined,
        teamId: string | undefined,
        organizationId: string
    ): Promise<string> {
        const masterKeyPair = IdentityService.getMasterKeyPair();
        if (!masterKeyPair) {
            throw new Error('Master Identity not loaded. Please unlock your vault first.');
        }

        let teamKey = '';

        // Strategy 1: User is a team member - use encryptedTeamKey
        if (encryptedTeamKey) {
            console.log('Decrypting team key via user\'s encrypted team key');
            teamKey = await EncryptionService.decryptKey(
                masterKeyPair.privateKey,
                encryptedTeamKey
            );
        }

        // Strategy 2: User is org owner/admin without team membership
        // Need to fetch team info and decrypt via org key
        if (!teamKey && teamId && organizationId) {
            console.log('Attempting decryption via organization key (org owner/admin path)');

            const orgKey = await orgStore.unlockOrganization(organizationId);
            if (!orgKey) {
                throw new Error('Unable to access organization key. You may not have sufficient permissions.');
            }

            const teams = await TeamService.getTeams(organizationId);
            const team = teams.find((t) => t.id === teamId);

            if (!team || !team.encryptedKey) {
                throw new Error('Team key not found. Unable to decrypt project.');
            }

            // Decrypt team key using org key (symmetric AES)
            teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);
        }

        if (!teamKey) {
            throw new Error('Unable to obtain team key for decryption.');
        }

        return teamKey;
    }

    /**
     * Decrypt team key from a team object (for operations like adding team to project)
     */
    async function decryptTeamKeyFromTeam(
        team: TeamListItem,
        organizationId: string
    ): Promise<string> {
        const masterKeyPair = IdentityService.getMasterKeyPair();
        if (!masterKeyPair) {
            throw new Error('Master Identity not loaded');
        }

        let teamKey = '';

        // Try to decrypt team key via user's encrypted team key (asymmetric)
        if (team.userEncryptedKey) {
            teamKey = await EncryptionService.decryptKey(masterKeyPair.privateKey, team.userEncryptedKey);
        }

        // Fallback: Use org key (symmetric)
        if (!teamKey) {
            const orgKey = await orgStore.unlockOrganization(organizationId);
            if (orgKey && team.encryptedKey) {
                teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);
            }
        }

        if (!teamKey) {
            throw new Error('Unable to decrypt team key');
        }

        return teamKey;
    }

    /**
     * Decrypt project key using the team key (symmetric AES)
     */
    async function decryptProjectKey(
        teamKey: string,
        encryptedProjectKey: string
    ): Promise<string> {
        console.log('Decrypting project key with team key');
        return await EncryptionService.decryptValue(teamKey, encryptedProjectKey);
    }

    /**
     * Full decryption chain: get both team key and project key
     * Returns { teamKey, projectKey }
     */
    async function decryptProjectKeys(info: ProjectKeyInfo): Promise<{
        teamKey: string;
        projectKey: string;
    }> {
        isDecrypting.value = true;
        decryptionError.value = '';

        try {
            const teamKey = await decryptTeamKey(
                info.encryptedTeamKey,
                info.teamId,
                info.organizationId
            );

            const projectKey = await decryptProjectKey(teamKey, info.encryptedProjectKey);

            return { teamKey, projectKey };
        } catch (e: any) {
            console.error('Decryption failed', e);
            decryptionError.value = 'Failed to unlock project: ' + (e.message || 'Unknown error');
            throw e;
        } finally {
            isDecrypting.value = false;
        }
    }

    return {
        isDecrypting,
        decryptionError,
        decryptTeamKey,
        decryptTeamKeyFromTeam,
        decryptProjectKey,
        decryptProjectKeys,
    };
}
