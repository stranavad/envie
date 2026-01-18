import { defineStore } from 'pinia';
import { ref } from 'vue';
import { EncryptionService } from "@/services/encryption.service";
import { useVaultStore } from "@/stores/vault";
import { useAuthStore } from "@/stores/auth";
import { OrganizationService, type OrganizationListItem, type OrganizationDetail } from "@/services/organization.service";
import { TeamService, type TeamListItem } from "@/services/team.service";
import { ProjectService } from "@/services/project.service";

export type Organization = OrganizationListItem;

export const useOrganizationStore = defineStore('organization', () => {
    const vaultStore = useVaultStore();
    const authStore = useAuthStore();

    const organizations = ref<Organization[]>([]);
    const currentOrganization = ref<(OrganizationDetail & { teams?: any[] }) | null>(null);
    const orgMasterKeys = ref<Record<string, string>>({}); // orgId -> masterKey (decrypted)

    async function fetchOrganizations() {
        try {
            organizations.value = await OrganizationService.getOrganizations();
        } catch (e) {
            console.error("Failed to fetch organizations", e);
        }
    }

    async function createOrganization(name: string) {
        if (!authStore.user) throw new Error("User not authenticated");
        if (!vaultStore.publicKey) throw new Error("Vault not unlocked");

        // 1. Generate Org Master Key
        const orgMasterKey = EncryptionService.generateAesKey();

        // 2. Encrypt for Self (Asymmetric)
        const encryptedOrgKey = await EncryptionService.encryptKey(vaultStore.publicKey, orgMasterKey);

        // 3. Generate General Team Key
        const teamKey = EncryptionService.generateAesKey();

        // 4. Encrypt Team Key with Org Master Key (Symmetric)
        const generalTeamEncryptedKey = await EncryptionService.encryptValue(orgMasterKey, teamKey);

        // 5. Encrypt Team Key with Self Public Key (Asymmetric)
        const generalTeamUserEncryptedKey = await EncryptionService.encryptKey(vaultStore.publicKey, teamKey);

        const org = await OrganizationService.createOrganization({
            name,
            encryptedOrganizationKey: encryptedOrgKey,
            generalTeamEncryptedKey,
            generalTeamUserEncryptedKey
        });

        // Add to list and cache key
        organizations.value.push({ ...org, role: 'owner', projectCount: 0, memberCount: 1 });
        orgMasterKeys.value[org.id] = orgMasterKey;

        return org;
    }

    async function getOrganization(id: string) {
        try {
            const orgDetail = await OrganizationService.getOrganization(id);
            currentOrganization.value = orgDetail;
            return currentOrganization.value;
        } catch (e) {
            console.error("Failed to get organization", e);
            throw e;
        }
    }

    async function unlockOrganization(orgId: string) {
        // If we already have the key, return
        if (orgMasterKeys.value[orgId]) return orgMasterKeys.value[orgId];

        if (!currentOrganization.value || currentOrganization.value.id !== orgId) {
            await getOrganization(orgId); // Ensure loaded
        }

        const encryptedKey = currentOrganization.value?.encryptedOrganizationKey;
        if (!encryptedKey) {
            // Maybe I don't have access or key is null?
            // If key is null, maybe I am not an admin/owner?
            // But I should be able to see the org.
            // If I am just a member, I might not have the Org Master Key.
            return null;
        }

        if (!vaultStore.privateKey) throw new Error("Vault locked");

        try {
            const key = await EncryptionService.decryptKey(vaultStore.privateKey, encryptedKey);
            orgMasterKeys.value[orgId] = key;
            return key;
        } catch (e) {
            console.error("Failed to decrypt org key", e);
            throw e;
        }
    }

    async function createTeam(orgId: string, name: string) {
        // Ensure Org Key is available
        let orgKey = orgMasterKeys.value[orgId];
        if (!orgKey) {
            orgKey = await unlockOrganization(orgId) || "";
        }

        if (!orgKey) {
            throw new Error("Cannot create team: Organization Master Key missing or inaccessible.");
        }

        // 1. Generate Team Key
        const teamKey = EncryptionService.generateAesKey();

        // 2. Encrypt with Org Key (Symmetric)
        const encryptedKey = await EncryptionService.encryptValue(orgKey, teamKey);

        // 3. Encrypt with Self Public Key (Asymmetric)
        const userEncryptedKey = await EncryptionService.encryptKey(vaultStore.publicKey!, teamKey);

        const team = await TeamService.createTeam({
            name,
            organizationId: orgId,
            encryptedKey,
            userEncryptedKey
        });

        return team;
    }

    async function createProject(orgId: string, teamId: string, name: string) {
        // 1. Fetch Team to get encrypted key
        const teams = await TeamService.getTeams(orgId);
        const team = teams.find((t) => t.id === teamId);
        if (!team) throw new Error("Team not found");

        let teamKey = "";

        // Strategy A: Use User-Specific Team Key (Preferred, works for all members)
        if (team.userEncryptedKey) {
            if (!vaultStore.privateKey) throw new Error("Vault locked");
            try {
                teamKey = await EncryptionService.decryptKey(vaultStore.privateKey, team.userEncryptedKey);
            } catch (e) {
                console.error("Failed to decrypt team key from userEncryptedKey", e);
            }
        }

        // Strategy B: Use Org Master Key (Fallback, works for owners/admins if they have it unlocked)
        if (!teamKey) {
            let orgKey = orgMasterKeys.value[orgId];
            if (!orgKey) {
                orgKey = await unlockOrganization(orgId) || "";
            }
            if (orgKey && team.encryptedKey) {
                try {
                    teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);
                } catch (e) {
                    console.error("Failed to decrypt team key from Org Key", e);
                }
            }
        }

        if (!teamKey) {
            throw new Error("Cannot create project: Unable to decrypt Team Key. Ensure you are a member of this team or have organization access.");
        }

        // 3. Generate Project Key
        const projectKey = EncryptionService.generateAesKey();

        // 4. Encrypt Project Key with Team Key
        const encryptedProjectKey = await EncryptionService.encryptValue(teamKey, projectKey);

        const project = await ProjectService.createProject({
            name,
            organizationId: orgId,
            teamId,
            encryptedKey: encryptedProjectKey
        });

        return project;
    }

    async function fetchTeams(orgId: string): Promise<TeamListItem[]> {
        return TeamService.getTeams(orgId);
    }

    return {
        organizations,
        currentOrganization,
        fetchOrganizations,
        createOrganization,
        getOrganization,
        createTeam,
        createProject,
        fetchTeams,
        unlockOrganization
    };
});
