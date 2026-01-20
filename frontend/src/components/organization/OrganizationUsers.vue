<script setup lang="ts">
import { ref } from 'vue';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Trash2, Loader2 } from 'lucide-vue-next';
import { OrganizationService } from '@/services/organization.service';
import { EncryptionService } from '@/services/encryption.service';
import { useOrganizationStore } from '@/stores/organization';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/lib/toast';

export interface OrgUser {
    id: string;
    name: string;
    email: string;
    avatarUrl?: string;
    publicKey?: string;
    role: string;
}

const props = defineProps<{
    organizationId: string;
    users: OrgUser[];
    canEdit: boolean;
}>();

const emit = defineEmits<{
    'users-updated': [];
}>();

const store = useOrganizationStore();
const authStore = useAuthStore();

const updatingUserId = ref<string | null>(null);
const removingUserId = ref<string | null>(null);
const showRemoveDialog = ref(false);
const userToRemove = ref<OrgUser | null>(null);
const error = ref('');

async function handleRoleChange(user: OrgUser, newRole: string) {
    if (user.role === newRole) return;

    updatingUserId.value = user.id;
    error.value = '';

    try {
        let encryptedOrganizationKey: string | undefined;

        if (user.role === 'member' && (newRole === 'admin' || newRole === 'owner')) {
            if (!user.publicKey) {
                throw new Error('User has not set up encryption keys');
            }

            const orgKey = await store.unlockOrganization(props.organizationId);
            if (!orgKey) {
                throw new Error('Failed to access organization key');
            }

            encryptedOrganizationKey = await EncryptionService.encryptKey(user.publicKey, orgKey);
        }

        await OrganizationService.updateOrganizationMember(props.organizationId, user.id, {
            role: newRole,
            encryptedOrganizationKey
        });

        toast.success(`${user.name}'s role updated to ${newRole}`);
        emit('users-updated');
    } catch (e: any) {
        console.error('Failed to update role', e);
        error.value = e.message || 'Failed to update role';
    } finally {
        updatingUserId.value = null;
    }
}

function confirmRemove(user: OrgUser) {
    userToRemove.value = user;
    showRemoveDialog.value = true;
}

async function handleRemove() {
    if (!userToRemove.value) return;

    removingUserId.value = userToRemove.value.id;
    error.value = '';

    try {
        const removedUserName = userToRemove.value.name;
        await OrganizationService.removeOrganizationMember(props.organizationId, userToRemove.value.id);
        showRemoveDialog.value = false;
        userToRemove.value = null;
        toast.success(`${removedUserName} has been removed from the organization`);
        emit('users-updated');
    } catch (e: any) {
        console.error('Failed to remove member', e);
        error.value = e.message || 'Failed to remove member';
    } finally {
        removingUserId.value = null;
    }
}

function isCurrentUser(userId: string): boolean {
    return authStore.user?.id === userId;
}
</script>

<template>
    <div class="space-y-4">
        <h3 class="text-lg font-medium">Users</h3>

        <div v-if="error" class="text-sm text-destructive p-3 bg-destructive/10 rounded-md">
            {{ error }}
        </div>

        <div v-if="users.length > 0" class="border rounded-lg divide-y">
            <div
                v-for="user in users"
                :key="user.id"
                class="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
            >
                <div class="flex items-center gap-4">
                    <Avatar class="h-10 w-10">
                        <AvatarImage :src="user.avatarUrl || ''" />
                        <AvatarFallback>{{ user.name?.[0] }}</AvatarFallback>
                    </Avatar>
                    <div>
                        <div class="font-medium flex items-center gap-2">
                            {{ user.name }}
                            <span v-if="isCurrentUser(user.id)" class="text-xs text-muted-foreground">(you)</span>
                        </div>
                        <div class="text-sm text-muted-foreground">{{ user.email }}</div>
                    </div>
                </div>
                <div class="flex items-center gap-2">
                    <!-- Role selector for editable users -->
                    <template v-if="canEdit">
                        <div class="relative">
                            <Loader2 v-if="updatingUserId === user.id" class="h-4 w-4 animate-spin absolute right-10 top-1/2 -translate-y-1/2 z-10" />
                            <Select
                                :model-value="user.role"
                                @update:model-value="(val) => val && handleRoleChange(user, String(val))"
                                :disabled="updatingUserId === user.id"
                            >
                                <SelectTrigger class="w-[100px] h-8 text-xs">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="member">Member</SelectItem>
                                    <SelectItem value="admin">Admin</SelectItem>
                                    <SelectItem value="owner">Owner</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            class="h-8 w-8 text-muted-foreground hover:text-destructive"
                            @click="confirmRemove(user)"
                            :disabled="removingUserId === user.id"
                        >
                            <Loader2 v-if="removingUserId === user.id" class="h-4 w-4 animate-spin" />
                            <Trash2 v-else class="h-4 w-4" />
                        </Button>
                    </template>
                    <!-- Read-only role badge -->
                    <span v-else class="text-sm px-2 py-1 bg-secondary rounded-md text-secondary-foreground font-medium uppercase text-xs">
                        {{ user.role }}
                    </span>
                </div>
            </div>
        </div>

        <div v-else class="p-8 text-center text-muted-foreground border rounded-lg bg-muted/20">
            No users loaded.
        </div>

        <!-- Remove Confirmation Dialog -->
        <Dialog v-model:open="showRemoveDialog">
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Remove Member</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to remove <strong>{{ userToRemove?.name }}</strong> from this organization?
                        They will also be removed from all teams within this organization.
                    </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                    <Button variant="outline" @click="showRemoveDialog = false">Cancel</Button>
                    <Button
                        variant="destructive"
                        @click="handleRemove"
                    >
                        Remove
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
</template>
