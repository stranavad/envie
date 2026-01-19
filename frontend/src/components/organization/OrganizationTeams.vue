<script setup lang="ts">
import { ref, computed } from 'vue';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {
    Plus,
    Users,
    ChevronDown,
    ChevronRight,
    UserPlus,
    Trash2,
    Loader2
} from 'lucide-vue-next';
import CreateTeamDialog from './dialogs/CreateTeamDialog.vue';
import AddTeamMemberDialog, { type OrgUser } from './dialogs/AddTeamMemberDialog.vue';
import { useOrganizationStore } from '@/stores/organization';
import { EncryptionService } from '@/services/encryption.service';
import { TeamService, type TeamMember } from '@/services/team.service';

export interface Team {
    id: string;
    name: string;
    encryptedKey: string;
    projectCount: number;
    memberCount: number;
    previewUsers?: { id: string; name: string; avatarUrl?: string }[];
}

const props = defineProps<{
    organizationId: string;
    teams: Team[];
    users: OrgUser[];
    canEdit: boolean;
}>();

const emit = defineEmits<{
    'teams-updated': [];
}>();

const store = useOrganizationStore();
const createTeamDialogRef = ref<InstanceType<typeof CreateTeamDialog>>();
const addMemberDialogRef = ref<InstanceType<typeof AddTeamMemberDialog>>();

// Dialog state
const isCreateTeamOpen = ref(false);
const isAddMemberOpen = ref(false);
const addMemberTeamId = ref<string | null>(null);

// Expandable team state
const expandedTeamId = ref<string | null>(null);
const teamMembers = ref<Record<string, TeamMember[]>>({});
const isLoadingMembers = ref<Record<string, boolean>>({});

const currentTeamMembers = computed(() => {
    if (!addMemberTeamId.value) return [];
    return teamMembers.value[addMemberTeamId.value] || [];
});

async function toggleTeamExpanded(teamId: string) {
    if (expandedTeamId.value === teamId) {
        expandedTeamId.value = null;
    } else {
        expandedTeamId.value = teamId;
        if (!teamMembers.value[teamId]) {
            await loadTeamMembers(teamId);
        }
    }
}

async function loadTeamMembers(teamId: string) {
    isLoadingMembers.value[teamId] = true;
    try {
        teamMembers.value[teamId] = await TeamService.getTeamMembers(teamId);
    } catch (e) {
        console.error('Failed to load team members', e);
        teamMembers.value[teamId] = [];
    } finally {
        isLoadingMembers.value[teamId] = false;
    }
}

async function handleCreateTeam(name: string) {
    try {
        await store.createTeam(props.organizationId, name);
        createTeamDialogRef.value?.handleCreated();
        emit('teams-updated');
    } catch (e) {
        console.error('Failed to create team', e);
    }
}

function openAddMemberDialog(teamId: string) {
    addMemberTeamId.value = teamId;
    isAddMemberOpen.value = true;
}

async function handleAddMember(userId: string, role: string) {
    if (!addMemberTeamId.value) return;

    try {
        const team = props.teams.find(t => t.id === addMemberTeamId.value);
        if (!team) throw new Error('Team not found');

        const targetUser = props.users.find(u => u.id === userId);
        if (!targetUser?.publicKey) throw new Error('User has no public key');

        // Decrypt the team key using org key
        const orgKey = await store.unlockOrganization(props.organizationId);
        if (!orgKey) throw new Error('Failed to unlock organization');

        const teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);

        // Encrypt team key for the target user
        const encryptedTeamKey = await EncryptionService.encryptKey(targetUser.publicKey, teamKey);

        await TeamService.addTeamMember(addMemberTeamId.value, {
            userId,
            encryptedTeamKey,
            role
        });

        await loadTeamMembers(addMemberTeamId.value);
        addMemberDialogRef.value?.handleSuccess();
        emit('teams-updated');
    } catch (e: any) {
        console.error('Failed to add member', e);
        addMemberDialogRef.value?.handleError(e.message || 'Failed to add member');
    }
}

async function handleRemoveMember(teamId: string, userId: string) {
    try {
        await TeamService.removeTeamMember(teamId, userId);
        await loadTeamMembers(teamId);
        emit('teams-updated');
    } catch (e: any) {
        console.error('Failed to remove member', e);
        alert(e.message || 'Failed to remove member');
    }
}

async function handleUpdateMemberRole(teamId: string, userId: string, newRole: string) {
    try {
        await TeamService.updateTeamMember(teamId, userId, { role: newRole });
        await loadTeamMembers(teamId);
    } catch (e: any) {
        console.error('Failed to update role', e);
        alert(e.message || 'Failed to update role');
    }
}
</script>

<template>
    <div class="space-y-4">
        <div class="flex justify-between items-center">
            <h3 class="text-lg font-medium">Teams</h3>
            <Button size="sm" @click="isCreateTeamOpen = true">
                <Plus class="mr-2 h-4 w-4" />
                New Team
            </Button>
        </div>

        <div class="space-y-3">
            <div v-for="team in teams" :key="team.id" class="border rounded-lg bg-card overflow-hidden">
                <!-- Team Header (Clickable) -->
                <div
                    class="flex items-center justify-between p-4 cursor-pointer hover:bg-muted/50 transition-colors"
                    @click="toggleTeamExpanded(team.id)"
                >
                    <div class="flex items-center gap-4">
                        <component
                            :is="expandedTeamId === team.id ? ChevronDown : ChevronRight"
                            class="h-4 w-4 text-muted-foreground"
                        />
                        <div class="bg-primary/10 p-2 rounded-full">
                            <Users class="h-4 w-4 text-primary" />
                        </div>
                        <div>
                            <div class="font-medium">{{ team.name }}</div>
                            <div class="text-xs text-muted-foreground">
                                {{ team.projectCount }} Projects • {{ team.memberCount }} Members
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center gap-3">
                        <div class="flex -space-x-2">
                            <Avatar
                                v-for="u in team.previewUsers?.slice(0, 3)"
                                :key="u.id"
                                class="h-8 w-8 border-2 border-background"
                            >
                                <AvatarImage :src="u.avatarUrl || ''" />
                                <AvatarFallback>{{ u.name?.[0] }}</AvatarFallback>
                            </Avatar>
                            <div
                                v-if="team.memberCount > 3"
                                class="h-8 w-8 rounded-full border-2 border-background bg-muted flex items-center justify-center text-xs font-medium"
                            >
                                +{{ team.memberCount - 3 }}
                            </div>
                        </div>
                        <Button
                            v-if="canEdit"
                            variant="ghost"
                            size="sm"
                            @click.stop="openAddMemberDialog(team.id)"
                        >
                            <UserPlus class="h-4 w-4" />
                        </Button>
                    </div>
                </div>

                <!-- Expanded Member List -->
                <div v-if="expandedTeamId === team.id" class="border-t">
                    <div v-if="isLoadingMembers[team.id]" class="flex items-center justify-center py-8 text-muted-foreground">
                        <Loader2 class="h-5 w-5 animate-spin mr-2" />
                        Loading members...
                    </div>
                    <div v-else-if="!teamMembers[team.id] || teamMembers[team.id].length === 0" class="py-6 text-center text-sm text-muted-foreground">
                        No members in this team yet.
                    </div>
                    <div v-else class="divide-y">
                        <div
                            v-for="member in teamMembers[team.id]"
                            :key="member.userId"
                            class="flex items-center justify-between p-3 hover:bg-muted/30 transition-colors"
                        >
                            <div class="flex items-center gap-3">
                                <Avatar class="h-8 w-8">
                                    <AvatarImage :src="member.avatarUrl || ''" />
                                    <AvatarFallback>{{ member.name?.[0] }}</AvatarFallback>
                                </Avatar>
                                <div>
                                    <div class="text-sm font-medium">{{ member.name }}</div>
                                    <div class="text-xs text-muted-foreground">{{ member.email }}</div>
                                </div>
                            </div>
                            <div class="flex items-center gap-2">
                                <Select
                                    v-if="canEdit"
                                    :model-value="member.role"
                                    @update:model-value="(val) => val && handleUpdateMemberRole(team.id, member.userId, String(val))"
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
                                <span v-else class="text-xs px-2 py-1 bg-secondary rounded-md text-secondary-foreground font-medium uppercase">
                                    {{ member.role }}
                                </span>
                                <Button
                                    v-if="canEdit"
                                    variant="ghost"
                                    size="icon"
                                    class="h-8 w-8 text-muted-foreground hover:text-destructive"
                                    @click="handleRemoveMember(team.id, member.userId)"
                                >
                                    <Trash2 class="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </div>
                    <!-- Add Member Button inside expanded section -->
                    <div v-if="canEdit" class="p-3 border-t bg-muted/20">
                        <Button
                            variant="outline"
                            size="sm"
                            class="w-full"
                            @click="openAddMemberDialog(team.id)"
                        >
                            <UserPlus class="h-4 w-4 mr-2" />
                            Add Member
                        </Button>
                    </div>
                </div>
            </div>

            <div v-if="teams.length === 0" class="text-sm text-muted-foreground border rounded-lg p-8 text-center bg-muted/20">
                No teams found. Create one to get started.
            </div>
        </div>

        <CreateTeamDialog
            ref="createTeamDialogRef"
            v-model:open="isCreateTeamOpen"
            @create="handleCreateTeam"
        />

        <AddTeamMemberDialog
            ref="addMemberDialogRef"
            v-model:open="isAddMemberOpen"
            :users="users"
            :current-members="currentTeamMembers"
            @add="handleAddMember"
        />
    </div>
</template>
