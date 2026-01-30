<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Loader2, ArrowRightLeft } from 'lucide-vue-next';
import { type ProjectDetail, type ConfigItem, ProjectService } from '@/services/project.service';
import { useConfigEncryption } from '@/composables/useConfigEncryption';
import { useProjectDecryption } from '@/composables/useProjectDecryption';
import { useProjects } from '@/queries';
import { toast } from '@/lib/toast';

type DiffStatus = 'unchanged' | 'changed' | 'added' | 'removed';
type DiffAction = 'none' | 'accept' | 'keep' | 'ignore' | 'delete';

interface DiffItem {
    name: string;
    category?: string;
    baseValue?: string;
    targetValue?: string;
    baseSensitive?: boolean;
    targetSensitive?: boolean;
    basePosition?: number;
    status: DiffStatus;
    action: DiffAction;
}

const props = defineProps<{
    open: boolean;
    project: ProjectDetail;
    decryptedKey: string;
}>();

const emit = defineEmits<{
    (e: 'update:open', value: boolean): void;
    (e: 'applied'): void;
}>();

const { fetchAndDecryptConfig, encryptAndSyncConfig } = useConfigEncryption();
const { decryptProjectKeys } = useProjectDecryption();

const isOpen = computed({
    get: () => props.open,
    set: (val) => emit('update:open', val),
});

// State
const selectedBaseId = ref('');
const isLoadingDiff = ref(false);
const isApplying = ref(false);
const error = ref('');

// Track current target (can be swapped)
const targetProjectId = ref('');
const targetProject = ref<ProjectDetail | null>(null);
const targetDecryptedKey = ref('');

// Data
const { data: allProjects, isLoading: projectsLoading } = useProjects();
const baseConfigItems = ref<ConfigItem[]>([]);
const targetConfigItems = ref<ConfigItem[]>([]);
const diffItems = ref<DiffItem[]>([]);

// Filter projects user can compare (exclude current target)
const availableProjects = computed(() => {
    if (!allProjects.value) return [];
    return allProjects.value.filter(p => p.id !== targetProjectId.value);
});

// Get target project name for display
const targetProjectName = computed(() => {
    return targetProject.value?.name || '';
});

// Reset when dialog opens
watch(() => props.open, async (newVal) => {
    if (newVal) {
        // Initialize target as the current project
        targetProjectId.value = props.project.id;
        targetProject.value = props.project;
        targetDecryptedKey.value = props.decryptedKey;

        selectedBaseId.value = '';
        baseConfigItems.value = [];
        targetConfigItems.value = [];
        diffItems.value = [];
        error.value = '';
        // Load target config items
        await loadTargetConfig();
    }
});

// Load diff when base project is selected
watch(selectedBaseId, async (newVal) => {
    if (newVal && newVal !== targetProjectId.value) {
        await loadBaseConfigAndComputeDiff();
    } else {
        baseConfigItems.value = [];
        diffItems.value = [];
    }
});

async function loadTargetConfig() {
    if (!targetProjectId.value || !targetDecryptedKey.value) return;

    try {
        targetConfigItems.value = await fetchAndDecryptConfig(targetProjectId.value, targetDecryptedKey.value);
    } catch (e: any) {
        console.error('Failed to load target config', e);
        error.value = 'Failed to load target config: ' + e.message;
    }
}

async function loadBaseConfigAndComputeDiff() {
    if (!selectedBaseId.value) return;

    isLoadingDiff.value = true;
    error.value = '';

    try {
        // Get base project details
        const baseProject = await ProjectService.getProject(selectedBaseId.value);

        // Decrypt base project keys
        const { projectKey: baseProjectKey } = await decryptProjectKeys({
            teamId: baseProject.teamId,
            organizationId: baseProject.organizationId,
            encryptedTeamKey: baseProject.encryptedTeamKey,
            encryptedProjectKey: baseProject.encryptedProjectKey,
        });

        // Fetch and decrypt base config
        baseConfigItems.value = await fetchAndDecryptConfig(selectedBaseId.value, baseProjectKey);

        // Compute diff
        computeDiff();
    } catch (e: any) {
        console.error('Failed to load base config', e);
        error.value = 'Failed to load base config: ' + e.message;
    } finally {
        isLoadingDiff.value = false;
    }
}

function computeDiff() {
    const result: DiffItem[] = [];
    const targetMap = new Map(targetConfigItems.value.map(item => [item.name, item]));
    const processedTargetNames = new Set<string>();

    // Process items from base (maintains base order)
    for (const baseItem of baseConfigItems.value) {
        const targetItem = targetMap.get(baseItem.name);
        processedTargetNames.add(baseItem.name);

        if (!targetItem) {
            // Added: exists in base, not in target
            result.push({
                name: baseItem.name,
                category: baseItem.category,
                baseValue: baseItem.value,
                baseSensitive: baseItem.sensitive,
                basePosition: baseItem.position,
                status: 'added',
                action: 'ignore', // Default: don't add
            });
        } else if (baseItem.value !== targetItem.value) {
            // Changed: exists in both but different values
            result.push({
                name: baseItem.name,
                category: baseItem.category,
                baseValue: baseItem.value,
                targetValue: targetItem.value,
                baseSensitive: baseItem.sensitive,
                targetSensitive: targetItem.sensitive,
                basePosition: baseItem.position,
                status: 'changed',
                action: 'keep', // Default: keep target value
            });
        } else {
            // Unchanged
            result.push({
                name: baseItem.name,
                category: baseItem.category,
                baseValue: baseItem.value,
                targetValue: targetItem.value,
                baseSensitive: baseItem.sensitive,
                targetSensitive: targetItem.sensitive,
                basePosition: baseItem.position,
                status: 'unchanged',
                action: 'none',
            });
        }
    }

    // Process items only in target (removed from base perspective)
    for (const targetItem of targetConfigItems.value) {
        if (!processedTargetNames.has(targetItem.name)) {
            result.push({
                name: targetItem.name,
                category: targetItem.category,
                targetValue: targetItem.value,
                targetSensitive: targetItem.sensitive,
                status: 'removed',
                action: 'keep', // Default: keep in target
            });
        }
    }

    diffItems.value = result;
}

async function swapDirection() {
    if (!selectedBaseId.value) return;

    isLoadingDiff.value = true;
    error.value = '';

    try {
        // Get the current base project details to make it the new target
        const newTargetProject = await ProjectService.getProject(selectedBaseId.value);

        // Decrypt new target project keys
        const { projectKey: newTargetKey } = await decryptProjectKeys({
            teamId: newTargetProject.teamId,
            organizationId: newTargetProject.organizationId,
            encryptedTeamKey: newTargetProject.encryptedTeamKey,
            encryptedProjectKey: newTargetProject.encryptedProjectKey,
        });

        // Store current target as new base
        const newBaseId = targetProjectId.value;

        // Swap: old base becomes new target
        targetProjectId.value = selectedBaseId.value;
        targetProject.value = newTargetProject;
        targetDecryptedKey.value = newTargetKey;

        // Swap: old target becomes new base
        selectedBaseId.value = newBaseId;

        // Reload configs with swapped roles
        await loadTargetConfig();
        await loadBaseConfigAndComputeDiff();
    } catch (e: any) {
        console.error('Failed to swap direction', e);
        error.value = 'Failed to swap direction: ' + e.message;
    } finally {
        isLoadingDiff.value = false;
    }
}

// Bulk actions
function acceptAllNew() {
    diffItems.value.forEach(item => {
        if (item.status === 'added') {
            item.action = 'accept';
        }
    });
}

function ignoreAllNew() {
    diffItems.value.forEach(item => {
        if (item.status === 'added') {
            item.action = 'ignore';
        }
    });
}

function acceptAllChanges() {
    diffItems.value.forEach(item => {
        if (item.status === 'changed') {
            item.action = 'accept';
        }
    });
}

function keepAllChanges() {
    diffItems.value.forEach(item => {
        if (item.status === 'changed') {
            item.action = 'keep';
        }
    });
}

function deleteAllRemoved() {
    diffItems.value.forEach(item => {
        if (item.status === 'removed') {
            item.action = 'delete';
        }
    });
}

function keepAllRemoved() {
    diffItems.value.forEach(item => {
        if (item.status === 'removed') {
            item.action = 'keep';
        }
    });
}

// Stats
const stats = computed(() => {
    let additions = 0;
    let updates = 0;
    let deletions = 0;

    for (const item of diffItems.value) {
        if (item.status === 'added' && item.action === 'accept') additions++;
        if (item.status === 'changed' && item.action === 'accept') updates++;
        if (item.status === 'removed' && item.action === 'delete') deletions++;
    }

    return { additions, updates, deletions, total: additions + updates + deletions };
});

const hasChanges = computed(() => {
    return diffItems.value.some(item => item.status !== 'unchanged');
});

const hasChangedItems = computed(() => {
    return diffItems.value.some(item => item.status === 'changed');
});

const hasAddedItems = computed(() => {
    return diffItems.value.some(item => item.status === 'added');
});

const hasRemovedItems = computed(() => {
    return diffItems.value.some(item => item.status === 'removed');
});

const hasPendingActions = computed(() => stats.value.total > 0);

// Value display helpers
function truncateValue(value: string): string {
    return value.length > 50 ? value.substring(0, 50) + '...' : value;
}

// Apply changes
async function handleApply() {
    if (!hasPendingActions.value || !targetProjectId.value || !targetDecryptedKey.value) return;

    isApplying.value = true;
    error.value = '';

    try {
        // Build new config items for target
        const newConfigItems: ConfigItem[] = [];
        let position = 0;

        // Start with items from base order for accepted/changed items
        for (const diff of diffItems.value) {
            if (diff.status === 'added' && diff.action === 'accept') {
                // Add new item from base
                newConfigItems.push({
                    id: crypto.randomUUID(),
                    projectId: targetProjectId.value,
                    name: diff.name,
                    value: diff.baseValue!,
                    sensitive: diff.baseSensitive || false,
                    position: position++,
                    category: diff.category,
                });
            } else if (diff.status === 'changed') {
                const targetItem = targetConfigItems.value.find(t => t.name === diff.name)!;
                if (diff.action === 'accept') {
                    // Update with base value, category, and sensitive flag
                    newConfigItems.push({
                        ...targetItem,
                        value: diff.baseValue!,
                        sensitive: diff.baseSensitive || false,
                        category: diff.category,
                        position: position++,
                    });
                } else {
                    // Keep target value
                    newConfigItems.push({
                        ...targetItem,
                        position: position++,
                    });
                }
            } else if (diff.status === 'unchanged') {
                // Keep unchanged items
                const targetItem = targetConfigItems.value.find(t => t.name === diff.name)!;
                newConfigItems.push({
                    ...targetItem,
                    position: position++,
                });
            } else if (diff.status === 'removed' && diff.action === 'keep') {
                // Keep items that only exist in target
                const targetItem = targetConfigItems.value.find(t => t.name === diff.name)!;
                newConfigItems.push({
                    ...targetItem,
                    position: position++,
                });
            }
            // If status === 'removed' && action === 'delete', we simply don't add it
            // If status === 'added' && action === 'ignore', we don't add it
        }

        // Save to backend
        await encryptAndSyncConfig(targetProjectId.value, targetDecryptedKey.value, newConfigItems);

        toast.success(`Applied ${stats.value.total} changes to ${targetProjectName.value}`);
        emit('applied');
        isOpen.value = false;
    } catch (e: any) {
        console.error('Failed to apply changes', e);
        error.value = 'Failed to apply changes: ' + e.message;
    } finally {
        isApplying.value = false;
    }
}

</script>

<template>
    <Dialog v-model:open="isOpen">
        <DialogContent class="sm:max-w-[900px] max-h-[85vh] flex flex-col">
            <DialogHeader>
                <DialogTitle>Compare Projects</DialogTitle>
                <DialogDescription>
                    Compare configuration between projects and sync changes.
                </DialogDescription>
            </DialogHeader>

            <div v-if="error" class="text-sm text-destructive bg-destructive/10 p-3 rounded-md">
                {{ error }}
            </div>

            <!-- Project Selection -->
            <div class="flex items-center gap-4 p-4 bg-muted/30 rounded-lg">
                <div class="flex-1 space-y-1">
                    <Label class="text-xs text-muted-foreground">Base (source)</Label>
                    <Select v-model="selectedBaseId" :disabled="isLoadingDiff">
                        <SelectTrigger>
                            <SelectValue placeholder="Select base project..." />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem v-for="p in availableProjects" :key="p.id" :value="p.id">
                                {{ p.name }}
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <Button variant="ghost" size="icon" class="mt-5" @click="swapDirection" :disabled="!selectedBaseId">
                    <ArrowRightLeft class="w-4 h-4" />
                </Button>

                <div class="flex-1 space-y-1">
                    <Label class="text-xs text-muted-foreground">Target (will be modified)</Label>
                    <div class="h-10 px-3 flex items-center border rounded-md bg-background text-sm">
                        {{ targetProjectName }}
                    </div>
                </div>
            </div>

            <!-- Loading State -->
            <div v-if="isLoadingDiff || projectsLoading" class="flex items-center justify-center py-12 text-muted-foreground">
                <Loader2 class="w-5 h-5 animate-spin mr-2" />
                Loading...
            </div>

            <!-- No Selection State -->
            <div v-else-if="!selectedBaseId" class="flex items-center justify-center py-12 text-muted-foreground">
                Select a base project to compare
            </div>

            <!-- Diff View -->
            <div v-else class="flex-1 overflow-hidden flex flex-col gap-3">
                <!-- Bulk Actions -->
                <div v-if="hasChanges" class="flex flex-wrap items-center gap-2 text-xs">
                    <span class="text-muted-foreground mr-2">Bulk:</span>
                    <template v-if="hasChangedItems">
                        <Button variant="outline" size="sm" class="h-7 text-xs" @click="acceptAllChanges">
                            Accept All Changes
                        </Button>
                        <Button variant="outline" size="sm" class="h-7 text-xs" @click="keepAllChanges">
                            Keep All Target Values
                        </Button>
                    </template>
                    <template v-if="hasAddedItems">
                        <Button variant="outline" size="sm" class="h-7 text-xs text-green-600" @click="acceptAllNew">
                            Accept All New
                        </Button>
                        <Button variant="outline" size="sm" class="h-7 text-xs" @click="ignoreAllNew">
                            Ignore All New
                        </Button>
                    </template>
                    <template v-if="hasRemovedItems">
                        <Button variant="outline" size="sm" class="h-7 text-xs text-red-600" @click="deleteAllRemoved">
                            Delete All Removed
                        </Button>
                        <Button variant="outline" size="sm" class="h-7 text-xs" @click="keepAllRemoved">
                            Keep All Removed
                        </Button>
                    </template>
                </div>

                <!-- Diff List -->
                <div class="flex-1 overflow-y-auto border rounded-md bg-background">
                    <!-- Header -->
                    <div class="flex items-center gap-2 px-3 py-2 bg-muted border-b text-xs font-medium text-muted-foreground sticky top-0 z-10">
                        <div class="w-3"></div>
                        <div class="w-48">Name</div>
                        <div class="flex-1">Base Value</div>
                        <div class="flex-1">Target Value</div>
                        <div class="w-48 text-right">Action</div>
                    </div>

                    <!-- No Changes -->
                    <div v-if="!hasChanges" class="py-8 text-center text-muted-foreground">
                        No differences found. Projects are in sync.
                    </div>

                    <!-- Items -->
                    <div v-else class="divide-y">
                        <div
                            v-for="item in diffItems"
                            :key="item.name"
                            class="flex items-center gap-2 px-3 py-2 text-sm"
                        >
                            <!-- Status Dot -->
                            <div
                                class="w-2 h-2 rounded-full shrink-0"
                                :class="{
                                    'bg-green-500': item.status === 'added',
                                    'bg-red-500': item.status === 'removed',
                                    'bg-orange-500': item.status === 'changed',
                                    'bg-muted-foreground/30': item.status === 'unchanged',
                                }"
                            ></div>
                            <!-- Name -->
                            <div class="w-48 min-w-0">
                                <div class="font-mono text-sm truncate" :title="item.name">{{ item.name }}</div>
                                <div v-if="item.category" class="text-xs text-muted-foreground truncate">
                                    {{ item.category }}
                                </div>
                            </div>

                            <!-- Base Value -->
                            <div class="flex-1 min-w-0">
                                <div v-if="item.baseValue !== undefined" class="font-mono text-xs truncate" :title="item.baseValue">
                                    {{ truncateValue(item.baseValue) }}
                                </div>
                                <div v-else class="text-xs text-muted-foreground italic">
                                    (not in base)
                                </div>
                            </div>

                            <!-- Target Value -->
                            <div class="flex-1 min-w-0">
                                <div v-if="item.targetValue !== undefined" class="font-mono text-xs truncate" :title="item.targetValue">
                                    {{ truncateValue(item.targetValue) }}
                                </div>
                                <div v-else class="text-xs text-muted-foreground italic">
                                    (not in target)
                                </div>
                            </div>

                            <!-- Actions -->
                            <div class="w-48 flex justify-end gap-1">
                                <!-- Changed: Accept Base or Keep Target -->
                                <template v-if="item.status === 'changed'">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-green-500/20 text-green-600': item.action === 'accept' }"
                                        @click="item.action = 'accept'"
                                    >
                                        Accept
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-muted': item.action === 'keep' }"
                                        @click="item.action = 'keep'"
                                    >
                                        Keep
                                    </Button>
                                </template>

                                <!-- Added: Accept or Ignore -->
                                <template v-else-if="item.status === 'added'">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-green-500/20 text-green-600': item.action === 'accept' }"
                                        @click="item.action = 'accept'"
                                    >
                                        Accept
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-muted': item.action === 'ignore' }"
                                        @click="item.action = 'ignore'"
                                    >
                                        Ignore
                                    </Button>
                                </template>

                                <!-- Removed: Delete or Keep -->
                                <template v-else-if="item.status === 'removed'">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-red-500/20 text-red-600': item.action === 'delete' }"
                                        @click="item.action = 'delete'"
                                    >
                                        Delete
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 text-xs"
                                        :class="{ 'bg-muted': item.action === 'keep' }"
                                        @click="item.action = 'keep'"
                                    >
                                        Keep
                                    </Button>
                                </template>

                                <!-- Unchanged: no actions -->
                                <template v-else>
                                    <span class="text-xs text-muted-foreground">No changes</span>
                                </template>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Summary -->
                <div class="flex items-center justify-between text-sm text-muted-foreground px-1">
                    <div>
                        <span v-if="stats.updates > 0" class="text-yellow-600">{{ stats.updates }} update{{ stats.updates !== 1 ? 's' : '' }}</span>
                        <span v-if="stats.updates > 0 && (stats.additions > 0 || stats.deletions > 0)">, </span>
                        <span v-if="stats.additions > 0" class="text-green-600">{{ stats.additions }} addition{{ stats.additions !== 1 ? 's' : '' }}</span>
                        <span v-if="stats.additions > 0 && stats.deletions > 0">, </span>
                        <span v-if="stats.deletions > 0" class="text-red-600">{{ stats.deletions }} deletion{{ stats.deletions !== 1 ? 's' : '' }}</span>
                        <span v-if="stats.total === 0">No pending changes</span>
                    </div>
                </div>
            </div>

            <DialogFooter>
                <Button variant="outline" @click="isOpen = false" :disabled="isApplying">
                    Cancel
                </Button>
                <Button @click="handleApply" :disabled="!hasPendingActions || isApplying">
                    <Loader2 v-if="isApplying" class="w-4 h-4 mr-2 animate-spin" />
                    Apply{{ stats.total > 0 ? ` (${stats.total})` : '' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
