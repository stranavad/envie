<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { type ConfigItem, type ProjectDetail, ProjectService } from '@/services/project.service';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
    AlertTriangle,
    CloudDownload,
    Copy,
    FileUp,
    Plus,
    Save,
    Check,
    ChevronRight,
    FolderPlus,
    GripVertical,
    Pencil,
    Trash2,
    FolderOpen,
    Loader2,
    Wand2
} from 'lucide-vue-next';
import ConfigItemRow from './ConfigItemRow.vue';
import SecretImportDialog from './SecretImportDialog.vue';
import AddItemDialog from './dialogs/AddItemDialog.vue';
import AddCategoryDialog from './dialogs/AddCategoryDialog.vue';
import RenameCategoryDialog from './dialogs/RenameCategoryDialog.vue';
import EnvImportDialog from './dialogs/EnvImportDialog.vue';
import { SecretManagerConfigService } from '@/services/secret-manager-config.service';
import { FileMappingService } from '@/services/file-mapping.service';
import { useCategoryManagement } from '@/composables/useCategoryManagement';
import { useConfigEncryption } from '@/composables/useConfigEncryption';
import { copyEnvToClipboard, parseEnvString, type ParsedEnvItem } from '@/utils/env-format';
import { IconButton } from '@/components/ui/icon-button';
import { SortableContainer } from '@/components/ui/sortable';
import { SectionHeader } from '@/components/ui/section-header';
import { RefreshCw } from 'lucide-vue-next';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
    localImportItems?: { name: string; value: string }[] | null;
    syncMode?: boolean;
}>();

const emit = defineEmits<{
    (e: 'syncComplete'): void;
}>();

const { fetchAndDecryptConfig, encryptAndSyncConfig } = useConfigEncryption();

// Watch for local import items from Push operation (review mode)
watch(() => props.localImportItems, (items) => {
    if (items && items.length > 0) {
        handleLocalImport(items);
    }
}, { immediate: true });

function handleLocalImport(items: { name: string; value: string }[]) {
    items.forEach(newItem => {
        const existingIndex = configItems.value.findIndex(ci => ci.name === newItem.name);

        if (existingIndex !== -1) {
            configItems.value[existingIndex].value = newItem.value;
        } else {
            configItems.value.push({
                id: crypto.randomUUID(),
                projectId: props.project.id,
                name: newItem.name,
                value: newItem.value,
                sensitive: true,
                position: configItems.value.length,
            });
        }
    });
}

// Core state
const isLoading = ref(true);
const originalConfigItems = ref<ConfigItem[]>([]);
const configItems = ref<ConfigItem[]>([]);
const isSaving = ref(false);
const saveError = ref('');
const isCopying = ref(false);
const hasSecretManagerConfigs = ref(false);
const isDragging = ref(false);

// Dialog visibility
const showSecretImportDialog = ref(false);
const showAddItemDialog = ref(false);
const showAddCategoryDialog = ref(false);
const showRenameCategoryDialog = ref(false);
const showEnvImportDialog = ref(false);
const renamingCategoryName = ref('');

// Empty state env input
const envInput = ref('');

// Category management
const {
    categories,
    hasCategories,
    categoriesForDrag,
    uncategorizedItems,
    getCategoryItems,
    toggleCategory,
    isCategoryCollapsed,
    addCategory,
    renameCategory,
    deleteCategory,
    recalculatePositions,
    onCategoryDragUpdate,
    onCategoryItemsChange,
    onUncategorizedChange,
    moveItemToCategory,
    autoGroupByPrefix,
} = useCategoryManagement(configItems);

// Helper to compare nullable/undefined values
function nullableEquals(a: unknown, b: unknown): boolean {
    if (a == null && b == null) return true;
    return a === b;
}

function isModified(item: ConfigItem, original: ConfigItem | null): boolean {
    if (!original) return false;

    return item.name !== original.name ||
        item.value !== original.value ||
        item.sensitive !== original.sensitive ||
        item.position !== original.position ||
        !nullableEquals(item.category, original.category) ||
        !nullableEquals(item.secretManagerName, original.secretManagerName) ||
        !nullableEquals(item.secretManagerConfigId, original.secretManagerConfigId) ||
        !nullableEquals(item.secretManagerVersion, original.secretManagerVersion) ||
        !nullableEquals(item.secretManagerLastSyncAt, original.secretManagerLastSyncAt);
}

function isAdded(item: ConfigItem): boolean {
    return !originalConfigItems.value.find(o => o.id === item.id);
}

const hasChanges = computed(() => {
    if (configItems.value.length !== originalConfigItems.value.length) return true;

    return configItems.value.some((item) => {
        const original = originalConfigItems.value.find(o => o.id === item.id);
        return isModified(item, original ?? null);
    });
});

async function loadConfig() {
    try {
        // Fetch and decrypt config items using composable
        const decryptedConfigs = await fetchAndDecryptConfig(props.project.id, props.decryptedKey);
        configItems.value = decryptedConfigs;
        originalConfigItems.value = structuredClone(decryptedConfigs);
    } catch (e) {
        console.error('Failed to load config', e);
    }
}

async function loadSecretManagerConfigs() {
    try {
        const configs = await SecretManagerConfigService.getConfigs(props.project.id);
        hasSecretManagerConfigs.value = configs.length > 0;
    } catch (e) {
        console.error('Failed to load secret manager configs', e);
        hasSecretManagerConfigs.value = false;
    }
}

async function handleSave() {
    isSaving.value = true;
    saveError.value = '';

    try {
        recalculatePositions();

        // Encrypt and sync using composable
        await encryptAndSyncConfig(props.project.id, props.decryptedKey, configItems.value);
        await loadConfig();
    } catch (e: unknown) {
        saveError.value = 'Failed to save: ' + String(e);
    } finally {
        isSaving.value = false;
    }
}

async function handleSaveAndSync() {
    isSaving.value = true;
    saveError.value = '';

    try {
        recalculatePositions();

        // Encrypt and save to remote using composable
        await encryptAndSyncConfig(props.project.id, props.decryptedKey, configItems.value);

        // Get the file mapping to write back to local file
        const mapping = await FileMappingService.getMapping(props.project.id);
        if (mapping) {
            // Write current config to local file
            const localChecksum = await FileMappingService.writeToLocalFile(
                mapping.filePath,
                configItems.value
            );

            // Get updated project checksum
            const updatedProject = await ProjectService.getProject(props.project.id);

            // Update mapping with new checksums
            await FileMappingService.updateSyncState(
                props.project.id,
                localChecksum,
                updatedProject.configChecksum || ''
            );
        }

        await loadConfig();
        emit('syncComplete');
    } catch (e: unknown) {
        saveError.value = 'Failed to save and sync: ' + String(e);
    } finally {
        isSaving.value = false;
    }
}

async function handleCopyEnv() {
    const success = await copyEnvToClipboard(
        configItems.value,
        categories.value,
        getCategoryItems,
        () => uncategorizedItems.value
    );

    if (success) {
        isCopying.value = true;
        setTimeout(() => isCopying.value = false, 2000);
    }
}

function handleAddItem(item: { name: string; value: string; sensitive: boolean; category?: string }) {
    configItems.value.push({
        id: crypto.randomUUID(),
        projectId: props.project.id,
        name: item.name,
        value: item.value,
        sensitive: item.sensitive,
        position: configItems.value.length,
        category: item.category,
    });
}

function handleEnvImport(items: ParsedEnvItem[], markAsSensitive: boolean) {
    items.forEach(newItem => {
        const existingIndex = configItems.value.findIndex(ci => ci.name === newItem.name);

        if (existingIndex !== -1) {
            configItems.value[existingIndex].value = newItem.value;
        } else {
            configItems.value.push({
                id: crypto.randomUUID(),
                projectId: props.project.id,
                name: newItem.name,
                value: newItem.value,
                sensitive: markAsSensitive,
                position: configItems.value.length,
            });
        }
    });
}

function handleSecretManagerImport(items: Partial<ConfigItem>[]) {
    items.forEach(newItem => {
        const existingIndex = configItems.value.findIndex(c => c.name === newItem.name);

        if (existingIndex !== -1) {
            configItems.value[existingIndex] = {
                ...configItems.value[existingIndex],
                ...newItem,
                value: newItem.value ?? configItems.value[existingIndex].value
            } as ConfigItem;
        } else {
            configItems.value.push({
                ...newItem,
                id: crypto.randomUUID(),
                projectId: props.project.id,
                position: configItems.value.length,
                sensitive: true
            } as ConfigItem);
        }
    });
}

function parseEnvEmptyState() {
    const items = parseEnvString(envInput.value);
    items.forEach(item => {
        configItems.value.push({
            id: crypto.randomUUID(),
            projectId: props.project.id,
            name: item.name,
            value: item.value,
            sensitive: true,
            position: configItems.value.length,
        });
    });
    envInput.value = '';
}

function deleteItem(id: string) {
    configItems.value = configItems.value.filter(item => item.id !== id);
}

function updateItem(configItem: ConfigItem) {
    configItems.value = configItems.value.map(item =>
        item.id === configItem.id ? configItem : item
    );
}

function startRenameCategory(category: string) {
    renamingCategoryName.value = category;
    showRenameCategoryDialog.value = true;
}

function handleRenameCategory(oldName: string, newName: string) {
    renameCategory(oldName, newName);
}

function onDragEnd() {
    isDragging.value = false;
    recalculatePositions();
}

// Handler for reordering when there are no categories
function onNoCategoryReorder(newItems: ConfigItem[]) {
    // Update positions directly based on new array order
    newItems.forEach((item, index) => {
        item.position = index;
    });
    configItems.value = newItems;
}

// Initialize
async function initialize() {
    try {
        await Promise.all([
            loadConfig(),
            loadSecretManagerConfigs()
        ]);
    } finally {
        isLoading.value = false;
    }
}
initialize();
</script>

<template>
    <div class="space-y-4 min-w-0">
        <!-- Loading State -->
        <div v-if="isLoading" class="flex items-center justify-center py-20">
            <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Empty State -->
        <div v-else-if="configItems.length === 0" class="border border-dashed rounded-lg p-8">
            <div class="max-w-lg mx-auto space-y-4">
                <div class="text-center space-y-2">
                    <h3 class="text-lg font-medium">No config items yet</h3>
                    <p class="text-sm text-muted-foreground">
                        Paste your .env file below to get started, or add items manually.
                    </p>
                </div>

                <Textarea
                    v-model="envInput"
                    placeholder="API_KEY=12345&#10;DB_HOST=localhost&#10;SECRET_TOKEN=abc123"
                    class="font-mono text-sm min-h-[120px]"
                />

                <div class="flex justify-center gap-3">
                    <Button @click="parseEnvEmptyState" :disabled="!envInput.trim()">
                        <FileUp class="w-4 h-4 mr-2" />
                        Import .env
                    </Button>
                </div>

                <div class="relative">
                    <div class="absolute inset-0 flex items-center">
                        <span class="w-full border-t" />
                    </div>
                    <div class="relative flex justify-center text-xs uppercase">
                        <span class="bg-background px-2 text-muted-foreground">or</span>
                    </div>
                </div>

                <div class="flex justify-center">
                    <Button variant="outline" @click="showAddItemDialog = true">
                        <Plus class="w-4 h-4 mr-2" />
                        Add Item Manually
                    </Button>
                </div>
            </div>
        </div>

        <!-- With Items: Toolbar + List -->
        <div v-else class="space-y-4 min-w-0 overflow-hidden">
            <!-- Sync Mode Banner -->
            <div v-if="syncMode" class="p-3 bg-orange-500/10 border border-orange-500/40 rounded-lg">
                <p class="text-sm text-orange-200">
                    <strong>Reviewing local changes.</strong> Make any adjustments, then click "Save & Sync" to save to Envie and update your local .env file.
                </p>
            </div>

            <!-- Toolbar -->
            <SectionHeader :title="`Config Items (${configItems.length})`">
                <template #actions>
                    <div v-if="hasChanges" class="flex items-center text-xs text-orange-500 font-medium whitespace-nowrap mr-2">
                        <AlertTriangle class="w-3 h-3 mr-1" />
                        <span>Unsaved Changes</span>
                    </div>
                    <IconButton tooltip="Add Item" @click="showAddItemDialog = true">
                        <Plus class="w-4 h-4" />
                    </IconButton>
                    <IconButton tooltip="Add Category" @click="showAddCategoryDialog = true">
                        <FolderPlus class="w-4 h-4" />
                    </IconButton>
                    <IconButton tooltip="Auto Group by Prefix" @click="autoGroupByPrefix">
                        <Wand2 class="w-4 h-4" />
                    </IconButton>
                    <IconButton tooltip="Import .env" @click="showEnvImportDialog = true">
                        <FileUp class="w-4 h-4" />
                    </IconButton>
                    <IconButton tooltip="From Secret Manager" @click="showSecretImportDialog = true">
                        <CloudDownload class="w-4 h-4" />
                    </IconButton>
                    <IconButton :tooltip="isCopying ? 'Copied!' : 'Copy as .env'" @click="handleCopyEnv">
                        <Check v-if="isCopying" class="w-4 h-4 text-green-500" />
                        <Copy v-else class="w-4 h-4" />
                    </IconButton>
                    <Button
                        v-if="syncMode"
                        size="sm"
                        variant="outline"
                        class="ml-2"
                        @click="handleSaveAndSync"
                        :disabled="isSaving || !decryptedKey || !hasChanges"
                    >
                        <RefreshCw class="w-4 h-4 mr-2" />
                        {{ isSaving ? 'Syncing...' : 'Save & Sync' }}
                    </Button>
                    <Button
                        v-else
                        size="sm"
                        variant="outline"
                        class="ml-2"
                        @click="handleSave"
                        :disabled="isSaving || !decryptedKey || !hasChanges"
                    >
                        <Save class="w-4 h-4 mr-2" />
                        {{ isSaving ? 'Saving...' : 'Save' }}
                    </Button>
                </template>
            </SectionHeader>

            <div v-if="saveError" class="text-sm text-destructive">{{ saveError }}</div>

            <!-- Categories (sortable for reordering) -->
            <SortableContainer
                v-if="hasCategories"
                :model-value="categoriesForDrag"
                @update:model-value="onCategoryDragUpdate"
                item-key="name"
                handle=".drag-handle-category"
                ghost-class="category-ghost"
                class="flex flex-col gap-3"
                @start="isDragging = true"
                @end="onDragEnd"
            >
                <template #item="{ element: cat }">
                    <div class="border rounded-lg overflow-hidden">
                        <!-- Category Header -->
                        <div
                            class="flex items-center gap-2 px-3 py-2 bg-muted/50 cursor-pointer hover:bg-muted/70 transition-colors"
                            @click="toggleCategory(cat.name)"
                        >
                            <div class="drag-handle-category cursor-grab active:cursor-grabbing p-1 -m-1 rounded hover:bg-muted transition-colors" @click.stop>
                                <GripVertical class="w-4 h-4 text-muted-foreground" />
                            </div>
                            <ChevronRight
                                class="w-4 h-4 text-muted-foreground transition-transform duration-200"
                                :class="{ 'rotate-90': !isCategoryCollapsed(cat.name) }"
                            />
                            <FolderOpen class="w-4 h-4 text-muted-foreground" />
                            <span class="font-medium text-sm flex-1">{{ cat.name }}</span>
                            <span class="text-xs text-muted-foreground">{{ getCategoryItems(cat.name).length }} items</span>
                            <div class="flex items-center gap-1" @click.stop>
                                <button
                                    class="p-1 rounded hover:bg-muted transition-colors"
                                    @click="startRenameCategory(cat.name)"
                                >
                                    <Pencil class="w-3 h-3 text-muted-foreground" />
                                </button>
                                <button
                                    class="p-1 rounded hover:bg-destructive/20 transition-colors"
                                    @click="deleteCategory(cat.name)"
                                >
                                    <Trash2 class="w-3 h-3 text-muted-foreground hover:text-destructive" />
                                </button>
                            </div>
                        </div>

                        <!-- Category Items -->
                        <div v-show="!isCategoryCollapsed(cat.name)" class="p-2 bg-background">
                            <SortableContainer
                                :model-value="getCategoryItems(cat.name)"
                                @update:model-value="onCategoryItemsChange(cat.name, $event)"
                                item-key="id"
                                handle=".drag-handle"
                                ghost-class="dragging-ghost"
                                chosen-class="dragging-chosen"
                                drag-class="dragging-drag"
                                group="config-items"
                                class="flex flex-col gap-2 min-h-[40px]"
                                @start="isDragging = true"
                                @end="onDragEnd"
                            >
                                <template #item="{ element: item }">
                                    <ConfigItemRow
                                        :model-value="item"
                                        :decrypted-key="decryptedKey"
                                        :is-added="isAdded(item)"
                                        :is-modified="isModified(item, originalConfigItems.find(o => o.id === item.id) ?? null)"
                                        :has-secret-manager-configs="hasSecretManagerConfigs"
                                        :categories="categories"
                                        @update:model-value="updateItem($event)"
                                        @delete="deleteItem(item.id)"
                                        @move-to-category="moveItemToCategory(item.id, $event)"
                                    />
                                </template>
                            </SortableContainer>
                            <div v-if="getCategoryItems(cat.name).length === 0" class="text-sm text-muted-foreground text-center py-4">
                                Drag items here or add new items to this category
                            </div>
                        </div>
                    </div>
                </template>
            </SortableContainer>

            <!-- Uncategorized Items (shown at bottom when categories exist) -->
            <div v-if="hasCategories && uncategorizedItems.length > 0" class="space-y-2">
                <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide px-1">
                    Uncategorized
                </div>
                <SortableContainer
                    :model-value="uncategorizedItems"
                    @update:model-value="onUncategorizedChange"
                    item-key="id"
                    handle=".drag-handle"
                    ghost-class="dragging-ghost"
                    chosen-class="dragging-chosen"
                    drag-class="dragging-drag"
                    group="config-items"
                    class="flex flex-col gap-2"
                    @start="isDragging = true"
                    @end="onDragEnd"
                >
                    <template #item="{ element: item }">
                        <ConfigItemRow
                            :model-value="item"
                            :decrypted-key="decryptedKey"
                            :is-added="isAdded(item)"
                            :is-modified="isModified(item, originalConfigItems.find(o => o.id === item.id) ?? null)"
                            :has-secret-manager-configs="hasSecretManagerConfigs"
                            :categories="categories"
                            @update:model-value="updateItem($event)"
                            @delete="deleteItem(item.id)"
                            @move-to-category="moveItemToCategory(item.id, $event)"
                        />
                    </template>
                </SortableContainer>
            </div>

            <!-- All items (no categories) -->
            <SortableContainer
                v-if="!hasCategories"
                :model-value="configItems"
                @update:model-value="onNoCategoryReorder"
                item-key="id"
                handle=".drag-handle"
                ghost-class="dragging-ghost"
                chosen-class="dragging-chosen"
                drag-class="dragging-drag"
                class="flex flex-col gap-2"
                @start="isDragging = true"
                @end="isDragging = false"
            >
                <template #item="{ element: item }">
                    <ConfigItemRow
                        :model-value="item"
                        :decrypted-key="decryptedKey"
                        :is-added="isAdded(item)"
                        :is-modified="isModified(item, originalConfigItems.find(o => o.id === item.id) ?? null)"
                        :has-secret-manager-configs="hasSecretManagerConfigs"
                        :categories="categories"
                        @update:model-value="updateItem($event)"
                        @delete="deleteItem(item.id)"
                        @move-to-category="moveItemToCategory(item.id, $event)"
                    />
                </template>
            </SortableContainer>
        </div>
    </div>

    <!-- Dialogs -->
    <AddItemDialog
        v-model:open="showAddItemDialog"
        :categories="categories"
        @add="handleAddItem"
    />

    <AddCategoryDialog
        v-model:open="showAddCategoryDialog"
        :existing-categories="categories"
        @add="addCategory"
    />

    <RenameCategoryDialog
        v-model:open="showRenameCategoryDialog"
        :category-name="renamingCategoryName"
        @rename="handleRenameCategory"
    />

    <EnvImportDialog
        v-model:open="showEnvImportDialog"
        @import="handleEnvImport"
    />

    <SecretImportDialog
        v-model:open="showSecretImportDialog"
        :project-id="project.id"
        :decrypted-key="decryptedKey"
        @import="handleSecretManagerImport"
    />
</template>
