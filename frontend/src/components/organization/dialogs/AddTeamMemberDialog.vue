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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';

export interface OrgUser {
    id: string;
    name: string;
    email: string;
    publicKey?: string;
}

export interface TeamMember {
    userId: string;
}

const props = defineProps<{
    open: boolean;
    users: OrgUser[];
    currentMembers: TeamMember[];
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'add': [userId: string, role: string];
}>();

const selectedUserId = ref('');
const selectedRole = ref('member');
const isAdding = ref(false);
const error = ref('');

watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        selectedUserId.value = '';
        selectedRole.value = 'member';
        isAdding.value = false;
        error.value = '';
    }
});

const availableUsers = computed(() => {
    const memberIds = new Set(props.currentMembers.map(m => m.userId));
    return props.users.filter(u => !memberIds.has(u.id));
});

async function handleAdd() {
    if (!selectedUserId.value) return;

    isAdding.value = true;
    error.value = '';

    emit('add', selectedUserId.value, selectedRole.value);
}

function handleSuccess() {
    isAdding.value = false;
    emit('update:open', false);
}

function handleError(message: string) {
    isAdding.value = false;
    error.value = message;
}

defineExpose({ handleSuccess, handleError });
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent>
            <DialogHeader>
                <DialogTitle>Add Team Member</DialogTitle>
                <DialogDescription>
                    Select an organization member to add to this team.
                </DialogDescription>
            </DialogHeader>
            <div class="py-4 space-y-4">
                <div class="space-y-2">
                    <Label>Select User</Label>
                    <Select v-model="selectedUserId">
                        <SelectTrigger>
                            <SelectValue placeholder="Choose a user..." />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem
                                v-for="user in availableUsers"
                                :key="user.id"
                                :value="user.id"
                            >
                                {{ user.name }} ({{ user.email }})
                            </SelectItem>
                        </SelectContent>
                    </Select>
                    <p v-if="availableUsers.length === 0" class="text-sm text-muted-foreground">
                        All organization members are already in this team.
                    </p>
                </div>
                <div class="space-y-2">
                    <Label>Role</Label>
                    <Select v-model="selectedRole">
                        <SelectTrigger>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="member">Member</SelectItem>
                            <SelectItem value="admin">Admin</SelectItem>
                            <SelectItem value="owner">Owner</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div v-if="error" class="text-sm text-destructive">
                    {{ error }}
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button
                    @click="handleAdd"
                    :disabled="isAdding || !selectedUserId"
                >
                    {{ isAdding ? 'Adding...' : 'Add Member' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
