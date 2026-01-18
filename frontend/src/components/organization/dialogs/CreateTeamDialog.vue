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
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'create': [name: string];
}>();

const teamName = ref('');
const isCreating = ref(false);

watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        teamName.value = '';
        isCreating.value = false;
    }
});

async function handleCreate() {
    if (!teamName.value.trim()) return;

    isCreating.value = true;
    emit('create', teamName.value.trim());
}

function handleCreated() {
    isCreating.value = false;
    emit('update:open', false);
}

defineExpose({ handleCreated });
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Create Team</DialogTitle>
                <DialogDescription>Create a new team to group members and projects.</DialogDescription>
            </DialogHeader>
            <div class="py-4">
                <Label for="teamName">Team Name</Label>
                <Input
                    id="teamName"
                    v-model="teamName"
                    placeholder="e.g. Backend"
                    class="mt-2"
                    @keyup.enter="handleCreate"
                />
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button @click="handleCreate" :disabled="isCreating || !teamName.trim()">
                    {{ isCreating ? 'Creating...' : 'Create Team' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
