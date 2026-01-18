<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { type ProjectDetail, ProjectService } from '@/services/project.service';
import { Button } from '@/components/ui/button';
import { TabNav } from '@/components/ui/tab-nav';
import { ArrowLeft, Loader2 } from 'lucide-vue-next';
import ProjectConfig from '@/components/project/ProjectConfig.vue';
import ProjectSettings from '@/components/project/ProjectSettings.vue';
import ProjectProviders from '@/components/project/ProjectProviders.vue';
import ProjectAccess from '@/components/project/ProjectAccess.vue';
import ProjectFiles from '@/components/project/ProjectFiles.vue';
import { EncryptionService } from '@/services/encryption.service';
import { TeamService } from '@/services/team.service';
import { useVaultStore } from '@/stores/vault';
import { useOrganizationStore } from '@/stores/organization';

const route = useRoute();
const router = useRouter();
const projectId = route.params.id as string;

const vaultStore = useVaultStore();
const orgStore = useOrganizationStore();

const project = ref<ProjectDetail | null>(null);
const isLoading = ref(false);
const error = ref('');
const activeTab = ref('config');

const tabs = [
    { key: 'config', label: 'Config' },
    { key: 'files', label: 'Files' },
    { key: 'access', label: 'Access' },
    { key: 'settings', label: 'Settings' },
    { key: 'providers', label: 'External providers' }
];

// Decryption state
const decryptedKey = ref('');
const decryptedTeamKey = ref('');
const isDecrypting = ref(false);
const decryptionError = ref('');

async function loadProject() {
    isLoading.value = true;
    error.value = '';

    try {
        project.value = await ProjectService.getProject(projectId);
    } catch (err: any) {
        error.value = 'Failed to load project: ' + err.toString();
    } finally {
        isLoading.value = false;
    }

    if (!project.value) {
        return;
    }

    await decryptProjectKey();
}

async function decryptProjectKey() {
    if (!project.value) return;

    isDecrypting.value = true;
    decryptionError.value = '';

    try {
        if (!vaultStore.privateKey) {
            throw new Error('Vault is locked. Please unlock your vault first.');
        }

        let teamKey = '';

        // Strategy 1: User is a team member - use encryptedTeamKey
        if (project.value.encryptedTeamKey) {
            console.log('Decrypting team key via user\'s encrypted team key');
            teamKey = await EncryptionService.decryptKey(
                vaultStore.privateKey,
                project.value.encryptedTeamKey
            );
        }

        // Strategy 2: User is org owner/admin without team membership
        // Need to fetch team info and decrypt via org key
        if (!teamKey && project.value.teamId && project.value.organizationId) {
            console.log('Attempting decryption via organization key (org owner/admin path)');

            // Get org key
            const orgKey = await orgStore.unlockOrganization(project.value.organizationId);
            if (!orgKey) {
                throw new Error('Unable to access organization key. You may not have sufficient permissions.');
            }

            // Fetch team to get team.encryptedKey (encrypted with org key)
            const teams = await TeamService.getTeams(project.value.organizationId);
            const team = teams.find((t) => t.id === project.value?.teamId);

            if (!team || !team.encryptedKey) {
                throw new Error('Team key not found. Unable to decrypt project.');
            }

            // Decrypt team key using org key (symmetric AES)
            teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);
        }

        if (!teamKey) {
            throw new Error('Unable to obtain team key for decryption.');
        }

        // Store team key for later use (e.g., deriving keys for old versions)
        decryptedTeamKey.value = teamKey;

        // Now decrypt the project key using the team key
        console.log('Decrypting project key with team key');
        decryptedKey.value = await EncryptionService.decryptValue(
            teamKey,
            project.value.encryptedProjectKey
        );
    } catch (e: any) {
        console.error('Decryption failed', e);
        decryptionError.value = 'Failed to unlock project: ' + (e.message || 'Unknown error');
    } finally {
        isDecrypting.value = false;
    }
}

function onProjectUpdated(updatedProject: ProjectDetail) {
    project.value = updatedProject;
}

loadProject();
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8 min-w-0 overflow-hidden">
        <div class="flex items-center gap-4">
            <Button
                variant="ghost"
                class="-ml-2 px-2 text-muted-foreground hover:text-foreground"
                @click="router.push('/')"
            >
                <ArrowLeft class="w-4 h-4 mr-2" />
                Back
            </Button>
        </div>

        <div v-if="isLoading" class="flex flex-col items-center py-12 text-muted-foreground">
            <Loader2 class="h-8 w-8 animate-spin mb-4" />
            <p>Loading project...</p>
        </div>

        <div v-else-if="error" class="bg-destructive/15 text-destructive p-4 rounded-md">
            {{ error }}
        </div>

        <div v-else-if="project" class="space-y-6 min-w-0">
            <!-- Header -->
            <div class="flex items-center justify-between">
                <div class="flex flex-col gap-1">
                    <div class="flex items-center gap-4">
                        <h1 class="text-3xl font-bold tracking-tight">{{ project.name }}</h1>
                    </div>
                    <div class="flex gap-4 text-sm text-muted-foreground">
                        <span class="font-mono">ID: {{ projectId }}</span>
                        <span v-if="project.teamName">Team: {{ project.teamName }}</span>
                    </div>
                </div>
            </div>

            <div v-if="decryptionError" class="bg-destructive/15 text-destructive p-4 rounded-md">
                {{ decryptionError }}
            </div>

            <div v-if="isDecrypting" class="text-muted-foreground text-sm">
                Decrypting project...
            </div>

            <TabNav v-model="activeTab" :tabs="tabs" />

            <div v-show="activeTab === 'config'" class="space-y-6 min-w-0">
                <ProjectConfig
                    :project="project"
                    :decrypted-key="decryptedKey"
                />
            </div>

            <div v-show="activeTab === 'files'" class="space-y-6">
                <ProjectFiles :project="project" :project-key="decryptedKey" />
            </div>

            <div v-show="activeTab === 'access'" class="space-y-6">
                <ProjectAccess :project="project" :decrypted-key="decryptedKey" />
            </div>

            <div v-show="activeTab === 'settings'" class="space-y-6">
                <ProjectSettings
                    :project="project"
                    :decrypted-key="decryptedKey"
                    :team-key="decryptedTeamKey"
                    @project-updated="onProjectUpdated"
                    @rotated="loadProject"
                />
            </div>

            <div v-show="activeTab === 'providers'" class="space-y-6">
                <ProjectProviders :project="project" :decrypted-key="decryptedKey" />
            </div>
        </div>
    </div>
</template>
