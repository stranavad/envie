<script setup lang="ts">
import { ref, watch, computed } from 'vue';
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Search, AlertCircle, CheckCircle2, UserX } from 'lucide-vue-next';
import { OrganizationService, type SearchUserResult } from '@/services/organization.service';
import { EncryptionService } from '@/services/encryption.service';
import { useOrganizationStore } from '@/stores/organization';

const props = defineProps<{
    open: boolean;
    organizationId: string;
    currentMemberIds: string[];
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'member-added': [];
}>();

const store = useOrganizationStore();

const email = ref('');
const searchedUser = ref<SearchUserResult | null>(null);
const selectedRole = ref('member');
const isSearching = ref(false);
const isAdding = ref(false);
const searchError = ref('');
const addError = ref('');

watch(() => props.open, (isOpen) => {
    if (!isOpen) {
        email.value = '';
        searchedUser.value = null;
        selectedRole.value = 'member';
        isSearching.value = false;
        isAdding.value = false;
        searchError.value = '';
        addError.value = '';
    }
});

const userStatus = computed(() => {
    if (!searchedUser.value) return null;

    if (props.currentMemberIds.includes(searchedUser.value.id)) {
        return { type: 'already-member', message: 'User is already a member of this organization' };
    }

    if (!searchedUser.value.publicKey) {
        return { type: 'no-key', message: 'User has not set up encryption keys yet' };
    }

    return { type: 'ready', message: 'Ready to add' };
});

const canAdd = computed(() => {
    return searchedUser.value && userStatus.value?.type === 'ready' && !isAdding.value;
});

async function handleSearch() {
    if (!email.value.trim()) return;

    isSearching.value = true;
    searchError.value = '';
    searchedUser.value = null;
    addError.value = '';

    try {
        searchedUser.value = await OrganizationService.searchUserByEmail(email.value.trim());
    } catch (e: any) {
        searchError.value = e.message || 'User not found';
    } finally {
        isSearching.value = false;
    }
}

async function handleAdd() {
    if (!searchedUser.value || !canAdd.value) return;

    isAdding.value = true;
    addError.value = '';

    try {
        let encryptedOrganizationKey: string | undefined;

        if (selectedRole.value === 'admin' || selectedRole.value === 'owner') {
            // Unlock organization to get the master key
            const orgKey = await store.unlockOrganization(props.organizationId);
            if (!orgKey) {
                throw new Error('Failed to access organization key. You may not have permission.');
            }

            // Encrypt the org master key for the target user
            encryptedOrganizationKey = await EncryptionService.encryptKey(
                searchedUser.value.publicKey!,
                orgKey
            );
        }

        await OrganizationService.addOrganizationMember(props.organizationId, {
            userId: searchedUser.value.id,
            role: selectedRole.value,
            encryptedOrganizationKey
        });

        emit('member-added');
        emit('update:open', false);
    } catch (e: any) {
        console.error('Failed to add member', e);
        addError.value = e.message || 'Failed to add member';
    } finally {
        isAdding.value = false;
    }
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent class="sm:max-w-md">
            <DialogHeader>
                <DialogTitle>Add Organization Member</DialogTitle>
                <DialogDescription>
                    Search for a user by email to add them to this organization.
                </DialogDescription>
            </DialogHeader>
            <div class="py-4 space-y-4">
                <!-- Email Search -->
                <div class="space-y-2">
                    <Label>Email Address</Label>
                    <div class="flex gap-2">
                        <Input
                            v-model="email"
                            type="email"
                            placeholder="user@example.com"
                            @keyup.enter="handleSearch"
                        />
                        <Button
                            variant="secondary"
                            @click="handleSearch"
                            :disabled="isSearching || !email.trim()"
                        >
                            <Search class="h-4 w-4" />
                        </Button>
                    </div>
                    <p v-if="searchError" class="text-sm text-destructive flex items-center gap-1">
                        <UserX class="h-4 w-4" />
                        {{ searchError }}
                    </p>
                </div>

                <!-- Found User Display -->
                <div v-if="searchedUser" class="border rounded-lg p-4 space-y-3">
                    <div class="flex items-center gap-3">
                        <Avatar class="h-12 w-12">
                            <AvatarImage :src="searchedUser.avatarUrl || ''" />
                            <AvatarFallback>{{ searchedUser.name?.[0] }}</AvatarFallback>
                        </Avatar>
                        <div class="flex-1 min-w-0">
                            <div class="font-medium truncate">{{ searchedUser.name }}</div>
                            <div class="text-sm text-muted-foreground truncate">{{ searchedUser.email }}</div>
                        </div>
                    </div>

                    <!-- Status Indicator -->
                    <div
                        v-if="userStatus"
                        class="flex items-center gap-2 text-sm py-2 px-3 rounded-md"
                        :class="{
                            'bg-destructive/10 text-destructive': userStatus.type === 'already-member' || userStatus.type === 'no-key',
                            'bg-green-500/10 text-green-600 dark:text-green-400': userStatus.type === 'ready'
                        }"
                    >
                        <AlertCircle v-if="userStatus.type !== 'ready'" class="h-4 w-4" />
                        <CheckCircle2 v-else class="h-4 w-4" />
                        {{ userStatus.message }}
                    </div>

                    <!-- Role Selector -->
                    <div v-if="userStatus?.type === 'ready'" class="space-y-2">
                        <Label>Role</Label>
                        <Select v-model="selectedRole">
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="member">
                                    <div class="flex flex-col items-start">
                                        <span>Member</span>
                                        <span class="text-xs text-muted-foreground">Can access teams they're added to</span>
                                    </div>
                                </SelectItem>
                                <SelectItem value="admin">
                                    <div class="flex flex-col items-start">
                                        <span>Admin</span>
                                        <span class="text-xs text-muted-foreground">Can manage teams and members</span>
                                    </div>
                                </SelectItem>
                                <SelectItem value="owner">
                                    <div class="flex flex-col items-start">
                                        <span>Owner</span>
                                        <span class="text-xs text-muted-foreground">Full organization access</span>
                                    </div>
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div v-if="addError" class="text-sm text-destructive">
                    {{ addError }}
                </div>
            </div>
            <DialogFooter>
                <Button variant="outline" @click="$emit('update:open', false)">Cancel</Button>
                <Button
                    @click="handleAdd"
                    :disabled="!canAdd"
                >
                    {{ isAdding ? 'Adding...' : 'Add Member' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
