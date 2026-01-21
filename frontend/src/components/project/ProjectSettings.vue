<script setup lang="ts">
import {ref, watch, onMounted} from 'vue';
import {useRouter} from 'vue-router';
import {type ProjectDetail, type ConfigItem, ProjectService} from '@/services/project.service';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Card, CardContent, CardDescription, CardHeader, CardTitle,} from '@/components/ui/card';
import KeyRotation from './KeyRotation.vue';
import PushPreviewDialog from './dialogs/PushPreviewDialog.vue';
import {FileMappingService, type FileMapping, type SyncStatus} from '@/services/file-mapping.service';
import {open} from '@tauri-apps/plugin-dialog';
import {FileText, Link2, Unlink, Download, Upload, AlertCircle, CheckCircle2, RefreshCw, Loader2} from 'lucide-vue-next';
import {useConfigEncryption} from '@/composables/useConfigEncryption';
import {useFileSync} from '@/composables/useFileSync';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
    teamKey: string;
}>();

const emit = defineEmits<{
    (e: 'projectUpdated', project: ProjectDetail): void;
    (e: 'rotated'): void;
    (e: 'reviewLocalChanges', items: { name: string; value: string }[]): void;
}>();

const router = useRouter();
const { fetchAndDecryptConfig } = useConfigEncryption();
const { pullToLocal, pushToRemote, readLocalItems } = useFileSync();
const error = ref('');
const success = ref('');

// General Settings State
const editName = ref(props.project.name);
const isSaving = ref(false);

// File Mapping State
const fileMapping = ref<FileMapping | null>(null);
const syncStatus = ref<SyncStatus>('not_linked');
const isLinking = ref(false);
const isPulling = ref(false);
const isPushing = ref(false);
const fileMappingError = ref('');
const fileMappingSuccess = ref('');

// Push Preview Dialog State
const showPushPreview = ref(false);
const pushPreviewLocalItems = ref<{ name: string; value: string }[]>([]);
const pushPreviewRemoteItems = ref<ConfigItem[]>([]);

watch(() => props.project, (newVal) => {
    editName.value = newVal.name;
    loadFileMapping();
});

onMounted(() => {
    loadFileMapping();
});

async function loadFileMapping() {
    try {
        const mapping = await FileMappingService.getMapping(props.project.id);
        fileMapping.value = mapping;

        if (mapping && props.project.configChecksum) {
            const result = await FileMappingService.checkSyncStatus(
                props.project.id,
                props.project.configChecksum
            );
            syncStatus.value = result.status;
        } else if (mapping) {
            syncStatus.value = 'synced'; // No remote checksum yet
        } else {
            syncStatus.value = 'not_linked';
        }
    } catch (e) {
        console.error('Failed to load file mapping', e);
    }
}

async function handleLinkFile() {
    isLinking.value = true;
    fileMappingError.value = '';
    fileMappingSuccess.value = '';

    try {
        const selected = await open({
            multiple: false,
        });

        if (selected && typeof selected === 'string') {
            const mapping = await FileMappingService.linkFile(
                props.project.id,
                selected,
                props.project.configChecksum || ''
            );
            fileMapping.value = mapping;
            syncStatus.value = 'synced';
            fileMappingSuccess.value = 'File linked successfully.';
        }
    } catch (e: any) {
        fileMappingError.value = 'Failed to link file: ' + e.toString();
    } finally {
        isLinking.value = false;
    }
}

async function handleUnlinkFile() {
    fileMappingError.value = '';
    fileMappingSuccess.value = '';

    try {
        await FileMappingService.unlinkFile(props.project.id);
        fileMapping.value = null;
        syncStatus.value = 'not_linked';
        fileMappingSuccess.value = 'File unlinked.';
    } catch (e: any) {
        fileMappingError.value = 'Failed to unlink file: ' + e.toString();
    }
}

async function handlePull() {
    if (!fileMapping.value) return;

    isPulling.value = true;
    fileMappingError.value = '';
    fileMappingSuccess.value = '';

    try {
        await pullToLocal(props.project.id, props.decryptedKey, props.project.configChecksum || '');
        syncStatus.value = 'synced';
        fileMappingSuccess.value = 'Local file updated from remote.';
        await loadFileMapping();
    } catch (e: any) {
        fileMappingError.value = 'Pull failed: ' + e.toString();
    } finally {
        isPulling.value = false;
    }
}

async function handlePushClick() {
    if (!fileMapping.value) return;

    fileMappingError.value = '';
    fileMappingSuccess.value = '';

    try {
        // Read local file and fetch remote config for comparison
        const [localItems, decryptedConfigs] = await Promise.all([
            readLocalItems(props.project.id),
            fetchAndDecryptConfig(props.project.id, props.decryptedKey),
        ]);

        // Open preview dialog
        pushPreviewLocalItems.value = localItems;
        pushPreviewRemoteItems.value = decryptedConfigs;
        showPushPreview.value = true;
    } catch (e: any) {
        fileMappingError.value = 'Failed to load changes: ' + e.toString();
    }
}

async function handleDirectPush() {
    if (!fileMapping.value) return;

    isPushing.value = true;
    fileMappingError.value = '';
    fileMappingSuccess.value = '';

    try {
        const { updatedProject } = await pushToRemote(
            props.project.id,
            props.decryptedKey,
            pushPreviewLocalItems.value,
            pushPreviewRemoteItems.value
        );

        showPushPreview.value = false;
        syncStatus.value = 'synced';
        fileMappingSuccess.value = 'Changes pushed successfully.';

        emit('projectUpdated', updatedProject);
        await loadFileMapping();
    } catch (e: any) {
        fileMappingError.value = 'Push failed: ' + e.toString();
    } finally {
        isPushing.value = false;
    }
}

function handleReviewChanges() {
    // Emit to parent to switch to config tab in sync mode
    emit('reviewLocalChanges', pushPreviewLocalItems.value);
}

function getSyncStatusDisplay(): { text: string; variant: 'success' | 'warning' | 'error' | 'muted' } {
    switch (syncStatus.value) {
        case 'synced':
            return { text: 'Synced', variant: 'success' };
        case 'local_changed':
            return { text: 'Local file changed', variant: 'warning' };
        case 'remote_changed':
            return { text: 'Remote config changed', variant: 'warning' };
        case 'both_changed':
            return { text: 'Both changed', variant: 'error' };
        case 'file_missing':
            return { text: 'File missing', variant: 'error' };
        default:
            return { text: 'Not linked', variant: 'muted' };
    }
}

async function handleUpdateName() {
    if (!editName.value || editName.value === props.project.name) return;

    isSaving.value = true;
    error.value = '';
    success.value = '';
    
    try {
        await ProjectService.updateProject(props.project.id, editName.value);
        // Create updated project object to emit
        const updatedProject = { ...props.project, name: editName.value };
        emit('projectUpdated', updatedProject);
        success.value = "Project name updated.";
    } catch (err: any) {
        error.value = "Failed to update project: " + err.toString();
    } finally {
        isSaving.value = false;
    }
}

async function handleDelete() {
    try {
        await ProjectService.deleteProject(props.project.id);
        await router.push('/');
    } catch (err: any) {
        error.value = "Failed to delete: " + err.toString();
    }
}
</script>

<template>
    <div class="space-y-6">
        <!-- GENERAL SETTINGS -->
        <Card>
            <CardHeader>
                <CardTitle>General Settings</CardTitle>
                <CardDescription>
                    Manage general project information.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <div class="grid gap-2">
                    <label class="text-sm font-medium">Project Name</label>
                    <div class="flex gap-2">
                        <Input v-model="editName" class="max-w-md" @keyup.enter="handleUpdateName"/>
                        <Button @click="handleUpdateName" :disabled="isSaving || editName === project.name">
                            {{ isSaving ? 'Saving...' : 'Save' }}
                        </Button>
                    </div>
                </div>
                <div v-if="success" class="text-sm text-green-600 font-medium">
                    {{ success }}
                </div>
            </CardContent>
        </Card>

        <!-- LOCAL .ENV FILE -->
        <Card>
            <CardHeader>
                <CardTitle class="flex items-center gap-2">
                    <FileText class="w-5 h-5" />
                    Local .env File
                </CardTitle>
                <CardDescription>
                    Link this project to a local .env file on your computer. Changes can be synced in either direction.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <!-- Not linked state -->
                <div v-if="!fileMapping" class="flex flex-col items-center gap-4 py-4">
                    <p class="text-sm text-muted-foreground text-center">
                        No local file linked. Link a .env file to enable sync between local and remote configs.
                    </p>
                    <Button @click="handleLinkFile" :disabled="isLinking">
                        <Link2 class="w-4 h-4 mr-2" />
                        {{ isLinking ? 'Selecting...' : 'Link .env File' }}
                    </Button>
                </div>

                <!-- Linked state -->
                <div v-else class="space-y-4">
                    <!-- File info -->
                    <div class="flex items-start justify-between p-3 bg-muted/50 rounded-lg">
                        <div class="flex items-start gap-3 min-w-0">
                            <FileText class="w-5 h-5 text-muted-foreground shrink-0 mt-0.5" />
                            <div class="min-w-0">
                                <p class="text-sm font-medium truncate" :title="fileMapping.filePath">
                                    {{ fileMapping.filePath }}
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    Linked {{ new Date(fileMapping.linkedAt).toLocaleDateString() }}
                                </p>
                            </div>
                        </div>
                        <Button variant="ghost" size="sm" @click="handleUnlinkFile">
                            <Unlink class="w-4 h-4" />
                        </Button>
                    </div>

                    <!-- Sync status -->
                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2">
                            <CheckCircle2 v-if="getSyncStatusDisplay().variant === 'success'" class="w-4 h-4 text-green-500" />
                            <AlertCircle v-else-if="getSyncStatusDisplay().variant === 'error'" class="w-4 h-4 text-destructive" />
                            <RefreshCw v-else-if="getSyncStatusDisplay().variant === 'warning'" class="w-4 h-4 text-orange-500" />
                            <span class="text-sm" :class="{
                                'text-green-600': getSyncStatusDisplay().variant === 'success',
                                'text-destructive': getSyncStatusDisplay().variant === 'error',
                                'text-orange-600': getSyncStatusDisplay().variant === 'warning',
                                'text-muted-foreground': getSyncStatusDisplay().variant === 'muted'
                            }">
                                {{ getSyncStatusDisplay().text }}
                            </span>
                        </div>
                        <Button variant="ghost" size="sm" @click="loadFileMapping">
                            <RefreshCw class="w-4 h-4" />
                        </Button>
                    </div>

                    <!-- Sync actions -->
                    <div class="flex items-center gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            @click="handlePull"
                            :disabled="isPulling || syncStatus === 'file_missing'"
                            class="flex-1"
                        >
                            <Loader2 v-if="isPulling" class="w-4 h-4 mr-2 animate-spin" />
                            <Download v-else class="w-4 h-4 mr-2" />
                            Pull (Remote → Local)
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            @click="handlePushClick"
                            :disabled="syncStatus === 'file_missing'"
                            class="flex-1"
                        >
                            <Upload class="w-4 h-4 mr-2" />
                            Push (Local → Remote)
                        </Button>
                    </div>

                    <p class="text-xs text-muted-foreground">
                        <strong>Pull:</strong> Overwrites local file with remote config.
                        <strong>Push:</strong> Push local changes to Envie.
                    </p>
                </div>

                <!-- Success/Error messages -->
                <div v-if="fileMappingSuccess" class="text-sm text-green-600 font-medium">
                    {{ fileMappingSuccess }}
                </div>
                <div v-if="fileMappingError" class="text-sm text-destructive">
                    {{ fileMappingError }}
                </div>
            </CardContent>
        </Card>

        <!-- KEY ROTATION -->
        <KeyRotation
            :project="project"
            :decrypted-key="decryptedKey"
            :team-key="teamKey"
            @rotated="emit('rotated')"
        />

        <div v-if="error" class="text-destructive text-sm">{{ error }}</div>

        <!-- DANGER ZONE -->
        <Card class="border-destructive/50">
            <CardHeader>
                <CardTitle class="text-destructive">Danger Zone</CardTitle>
                <CardDescription>
                    Irreversible actions for this project.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div class="flex items-center justify-between p-4 border border-destructive/20 rounded-md bg-destructive/5">
                    <div class="space-y-1">
                        <div class="font-medium text-destructive">Delete Project</div>
                        <div class="text-sm text-muted-foreground">Once you delete a project, there is no going back. Please be certain.</div>
                    </div>
                    <Button variant="destructive" @click="handleDelete">Delete Project</Button>
                </div>
            </CardContent>
        </Card>
    </div>

    <!-- Push Preview Dialog -->
    <PushPreviewDialog
        v-model:open="showPushPreview"
        :local-items="pushPreviewLocalItems"
        :remote-items="pushPreviewRemoteItems"
        :is-pushing="isPushing"
        @push="handleDirectPush"
        @review="handleReviewChanges"
    />
</template>
