<script setup lang="ts">
import {type ConfigItem} from '@/services/project.service';
import {Button} from '@/components/ui/button';
import {Switch} from '@/components/ui/switch';
import {Label} from '@/components/ui/label';
import {Textarea} from '@/components/ui/textarea';
import {
  Check,
  ChevronsUpDown,
  ClipboardCheck,
  Eye,
  EyeOff,
  FolderInput,
  GripVertical,
  Link as LinkIcon,
  RefreshCw,
  Settings,
  Trash2
} from 'lucide-vue-next';
import {ref, watch} from 'vue';
import {type SecretManagerConfig, SecretManagerConfigService} from '@/services/secret-manager-config.service';

import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue,} from '@/components/ui/select'
import {Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList,} from '@/components/ui/command'
import {Popover, PopoverContent, PopoverTrigger,} from '@/components/ui/popover'
import {EncryptionService} from "@/services/encryption.service.ts";

const props = defineProps<{
    modelValue: ConfigItem;
    decryptedKey: string;
    isAdded?: boolean;
    isModified?: boolean;
    hasSecretManagerConfigs?: boolean;
    categories?: string[];
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: ConfigItem): void;
    (e: 'delete'): void;
    (e: 'move-to-category', category: string | undefined): void;
}>();

const expanded = ref(false)
const show = ref(!props.modelValue.sensitive)
const showCopiedToast = ref(false)

// Secret Manager State
const isLoadingSecrets = ref(false);
const secretConfigs = ref<SecretManagerConfig[]>([]);
const availableSecrets = ref<{ name: string; providerId: string; providerName: string }[]>([]);
const selectedProviderId = ref<string>('');
const selectedSecretName = ref<string>(props.modelValue.secretManagerName || '');
const isSecretLinked = ref(!!props.modelValue.secretManagerName);
const openCombobox = ref(false)

watch(() => props.modelValue.secretManagerName, (val) => {
    isSecretLinked.value = !!val;
    if (val) selectedSecretName.value = val;
});

watch(expanded, async (val) => {
    if (val && secretConfigs.value.length === 0) {
        await loadSecretOptions();
    }
});

// ... imports
import { useSecretManagerStore } from '@/stores/secret-manager.store';

// ...

const secretManagerStore = useSecretManagerStore();

async function loadSecretOptions() {
    isLoadingSecrets.value = true;
    try {
        const configs = await SecretManagerConfigService.getConfigs(props.modelValue.projectId);
        secretConfigs.value = configs;

        const secrets: { name: string; providerId: string; providerName: string }[] = [];

        // Fetch secrets from all providers
        for (const config of configs) {
             try {
               const decryptedJson = await EncryptionService.decryptValue(props.decryptedKey, config.encryptedKey)

                const list = await secretManagerStore.listSecrets(config.id, decryptedJson);
                list.forEach(name => {
                    secrets.push({
                        name,
                        providerId: config.id,
                        providerName: config.name
                    });
                });
             } catch (e) {
                 console.error("Failed to load secrets for provider", config.name, e);
             }
        }
        availableSecrets.value = secrets;

        // Auto-select provider if secret name matches or specific ID is set
        if (props.modelValue.secretManagerConfigId) {
             selectedProviderId.value = props.modelValue.secretManagerConfigId;
        } else if (props.modelValue.secretManagerName) {
            const match = secrets.find(s => s.name === props.modelValue.secretManagerName);
            if (match) {
                selectedProviderId.value = match.providerId;
            } else if (configs.length > 0) {
                 selectedProviderId.value = configs[0].id; // Default
            }
        } else if (configs.length > 0) {
            selectedProviderId.value = configs[0].id;
        }

    } catch (e) {
        console.error("Failed to load secret configs", e);
    } finally {
        isLoadingSecrets.value = false;
    }
}

async function syncSecret() {
    if (!selectedProviderId.value || !selectedSecretName.value) return;

    // Find provider
    const config = secretConfigs.value.find(c => c.id === selectedProviderId.value);
    if (!config) return;

    isLoadingSecrets.value = true;
    try {
        const decryptedJson = await EncryptionService.decryptValue(props.decryptedKey, config.encryptedKey)

        const result = await secretManagerStore.getSecretValue(config.id, decryptedJson, selectedSecretName.value);

        const newItem = {
            ...props.modelValue,
            value: result.value,
            secretManagerName: selectedSecretName.value,
            secretManagerConfigId: selectedProviderId.value,
            secretManagerVersion: result.version,
            secretManagerLastSyncAt: new Date().toISOString()
        };

        emit('update:modelValue', newItem);
    } catch (e) {
        alert("Failed to sync secret: " + e);
    } finally {
        isLoadingSecrets.value = false;
    }
}

function updateSensitive(val: boolean) {
    const newItem = { ...props.modelValue, sensitive: val };
    emit('update:modelValue', newItem);
}

function updateLink(val: boolean) {
    isSecretLinked.value = val;
    if (!val) {
        // Clear link
         const newItem = {
             ...props.modelValue,
             secretManagerName: undefined,
             secretManagerConfigId: undefined,
             secretManagerVersion: undefined,
             secretManagerLastSyncAt: undefined
         };
         emit('update:modelValue', newItem);
    } else {
        if (availableSecrets.value.length === 0) loadSecretOptions();
    }
}


function onValueInput(val: string | undefined) {
    const newItem = { ...props.modelValue, value: val ?? '' };
    emit('update:modelValue', newItem);
}


function copyToClipboard() {
    navigator.clipboard.writeText(props.modelValue.value);
    showCopiedToast.value = true;
    setTimeout(() => {
        showCopiedToast.value = false;
    }, 2000);
}
</script>

<template>
    <div class="border rounded-md bg-card transition-all" :class="{'border-primary/50': expanded}">
        <!-- Header (Always Visible) -->
        <div class="flex items-center gap-3 p-3 hover:bg-muted/50 transition-colors">
            <!-- Drag Handle -->
            <div class="drag-handle cursor-grab active:cursor-grabbing p-1 -m-1 rounded hover:bg-muted transition-colors">
                <GripVertical class="w-4 h-4 text-muted-foreground" />
            </div>

            <!-- Status Badge -->
            <div class="w-2 h-2 rounded-full shrink-0"
                :class="{
                    'bg-green-500': isAdded,
                    'bg-orange-500': isModified,
                    'bg-muted-foreground/30': !isAdded && !isModified
                }"
                :title="isAdded ? 'New Item' : (isModified ? 'Modified' : 'Unchanged')"
            ></div>

            <!-- Key -->
            <div class="w-1/3 min-w-[150px] flex items-center gap-2">
                 <LinkIcon v-if="modelValue.secretManagerName" class="w-3 h-3 text-blue-500" title="Linked to Secret Manager" />
                <p class="font-mono text-sm font-medium truncate" :title="modelValue.name">{{ modelValue.name }}</p>
            </div>

            <!-- Value Preview (Clickable to Copy) -->
            <div class="flex-1 min-w-0 relative">
                <button
                    type="button"
                    @click="copyToClipboard"
                    class="w-full text-left cursor-pointer hover:bg-muted rounded px-2 py-1 -mx-2 -my-1 transition-colors group"
                    :title="show ? 'Click to copy value' : 'Click to copy value (hidden)'"
                >
                    <div v-if="show" class="flex items-center gap-2">
                        <span class="text-sm font-mono truncate">{{ modelValue.value || '(empty)' }}</span>
                        <ClipboardCheck v-if="showCopiedToast" class="w-3 h-3 text-green-500 shrink-0" />
                    </div>
                    <div v-else class="flex items-center gap-2">
                        <span class="tracking-widest select-none text-muted-foreground text-sm">••••••••••••</span>
                        <ClipboardCheck v-if="showCopiedToast" class="w-3 h-3 text-green-500 shrink-0" />
                    </div>
                </button>

                <!-- Toast notification -->
                <Transition
                    enter-active-class="transition-all duration-200 ease-out"
                    enter-from-class="opacity-0 translate-y-1"
                    enter-to-class="opacity-100 translate-y-0"
                    leave-active-class="transition-all duration-150 ease-in"
                    leave-from-class="opacity-100 translate-y-0"
                    leave-to-class="opacity-0 -translate-y-1"
                >
                    <div
                        v-if="showCopiedToast"
                        class="absolute left-0 -top-8 bg-foreground text-background text-xs px-2 py-1 rounded shadow-lg whitespace-nowrap z-10"
                    >
                        Copied to clipboard!
                    </div>
                </Transition>
            </div>

            <!-- Header Actions -->
             <div class="flex items-center gap-1">
                <Button variant="ghost" size="icon" class="h-8 w-8" @click="show = !show" title="Toggle visibility">
                    <Eye v-if="!show" class="w-4 h-4" />
                    <EyeOff v-else class="w-4 h-4" />
                </Button>
                 <!-- Settings / Expand Button -->
                <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8"
                    :class="{'bg-accent text-accent-foreground': expanded}"
                    @click="expanded = !expanded"
                    title="Edit settings"
                >
                    <Settings class="w-4 h-4" />
                </Button>

                 <Button variant="ghost" size="icon" class="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" @click="$emit('delete')">
                    <Trash2 class="w-4 h-4" />
                </Button>
            </div>
        </div>

        <!-- Expanded Body -->
        <div v-if="expanded" class="px-4 pb-4 pt-4 border-t bg-muted/20 space-y-6">

            <!-- Value Editor -->
            <div class="space-y-2">
                <Label class="text-sm">Value</Label>
                <Textarea
                    :model-value="modelValue.value"
                    @update:model-value="onValueInput"
                    class="font-mono text-sm min-h-[80px]"
                    placeholder="Enter value..."
                    :readonly="!!modelValue.secretManagerName"
                />
                <p v-if="modelValue.secretManagerName" class="text-xs text-muted-foreground">
                    Value is managed by Secret Manager. Disable sync to edit manually.
                </p>
            </div>

            <!-- Secret Manager Link -->
            <div class="space-y-3 pt-4 border-t border-border/50">
                 <div class="flex items-center justify-between">
                     <div class="space-y-0.5">
                        <Label class="text-base" :class="{'text-muted-foreground': !hasSecretManagerConfigs}">Secret Manager Sync</Label>
                        <p class="text-xs text-muted-foreground">
                            <template v-if="hasSecretManagerConfigs">
                                Link this value to a Google Secret Manager secret.
                            </template>
                            <template v-else>
                                No secret manager providers configured. Add one in the External Providers tab.
                            </template>
                        </p>
                     </div>
                     <Switch
                         :model-value="isSecretLinked"
                         @update:model-value="updateLink"
                         :disabled="!hasSecretManagerConfigs"
                     />
                 </div>

                   <div v-if="isSecretLinked" class="animate-in fade-in zoom-in-95 duration-200 mt-2">

                      <div v-if="secretConfigs.length === 0 && !isLoadingSecrets" class="text-sm text-yellow-600">No Secret Manager providers configured.</div>

                      <div v-else class="grid gap-3">
                           <!-- Provider Select (Only if multiple) -->
                           <div v-if="secretConfigs.length > 1" class="grid gap-1">
                                <Label class="text-xs">Provider</Label>
                                <Select v-model="selectedProviderId">
                                  <SelectTrigger class="w-full h-9">
                                    <SelectValue placeholder="Select provider" />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem v-for="c in secretConfigs" :key="c.id" :value="c.id">
                                      {{ c.name }}
                                    </SelectItem>
                                  </SelectContent>
                                </Select>
                           </div>

                             <div class="flex items-center justify-between">
                                <Label class="text-xs">Secret Name</Label>
                                <Button variant="ghost" size="sm" class="h-6 text-xs" @click="loadSecretOptions" :disabled="isLoadingSecrets">
                                    <RefreshCw class="w-3 h-3 mr-1" :class="{'animate-spin': isLoadingSecrets}" />
                                    Refresh List
                                </Button>
                             </div>

                           <!-- Secret Search (Combobox) -->
                           <div class="grid gap-1">
                                <div class="flex gap-2">
                                    <Popover v-model:open="openCombobox">
                                        <PopoverTrigger as-child>
                                            <Button
                                                variant="outline"
                                                role="combobox"
                                                :aria-expanded="openCombobox"
                                                class="w-full justify-between h-9 font-normal bg-transparent"
                                                :disabled="isLoadingSecrets"
                                            >
                                                {{ selectedSecretName || "Select secret..." }}
                                                <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                            </Button>
                                        </PopoverTrigger>
                                        <PopoverContent class="w-[300px] p-0" align="start">
                                            <Command>
                                                <CommandInput class="h-9" placeholder="Search secret..." />
                                                <CommandEmpty>No secret found.</CommandEmpty>
                                                <CommandList>
                                                    <CommandGroup>
                                                        <CommandItem
                                                            v-for="s in availableSecrets.filter(x => x.providerId === selectedProviderId)"
                                                            :key="s.name"
                                                            :value="s.name"
                                                            @select="() => {
                                                                selectedSecretName = s.name
                                                                openCombobox = false
                                                            }"
                                                        >
                                                            {{ s.name }}
                                                            <Check
                                                                class="ml-auto h-4 w-4"
                                                                :class="selectedSecretName === s.name ? 'opacity-100' : 'opacity-0'"
                                                            />
                                                        </CommandItem>
                                                    </CommandGroup>
                                                </CommandList>
                                            </Command>
                                        </PopoverContent>
                                    </Popover>

                                     <Button size="sm" @click="syncSecret" :disabled="isLoadingSecrets || !selectedSecretName">
                                        <RefreshCw class="w-3 h-3 mr-2" :class="{'animate-spin': isLoadingSecrets}" />
                                        Sync
                                     </Button>
                                </div>
                                <p v-if="modelValue.secretManagerLastSyncAt" class="text-[10px] text-muted-foreground pt-1">
                                    Last synced: {{ new Date(modelValue.secretManagerLastSyncAt).toLocaleString() }} (v{{ modelValue.secretManagerVersion }})
                                </p>
                                <p v-if="availableSecrets.length === 0 && !isLoadingSecrets" class="text-[10px] text-muted-foreground pt-1">
                                    No secrets found. Check provider connection manually or refresh.
                                </p>
                           </div>
                      </div>
                 </div>
            </div>

             <!-- Sensitive Switch -->
            <div class="flex items-center justify-between pt-4 border-t border-border/50">
                  <div class="space-y-0.5">
                        <Label class="text-sm">Sensitive Value</Label>
                        <p class="text-xs text-muted-foreground">Hide value in UI by default.</p>
                  </div>
                 <Switch
                    :model-value="modelValue.sensitive"
                    @update:model-value="updateSensitive"
                 />
            </div>

            <!-- Category -->
            <div v-if="categories && categories.length > 0" class="flex items-center justify-between pt-4 border-t border-border/50">
                <div class="space-y-0.5">
                    <Label class="text-sm flex items-center gap-2">
                        <FolderInput class="w-4 h-4" />
                        Category
                    </Label>
                    <p class="text-xs text-muted-foreground">Organize this item into a category.</p>
                </div>
                <Select
                    :model-value="modelValue.category || '__none__'"
                    @update:model-value="(val) => emit('move-to-category', val === '__none__' || !val ? undefined : String(val))"
                >
                    <SelectTrigger class="w-[180px] h-9">
                        <SelectValue placeholder="No category" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="__none__">No category</SelectItem>
                        <SelectItem v-for="cat in categories" :key="cat" :value="cat">
                            {{ cat }}
                        </SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs text-muted-foreground pt-2 border-t border-border/50">
                <div class="space-y-1">
                    <p><span class="font-semibold">Created:</span> {{ modelValue.createdAt ? new Date(modelValue.createdAt).toLocaleString() : 'Pending Save' }}</p>
                    <p v-if="modelValue.creator">
                        <span class="font-semibold">By:</span> {{ modelValue.creator.name }}
                        <span class="opacity-70 ml-1">&lt;{{ modelValue.creator.email }}&gt;</span>
                    </p>
                </div>
                <div class="space-y-1">
                    <p><span class="font-semibold">Updated:</span> {{ modelValue.updatedAt ? new Date(modelValue.updatedAt).toLocaleString() : 'Pending Save' }}</p>
                    <p v-if="modelValue.updater">
                        <span class="font-semibold">By:</span> {{ modelValue.updater.name }}
                         <span class="opacity-70 ml-1">&lt;{{ modelValue.updater.email }}&gt;</span>
                    </p>
                </div>
            </div>
        </div>
    </div>
</template>
