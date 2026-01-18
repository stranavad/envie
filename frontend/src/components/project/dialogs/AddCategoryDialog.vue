<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { FolderPlus } from 'lucide-vue-next';

const props = defineProps<{
    open: boolean;
    existingCategories: string[];
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'add': [name: string];
}>();

const categoryName = ref('');

// Reset form when dialog closes
watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        categoryName.value = '';
    }
});

function handleAdd() {
    const name = categoryName.value.trim();
    if (!name) return;

    // Check if category already exists
    if (props.existingCategories.includes(name)) {
        return;
    }

    emit('add', name);
    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Add Category</DialogTitle>
                <DialogDescription>
                    Create a new category to organize your config items.
                </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-4">
                <div class="space-y-2">
                    <Label for="categoryName">Category Name</Label>
                    <Input
                        id="categoryName"
                        v-model="categoryName"
                        placeholder="e.g., Database, AWS, Stripe"
                        @keyup.enter="handleAdd"
                    />
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button @click="handleAdd" :disabled="!categoryName.trim()">
                    <FolderPlus class="w-4 h-4 mr-2" />
                    Add Category
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
