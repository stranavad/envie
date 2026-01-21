<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { type ProjectDetail, ProjectService } from '@/services/project.service';
import { Button } from '@/components/ui/button';
import { TabNav } from '@/components/ui/tab-nav';
import { ArrowLeft, Loader2, Upload, Download, AlertTriangle } from 'lucide-vue-next';
import ProjectConfig from '@/components/project/ProjectConfig.vue';
import ProjectSettings from '@/components/project/ProjectSettings.vue';
import ProjectProviders from '@/components/project/ProjectProviders.vue';
import ProjectAccess from '@/components/project/ProjectAccess.vue';
import ProjectFiles from '@/components/project/ProjectFiles.vue';
import { FileMappingService, type SyncStatus } from '@/services/file-mapping.service';
import { useProjectDecryption } from '@/composables/useProjectDecryption';
import { useConfigEncryption } from '@/composables/useConfigEncryption';

const route = useRoute();
const router = useRouter();
const projectId = route.params.id as string;

// Composables
const { isDecrypting, decryptionError, decryptProjectKeys } = useProjectDecryption();
const { fetchAndDecryptConfig } = useConfigEncryption();

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

// Local file sync mode state
const localImportItems = ref<{ name: string; value: string }[] | null>(null);
const syncMode = ref(false);

// Sync status state
const syncStatus = ref<SyncStatus>('not_linked');
const isPulling = ref(false);
const syncError = ref('');

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

    await decryptProjectKeyData();
    await loadSyncStatus();
}

async function loadSyncStatus() {
    if (!project.value) return;

    try {
        const result = await FileMappingService.checkSyncStatus(
            projectId,
            project.value.configChecksum || ''
        );
        syncStatus.value = result.status;
    } catch (e) {
        console.error('Failed to load sync status', e);
    }
}

async function decryptProjectKeyData() {
    if (!project.value) return;

    try {
        const { teamKey, projectKey } = await decryptProjectKeys({
            teamId: project.value.teamId,
            organizationId: project.value.organizationId,
            encryptedTeamKey: project.value.encryptedTeamKey,
            encryptedProjectKey: project.value.encryptedProjectKey,
        });

        decryptedTeamKey.value = teamKey;
        decryptedKey.value = projectKey;
    } catch (e) {
        // Error is already logged and set in decryptionError by the composable
    }
}

function onProjectUpdated(updatedProject: ProjectDetail) {
    project.value = updatedProject;
}

function handleReviewLocalChanges(items: { name: string; value: string }[]) {
    localImportItems.value = items;
    syncMode.value = true;
    activeTab.value = 'config';
}

function handleSyncComplete() {
    localImportItems.value = null;
    syncMode.value = false;
    // Reload project to get updated checksum
    loadProject();
}

async function handlePull() {
    if (!project.value || !decryptedKey.value) return;

    isPulling.value = true;
    syncError.value = '';

    try {
        const mapping = await FileMappingService.getMapping(projectId);
        if (!mapping) {
            throw new Error('No file mapping found');
        }

        // Fetch and decrypt config items using composable
        const decryptedItems = await fetchAndDecryptConfig(projectId, decryptedKey.value);

        // Write to local file
        const localChecksum = await FileMappingService.writeToLocalFile(
            mapping.filePath,
            decryptedItems
        );

        // Update sync state
        await FileMappingService.updateSyncState(
            projectId,
            localChecksum,
            project.value.configChecksum || ''
        );

        syncStatus.value = 'synced';
    } catch (e: any) {
        console.error('Pull failed', e);
        syncError.value = 'Pull failed: ' + (e.message || e.toString());
    } finally {
        isPulling.value = false;
    }
}

async function handlePushReview() {
    if (!project.value) return;

    try {
        const mapping = await FileMappingService.getMapping(projectId);
        if (!mapping) {
            throw new Error('No file mapping found');
        }

        // Read local file and trigger review mode
        const localItems = await FileMappingService.readLocalFile(mapping.filePath);
        handleReviewLocalChanges(localItems);
    } catch (e: any) {
        console.error('Failed to load local changes', e);
        syncError.value = 'Failed to load local changes: ' + (e.message || e.toString());
    }
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

            <!-- Sync Status Banner -->
            <div
                v-if="syncStatus === 'local_changed' && !syncMode"
                class="flex items-center justify-between p-4 rounded-lg border bg-orange-500/10 border-orange-500/40"
            >
                <div class="flex items-center gap-3">
                    <Upload class="w-5 h-5 text-orange-400" />
                    <div>
                        <p class="font-medium text-orange-200">Local .env file has changed</p>
                        <p class="text-sm text-orange-300/70">Review and push your local changes to sync with Envie.</p>
                    </div>
                </div>
                <Button size="sm" variant="outline" @click="handlePushReview">
                    Review Changes
                </Button>
            </div>

            <div
                v-if="syncStatus === 'remote_changed'"
                class="flex items-center justify-between p-4 rounded-lg border bg-blue-500/10 border-blue-500/40"
            >
                <div class="flex items-center gap-3">
                    <Download class="w-5 h-5 text-blue-400" />
                    <div>
                        <p class="font-medium text-blue-200">Remote config has changed</p>
                        <p class="text-sm text-blue-300/70">Pull the latest changes to update your local .env file.</p>
                    </div>
                </div>
                <Button size="sm" variant="outline" :disabled="isPulling" @click="handlePull">
                    <Loader2 v-if="isPulling" class="w-4 h-4 mr-2 animate-spin" />
                    <Download v-else class="w-4 h-4 mr-2" />
                    Pull Changes
                </Button>
            </div>

            <div
                v-if="syncStatus === 'both_changed'"
                class="flex items-center justify-between p-4 rounded-lg border bg-red-500/10 border-red-500/40"
            >
                <div class="flex items-center gap-3">
                    <AlertTriangle class="w-5 h-5 text-red-400" />
                    <div>
                        <p class="font-medium text-red-200">Sync conflict detected</p>
                        <p class="text-sm text-red-300/70">Both local and remote have changed. Go to Settings to resolve.</p>
                    </div>
                </div>
                <Button size="sm" variant="outline" @click="activeTab = 'settings'">
                    Go to Settings
                </Button>
            </div>

            <div v-if="syncError" class="bg-destructive/15 text-destructive p-4 rounded-md">
                {{ syncError }}
            </div>

            <TabNav v-model="activeTab" :tabs="tabs" />

            <div v-if="activeTab === 'config' && decryptedKey" class="space-y-6 min-w-0">
                <ProjectConfig
                    :project="project"
                    :decrypted-key="decryptedKey"
                    :local-import-items="localImportItems"
                    :sync-mode="syncMode"
                    @sync-complete="handleSyncComplete"
                />
            </div>

            <div v-if="activeTab === 'files' && decryptedKey" class="space-y-6">
                <ProjectFiles :project="project" :project-key="decryptedKey" />
            </div>

            <div v-if="activeTab === 'access' && decryptedKey" class="space-y-6">
                <ProjectAccess :project="project" :decrypted-key="decryptedKey" />
            </div>

            <div v-if="activeTab === 'settings' && decryptedKey" class="space-y-6">
                <ProjectSettings
                    :project="project"
                    :decrypted-key="decryptedKey"
                    :team-key="decryptedTeamKey"
                    @project-updated="onProjectUpdated"
                    @rotated="loadProject"
                    @review-local-changes="handleReviewLocalChanges"
                />
            </div>

            <div v-if="activeTab === 'providers' && decryptedKey" class="space-y-6">
                <ProjectProviders :project="project" :decrypted-key="decryptedKey" />
            </div>
        </div>
    </div>
</template>
