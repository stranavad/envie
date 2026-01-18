<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
import { Plus } from 'lucide-vue-next';

const props = defineProps<{
    open: boolean;
    categories: string[];
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'add': [item: { name: string; value: string; sensitive: boolean; category?: string }];
}>();

const newItem = ref({
    name: '',
    value: '',
    sensitive: false,
    category: ''
});

// Reset form when dialog closes
watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        newItem.value = {
            name: '',
            value: '',
            sensitive: false,
            category: ''
        };
    }
});

function handleAdd() {
    if (!newItem.value.name) return;

    emit('add', {
        name: newItem.value.name,
        value: newItem.value.value,
        sensitive: newItem.value.sensitive,
        category: newItem.value.category || undefined
    });

    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Add Config Item</DialogTitle>
                <DialogDescription>
                    Add a new environment variable to your project.
                </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-4">
                <div class="space-y-2">
                    <Label for="itemName">Name</Label>
                    <Input
                        id="itemName"
                        v-model="newItem.name"
                        placeholder="MY_VARIABLE"
                        class="font-mono"
                        @keyup.enter="handleAdd"
                    />
                </div>
                <div class="space-y-2">
                    <Label for="itemValue">Value</Label>
                    <Input
                        id="itemValue"
                        v-model="newItem.value"
                        placeholder="Enter value..."
                        @keyup.enter="handleAdd"
                    />
                </div>
                <div class="space-y-2" v-if="categories.length > 0">
                    <Label for="itemCategory">Category (optional)</Label>
                    <select
                        id="itemCategory"
                        v-model="newItem.category"
                        class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    >
                        <option value="">No category</option>
                        <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                    </select>
                </div>
                <div class="flex items-center space-x-2">
                    <Switch id="itemSensitive" v-model:checked="newItem.sensitive" />
                    <Label for="itemSensitive">Mark as sensitive</Label>
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button @click="handleAdd" :disabled="!newItem.name">
                    <Plus class="w-4 h-4 mr-2" />
                    Add Item
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
