<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Loader2, Copy } from 'lucide-vue-next';
import { type ProjectDetail, type ConfigItem, ProjectService } from '@/services/project.service';
import { type TeamListItem, TeamService } from '@/services/team.service';
import { useOrganizationStore } from '@/stores/organization';
import { useConfigEncryption } from '@/composables/useConfigEncryption';
import { EncryptionService } from '@/services/encryption.service';
import { IdentityService } from '@/services/identity.service';
import { toast } from '@/lib/toast';

interface ConfigItemSelection {
    item: ConfigItem;
    include: boolean;
    includeValue: boolean;
}

const props = defineProps<{
    open: boolean;
    project: ProjectDetail;
    decryptedKey: string;
}>();

const emit = defineEmits<{
    (e: 'update:open', value: boolean): void;
    (e: 'cloned', projectId: string): void;
}>();

const router = useRouter();
const organizationStore = useOrganizationStore();
const { fetchAndDecryptConfig, encryptConfigItems } = useConfigEncryption();

const isOpen = computed({
    get: () => props.open,
    set: (val) => emit('update:open', val),
});

// Form state
const newProjectName = ref('');
const selectedTeamId = ref('');
const configSelections = ref<ConfigItemSelection[]>([]);

// Loading states
const isLoadingTeams = ref(false);
const isLoadingConfig = ref(false);
const isCloning = ref(false);
const error = ref('');

// Data
const availableTeams = ref<TeamListItem[]>([]);

// Reset form when dialog opens
watch(() => props.open, async (newVal) => {
    if (newVal) {
        newProjectName.value = `${props.project.name} (Copy)`;
        selectedTeamId.value = props.project.teamId;
        error.value = '';
        await loadData();
    }
});

async function loadData() {
    await Promise.all([
        loadTeams(),
        loadConfigItems(),
    ]);
}

async function loadTeams() {
    isLoadingTeams.value = true;
    try {
        const teams = await TeamService.getTeams(props.project.organizationId);
        // Filter to teams where user has access (they'll be returned by API if user has access)
        availableTeams.value = teams;
    } catch (e: any) {
        console.error('Failed to load teams', e);
        error.value = 'Failed to load teams: ' + e.message;
    } finally {
        isLoadingTeams.value = false;
    }
}

async function loadConfigItems() {
    isLoadingConfig.value = true;
    try {
        const decryptedItems = await fetchAndDecryptConfig(props.project.id, props.decryptedKey);
        configSelections.value = decryptedItems.map(item => ({
            item,
            include: true,
            includeValue: true,
        }));
    } catch (e: any) {
        console.error('Failed to load config items', e);
        error.value = 'Failed to load config items: ' + e.message;
    } finally {
        isLoadingConfig.value = false;
    }
}

const isLoading = computed(() => isLoadingTeams.value || isLoadingConfig.value);

const canClone = computed(() => {
    return newProjectName.value.trim() && selectedTeamId.value && !isLoading.value && !isCloning.value;
});

const selectedItemsCount = computed(() => {
    return configSelections.value.filter(s => s.include).length;
});

const allSelected = computed(() => {
    if (configSelections.value.length === 0) return false;
    return configSelections.value.every(s => s.include);
});

const isIndeterminate = computed(() => {
    const count = selectedItemsCount.value;
    return count > 0 && count < configSelections.value.length;
});

function toggleAll() {
    const newValue = !allSelected.value;
    configSelections.value.forEach(s => {
        s.include = newValue;
        if (!newValue) {
            s.includeValue = false;
        }
    });
}

// Toggle include and reset includeValue when unchecked
function onIncludeChange(index: number) {
    const selection = configSelections.value[index];
    selection.include = !selection.include;
    if (!selection.include) {
        selection.includeValue = false;
    }
}

async function handleClone() {
    if (!canClone.value) return;

    isCloning.value = true;
    error.value = '';

    try {
        // 1. Get target team to encrypt project key for it
        const targetTeam = availableTeams.value.find(t => t.id === selectedTeamId.value);
        if (!targetTeam) throw new Error('Target team not found');

        // Decrypt target team key
        let targetTeamKey = '';
        const masterKeyPair = IdentityService.getMasterKeyPair();
        if (!masterKeyPair) throw new Error('Master Identity not loaded');

        if (targetTeam.userEncryptedKey) {
            try {
                targetTeamKey = await EncryptionService.decryptKey(masterKeyPair.privateKey, targetTeam.userEncryptedKey);
            } catch (e) {
                console.error('Failed to decrypt team key from userEncryptedKey', e);
            }
        }

        // Fallback to org key decryption
        if (!targetTeamKey) {
            const orgKey = await organizationStore.unlockOrganization(props.project.organizationId);
            if (orgKey && targetTeam.encryptedKey) {
                try {
                    targetTeamKey = await EncryptionService.decryptValue(orgKey, targetTeam.encryptedKey);
                } catch (e) {
                    console.error('Failed to decrypt team key from org key', e);
                }
            }
        }

        if (!targetTeamKey) {
            throw new Error('Cannot clone: Unable to access target team encryption key');
        }

        // 2. Generate new project key
        const newProjectKey = EncryptionService.generateAesKey();

        // 3. Encrypt new project key with target team key
        const encryptedProjectKey = await EncryptionService.encryptValue(targetTeamKey, newProjectKey);

        // 4. Create new project
        const newProject = await ProjectService.createProject({
            name: newProjectName.value.trim(),
            organizationId: props.project.organizationId,
            teamId: selectedTeamId.value,
            encryptedKey: encryptedProjectKey,
        });

        // 5. Prepare config items for new project
        const selectedItems = configSelections.value.filter(s => s.include);

        if (selectedItems.length > 0) {
            const newConfigItems: ConfigItem[] = selectedItems.map((selection, index) => ({
                id: crypto.randomUUID(),
                projectId: newProject.id,
                name: selection.item.name,
                value: selection.includeValue ? selection.item.value : '', // plaintext value
                sensitive: selection.item.sensitive,
                position: index,
                category: selection.item.category,
                description: selection.item.description,
                expiresAt: selection.item.expiresAt,
            }));

            // 6. Encrypt and sync config items to new project
            const encryptedItems = await encryptConfigItems(newProjectKey, newConfigItems);
            await ProjectService.syncConfig(newProject.id, encryptedItems);
        }

        toast.success(`Project cloned successfully`);
        emit('cloned', newProject.id);
        isOpen.value = false;

        // Navigate to new project
        router.push(`/projects/${newProject.id}`);
    } catch (e: any) {
        console.error('Failed to clone project', e);
        error.value = e.message || 'Failed to clone project';
    } finally {
        isCloning.value = false;
    }
}
</script>

<template>
    <Dialog v-model:open="isOpen">
        <DialogContent class="sm:max-w-[600px] max-h-[85vh] flex flex-col">
            <DialogHeader>
                <DialogTitle class="flex items-center gap-2">
                    <Copy class="w-5 h-5" />
                    Clone Project
                </DialogTitle>
                <DialogDescription>
                    Create a copy of "{{ project.name }}" with selected configuration items.
                </DialogDescription>
            </DialogHeader>

            <div v-if="error" class="text-sm text-destructive bg-destructive/10 p-3 rounded-md">
                {{ error }}
            </div>

            <div v-if="isLoading" class="flex items-center justify-center py-8 text-muted-foreground">
                <Loader2 class="w-5 h-5 animate-spin mr-2" />
                Loading...
            </div>

            <div v-else class="flex-1 overflow-hidden flex flex-col gap-4 px-0.5">
                <!-- Project Name -->
                <div class="space-y-2">
                    <Label for="newProjectName">New Project Name</Label>
                    <Input
                        id="newProjectName"
                        v-model="newProjectName"
                        placeholder="Enter project name"
                        @keyup.enter="handleClone"
                    />
                </div>

                <!-- Team Selection -->
                <div class="space-y-2">
                    <Label for="teamSelect">Target Team</Label>
                    <Select v-model="selectedTeamId">
                        <SelectTrigger>
                            <SelectValue placeholder="Select a team" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem v-for="team in availableTeams" :key="team.id" :value="team.id">
                                {{ team.name }}
                            </SelectItem>
                        </SelectContent>
                    </Select>
                    <p class="text-xs text-muted-foreground">
                        The cloned project will be created in this team.
                    </p>
                </div>

                <!-- Config Items Selection -->
                <div class="space-y-2 flex-1 overflow-hidden flex flex-col">
                    <Label>Configuration Items ({{ selectedItemsCount }}/{{ configSelections.length }})</Label>

                    <div v-if="configSelections.length === 0" class="text-sm text-muted-foreground py-4 text-center">
                        No configuration items to clone.
                    </div>

                    <div v-else class="flex-1 overflow-y-auto border rounded-md bg-background">
                        <!-- Header -->
                        <div class="flex items-center gap-4 px-3 py-2 bg-muted border-b text-xs font-medium text-muted-foreground sticky top-0 z-10">
                            <Checkbox
                                :model-value="allSelected ? true : isIndeterminate ? 'indeterminate' : false"
                                @update:model-value="toggleAll"
                            />
                            <div class="flex-1">Name</div>
                            <div class="w-24 text-center">Include Value</div>
                        </div>

                        <!-- Items -->
                        <div class="divide-y">
                            <div
                                v-for="(selection, index) in configSelections"
                                :key="selection.item.id"
                                class="flex items-center gap-4 px-3 py-2 hover:bg-muted/30 transition-colors"
                                :class="{ 'opacity-50': !selection.include }"
                            >
                                <Checkbox
                                    :model-value="selection.include"
                                    @update:model-value="onIncludeChange(index)"
                                />
                                <div class="flex-1 min-w-0">
                                    <div class="font-mono text-sm truncate">{{ selection.item.name }}</div>
                                    <div v-if="selection.item.category" class="text-xs text-muted-foreground">
                                        {{ selection.item.category }}
                                    </div>
                                </div>
                                <div class="w-24 flex justify-center">
                                    <Checkbox
                                        v-model:model-value="selection.includeValue"
                                        :disabled="!selection.include"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>

                    <p class="text-xs text-muted-foreground">
                        Unchecking "Include Value" will copy only the variable name with an empty value.
                    </p>
                </div>
            </div>

            <DialogFooter>
                <Button variant="outline" @click="isOpen = false" :disabled="isCloning">
                    Cancel
                </Button>
                <Button @click="handleClone" :disabled="!canClone">
                    <Loader2 v-if="isCloning" class="w-4 h-4 mr-2 animate-spin" />
                    <Copy v-else class="w-4 h-4 mr-2" />
                    {{ isCloning ? 'Cloning...' : 'Clone Project' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
