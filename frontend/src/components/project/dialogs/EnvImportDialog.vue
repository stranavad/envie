<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { FileUp } from 'lucide-vue-next';
import { parseEnvString, type ParsedEnvItem } from '@/utils/env-format';

const props = defineProps<{
    open: boolean;
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'import': [items: ParsedEnvItem[], markAsSensitive: boolean];
}>();

const envInput = ref('');
const markNewAsSensitive = ref(true);

// Reset form when dialog closes
watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        envInput.value = '';
        markNewAsSensitive.value = true;
    }
});

function handleImport() {
    const items = parseEnvString(envInput.value);
    if (items.length > 0) {
        emit('import', items, markNewAsSensitive.value);
    }
    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent class="sm:max-w-lg">
            <DialogHeader>
                <DialogTitle>Import from .env</DialogTitle>
                <DialogDescription>
                    Paste your .env file content below. Existing keys will be updated, new keys will be added.
                </DialogDescription>
            </DialogHeader>
            <div class="py-4 space-y-4">
                <Textarea
                    v-model="envInput"
                    placeholder="API_KEY=12345&#10;DB_HOST=localhost&#10;SECRET_TOKEN=abc123"
                    class="font-mono text-sm min-h-[200px]"
                />
                <div class="flex items-center space-x-2">
                    <Switch id="markSensitive" v-model="markNewAsSensitive" />
                    <Label for="markSensitive">Mark new items as sensitive</Label>
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button @click="handleImport" :disabled="!envInput.trim()">
                    <FileUp class="w-4 h-4 mr-2" />
                    Import
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
