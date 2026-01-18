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

const props = defineProps<{
    open: boolean;
    categoryName: string;
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'rename': [oldName: string, newName: string];
}>();

const newName = ref('');

// Initialize with current name when dialog opens
watch(() => props.open, (isOpen) => {
    if (isOpen) {
        newName.value = props.categoryName;
    }
});

function handleRename() {
    const name = newName.value.trim();
    if (!name) return;

    emit('rename', props.categoryName, name);
    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Rename Category</DialogTitle>
                <DialogDescription>
                    Enter a new name for the category "{{ categoryName }}".
                </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-4">
                <div class="space-y-2">
                    <Label for="newCategoryName">New Name</Label>
                    <Input
                        id="newCategoryName"
                        v-model="newName"
                        @keyup.enter="handleRename"
                    />
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button @click="handleRename" :disabled="!newName.trim()">
                    Rename
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
