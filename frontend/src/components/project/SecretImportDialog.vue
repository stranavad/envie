<script setup lang="ts">
import {ref, watch} from 'vue';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue,} from '@/components/ui/select'
import {Button} from '@/components/ui/button';
import {Checkbox} from '@/components/ui/checkbox';
import {Input} from '@/components/ui/input';
import {Label} from '@/components/ui/label';
import {ScrollArea} from '@/components/ui/scroll-area';
import {Progress} from '@/components/ui/progress'
import {Loader2, Search} from 'lucide-vue-next';
import {type SecretManagerConfig, SecretManagerConfigService} from '@/services/secret-manager-config.service';

import type {ConfigItem} from '@/services/project.service';
import {EncryptionService} from "@/services/encryption.service.ts";
import { useSecretManagerStore } from '@/stores/secret-manager.store';

const props = defineProps<{
    open: boolean;
    projectId: string;
    decryptedKey: string;
}>();

const emit = defineEmits<{
    (e: 'update:open', value: boolean): void;
    (e: 'import', items: Partial<ConfigItem>[]): void;
}>();

const isOpen = ref(props.open);
const isLoadingProviders = ref(false);
const isLoadingSecrets = ref(false);
const isImporting = ref(false);

const providers = ref<SecretManagerConfig[]>([]);
const selectedProviderId = ref<string>('');
const secrets = ref<string[]>([]);
const selectedSecrets = ref<Set<string>>(new Set());
const searchQuery = ref('');
const importProgress = ref(0);
const processedCount = ref(0);
const totalCount = ref(0);

const secretManagerStore = useSecretManagerStore();

watch(() => props.open, (val) => {
    isOpen.value = val;
    if (val) {
        loadProviders();
    } else {
        // Reset state on close
        selectedSecrets.value.clear();
        searchQuery.value = '';
    }
});

watch(isOpen, (val) => {
    emit('update:open', val);
});

watch(selectedProviderId, (val) => {
    if (val) loadSecrets(val);
});

async function loadProviders() {
    isLoadingProviders.value = true;
    try {
        providers.value = await SecretManagerConfigService.getConfigs(props.projectId);
        if (providers.value.length > 0) {
            selectedProviderId.value = providers.value[0].id;
        }
    } catch (e) {
        console.error("Failed to load providers", e);
    } finally {
        isLoadingProviders.value = false;
    }
}

async function loadSecrets(providerId: string) {
    isLoadingSecrets.value = true;
    secrets.value = [];
    selectedSecrets.value.clear();
    
    try {
        const config = providers.value.find(p => p.id === providerId);
        if (!config) return;

        const decryptedJson = await EncryptionService.decryptValue(props.decryptedKey, config.encryptedKey)

        secrets.value = await secretManagerStore.listSecrets(providerId, decryptedJson);
    } catch (e) {
        console.error("Failed to load secrets", e);
    } finally {
        isLoadingSecrets.value = false;
    }
}


function toggleSecret(name: string) {
  if(selectedSecrets.value.has(name)){
    selectedSecrets.value.delete(name)
  } else {
    selectedSecrets.value.add(name)
  }
}

function toggleAll() {
    const visible = filteredSecrets.value;
    const current = selectedSecrets.value;
    const allSelected = visible.every(s => current.has(s));
    
    if (allSelected) {
        visible.forEach(s => selectedSecrets.value.delete(s));
    } else {
        visible.forEach(s => selectedSecrets.value.add(s));
    }
}

const filteredSecrets = ref<string[]>([]);
watch([secrets, searchQuery], () => {
    if (!searchQuery.value) {
        filteredSecrets.value = secrets.value;
    } else {
        const q = searchQuery.value.toLowerCase();
        filteredSecrets.value = secrets.value.filter(s => s.toLowerCase().includes(q));
    }
}, { immediate: true });


async function handleImport() {
    if (selectedSecrets.value.size === 0) return;
    
    isImporting.value = true;
    try {
        const config = providers.value.find(p => p.id === selectedProviderId.value);
        if (!config) return;

        const decryptedJson = await EncryptionService.decryptValue(props.decryptedKey, config.encryptedKey)

        const items: Partial<ConfigItem>[] = [];
        totalCount.value = selectedSecrets.value.size;
        processedCount.value = 0;
        importProgress.value = 0;
        
        // Fetch sequentially to avoid rate limits or overwhelming
        for (const secretName of selectedSecrets.value) {
            try {
                const result = await secretManagerStore.getSecretValue(selectedProviderId.value, decryptedJson, secretName);
                items.push({
                    name: secretName,
                    value: result.value,
                    sensitive: true, // Requested by user
                    secretManagerName: secretName,
                    secretManagerConfigId: selectedProviderId.value,
                    secretManagerVersion: result.version,
                    secretManagerLastSyncAt: new Date().toISOString()
                });
            } catch (e) {
                 console.error(`Failed to fetch value for ${secretName}`, e);
            } finally {
                processedCount.value++;
                importProgress.value = (processedCount.value / totalCount.value) * 100;
            }
        }

        emit('import', items);
        isOpen.value = false;
    } catch (e) {
        console.error("Import failed", e);
    } finally {
        isImporting.value = false;
    }
}

</script>

<template>
    <Dialog v-model:open="isOpen">
        <DialogContent class="sm:max-w-[500px]">
            <DialogHeader>
                <DialogTitle>Import from Secret Manager</DialogTitle>
                <DialogDescription>
                    Select secrets to import as configuration items. They will be added with their current values and linked automatically.
                </DialogDescription>
            </DialogHeader>

            <div class="space-y-4 py-4">
                <!-- Provider Selector -->
                <div v-if="providers.length > 1" class="space-y-2">
                    <Label>Provider</Label>
                     <Select v-model="selectedProviderId">
                        <SelectTrigger>
                            <SelectValue placeholder="Select provider" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem v-for="p in providers" :key="p.id" :value="p.id">
                                {{ p.name }}
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <!-- Search and List -->
                <div class="space-y-2">
                    <div class="flex items-center gap-2">
                        <div class="relative flex-1">
                            <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input v-model="searchQuery"  placeholder="Search secrets..." class="pl-8" />
                        </div>
                        <Button variant="outline" size="sm" @click="toggleAll" :disabled="isLoadingSecrets || filteredSecrets.length === 0">
                             {{ filteredSecrets.every(s => selectedSecrets.has(s)) && filteredSecrets.length > 0 ? 'Deselect All' : 'Select All' }}
                        </Button>
                    </div>

                    <div class="border rounded-md">
                        <div v-if="isLoadingSecrets" class="p-8 flex justify-center items-center text-muted-foreground">
                            <Loader2 class="h-6 w-6 animate-spin mr-2" />
                            Loading secrets...
                        </div>
                        <div v-else-if="secrets.length === 0" class="p-8 text-center text-muted-foreground text-sm">
                            No secrets found in this provider.
                        </div>
                        <ScrollArea v-else class="h-[300px]">
                            <div class="p-4 space-y-2">
                                <div v-for="secret in filteredSecrets" :key="secret" class="flex items-center space-x-2 p-1 hover:bg-muted/50 rounded">
                                    <Checkbox 
                                        :id="secret"
                                        :model-value="selectedSecrets.has(secret)"
                                        @update:model-value="toggleSecret(secret)"
                                    />
                                    <Label :for="secret" class="flex-1 cursor-pointer font-mono text-sm break-all">
                                        {{ secret }}
                                    </Label>
                                </div>
                            </div>
                        </ScrollArea>
                    </div>
                </div>
                
                <div class="text-sm text-muted-foreground text-right">
                    Selected: {{ selectedSecrets.size }}
                </div>
            </div>

            <DialogFooter>
                <div v-if="isImporting" class="w-full flex flex-col gap-2">
                     <div class="flex items-center justify-between text-xs text-muted-foreground">
                        <span>Importing secrets...</span>
                        <span>{{ processedCount }} / {{ totalCount }}</span>
                    </div>
                    <Progress :model-value="importProgress" class="h-2" />
                </div>
                <div v-else class="flex gap-2 justify-end w-full">
                    <Button variant="outline" @click="isOpen = false">Cancel</Button>
                    <Button @click="handleImport" :disabled="selectedSecrets.size === 0">
                        Import Selected
                    </Button>
                </div>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
