<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Users, Plus, Shield, ChevronDown, ChevronRight, Loader2 } from 'lucide-vue-next';
import { ProjectService, type ProjectDetail, type ProjectAccessData } from '@/services/project.service';
import { TeamService } from '@/services/team.service';
import { EncryptionService } from '@/services/encryption.service';
import { IdentityService } from '@/services/identity.service';
import { useOrganizationStore } from '@/stores/organization';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
}>();

const orgStore = useOrganizationStore();

const accessData = ref<ProjectAccessData | null>(null);
const isLoading = ref(false);
const error = ref('');
const selectedTeamId = ref('');
const isAddingTeam = ref(false);
const addTeamError = ref('');
const expandedTeams = ref<Set<string>>(new Set());

function toggleTeam(teamId: string) {
    if (expandedTeams.value.has(teamId)) {
        expandedTeams.value.delete(teamId);
    } else {
        expandedTeams.value.add(teamId);
    }
    expandedTeams.value = new Set(expandedTeams.value);
}

async function loadAccessData() {
    isLoading.value = true;
    error.value = '';
    try {
        accessData.value = await ProjectService.getProjectTeams(props.project.id);
    } catch (e: any) {
        error.value = 'Failed to load access data: ' + (e.message || e);
    } finally {
        isLoading.value = false;
    }
}

async function handleAddTeam() {
    if (!selectedTeamId.value || !accessData.value) return;

    isAddingTeam.value = true;
    addTeamError.value = '';

    try {
        // Get the team's encrypted key to decrypt the team key
        const teams = await TeamService.getTeams(props.project.organizationId);
        const team = teams.find((t) => t.id === selectedTeamId.value);

        if (!team) {
            throw new Error('Team not found');
        }

        let teamKey = '';

        const masterKeyPair = IdentityService.getMasterKeyPair();
        if (!masterKeyPair) {
            throw new Error('Master Identity not loaded');
        }

        // Try to decrypt team key via user's encrypted team key
        if (team.userEncryptedKey) {
            teamKey = await EncryptionService.decryptKey(masterKeyPair.privateKey, team.userEncryptedKey);
        }

        // Fallback: Use org key
        if (!teamKey) {
            const orgKey = await orgStore.unlockOrganization(props.project.organizationId);
            if (orgKey && team.encryptedKey) {
                teamKey = await EncryptionService.decryptValue(orgKey, team.encryptedKey);
            }
        }

        if (!teamKey) {
            throw new Error('Unable to decrypt team key');
        }

        // Encrypt project key with team key
        const encryptedProjectKey = await EncryptionService.encryptValue(teamKey, props.decryptedKey);

        await ProjectService.addTeamToProject(props.project.id, {
            teamId: selectedTeamId.value,
            encryptedProjectKey: encryptedProjectKey
        });

        selectedTeamId.value = '';
        await loadAccessData();
    } catch (e: any) {
        addTeamError.value = 'Failed to add team: ' + (e.message || e);
    } finally {
        isAddingTeam.value = false;
    }
}

function getRoleBadgeClass(role: string): string {
    if (role === 'owner' || role === 'Owner') return 'bg-primary text-primary-foreground';
    if (role === 'admin') return 'bg-secondary text-secondary-foreground';
    return 'bg-muted text-muted-foreground';
}

onMounted(() => {
    loadAccessData();
});
</script>

<template>
    <div class="space-y-6">
        <div v-if="error" class="bg-destructive/15 text-destructive p-4 rounded-md">
            {{ error }}
        </div>

        <div v-if="isLoading" class="flex flex-col items-center py-12 text-muted-foreground">
            <Loader2 class="h-8 w-8 animate-spin mb-4" />
            <p>Loading access information...</p>
        </div>

        <div v-else-if="accessData" class="space-y-6">
            <!-- Teams with Access -->
            <Card>
                <CardHeader>
                    <CardTitle class="flex items-center gap-2">
                        <Users class="w-5 h-5" />
                        Teams with Access
                    </CardTitle>
                    <CardDescription>
                        Teams that can access this project and their members.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div v-if="accessData.teams.length === 0" class="text-center py-6 text-muted-foreground">
                        No teams have access to this project yet.
                    </div>

                    <div v-else class="space-y-2">
                        <div v-for="team in accessData.teams" :key="team.id" class="border rounded-lg">
                            <button
                                @click="toggleTeam(team.id)"
                                class="w-full flex items-center justify-between p-4 hover:bg-muted/50 transition-colors text-left"
                            >
                                <div class="flex items-center gap-3">
                                    <div class="p-2 bg-primary/10 rounded-full">
                                        <Users class="w-4 h-4 text-primary" />
                                    </div>
                                    <div>
                                        <div class="font-medium">{{ team.name }}</div>
                                        <div class="text-sm text-muted-foreground">
                                            {{ team.memberCount }} member{{ team.memberCount !== 1 ? 's' : '' }}
                                        </div>
                                    </div>
                                </div>
                                <ChevronDown
                                    v-if="expandedTeams.has(team.id)"
                                    class="w-5 h-5 text-muted-foreground"
                                />
                                <ChevronRight
                                    v-else
                                    class="w-5 h-5 text-muted-foreground"
                                />
                            </button>

                            <div v-if="expandedTeams.has(team.id)" class="border-t px-4 pb-4">
                                <div class="pt-3 space-y-3">
                                    <div
                                        v-for="user in team.users"
                                        :key="user.id"
                                        class="flex items-center justify-between py-2"
                                    >
                                        <div class="flex items-center gap-3">
                                            <Avatar class="h-8 w-8">
                                                <AvatarImage :src="user.avatarUrl" :alt="user.name" />
                                                <AvatarFallback>{{ user.name.charAt(0).toUpperCase() }}</AvatarFallback>
                                            </Avatar>
                                            <div>
                                                <div class="font-medium text-sm">{{ user.name }}</div>
                                                <div class="text-xs text-muted-foreground">{{ user.email }}</div>
                                            </div>
                                        </div>
                                        <span
                                            class="text-xs font-medium px-2 py-1 rounded-md"
                                            :class="getRoleBadgeClass(user.role)"
                                        >
                                            {{ user.role }}
                                        </span>
                                    </div>
                                    <div v-if="team.users.length === 0" class="text-sm text-muted-foreground py-2">
                                        No members in this team.
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <!-- Organization Admins -->
            <Card>
                <CardHeader>
                    <CardTitle class="flex items-center gap-2">
                        <Shield class="w-5 h-5" />
                        Organization Admins
                    </CardTitle>
                    <CardDescription>
                        Organization owners and admins have implicit access to all projects.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div class="space-y-3">
                        <div
                            v-for="user in accessData.organizationAdmins"
                            :key="user.id"
                            class="flex items-center justify-between py-2"
                        >
                            <div class="flex items-center gap-3">
                                <Avatar class="h-8 w-8">
                                    <AvatarImage :src="user.avatarUrl" :alt="user.name" />
                                    <AvatarFallback>{{ user.name.charAt(0).toUpperCase() }}</AvatarFallback>
                                </Avatar>
                                <div>
                                    <div class="font-medium text-sm">{{ user.name }}</div>
                                    <div class="text-xs text-muted-foreground">{{ user.email }}</div>
                                </div>
                            </div>
                            <span
                                class="text-xs font-medium px-2 py-1 rounded-md"
                                :class="getRoleBadgeClass(user.role)"
                            >
                                {{ user.role }}
                            </span>
                        </div>
                        <div v-if="accessData.organizationAdmins.length === 0" class="text-sm text-muted-foreground py-2">
                            No organization admins found.
                        </div>
                    </div>
                </CardContent>
            </Card>

            <!-- Add Team -->
            <Card v-if="project.canEdit && accessData.availableTeams.length > 0">
                <CardHeader>
                    <CardTitle class="flex items-center gap-2">
                        <Plus class="w-5 h-5" />
                        Add Team Access
                    </CardTitle>
                    <CardDescription>
                        Grant additional teams access to this project.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div class="flex gap-3">
                        <Select v-model="selectedTeamId">
                            <SelectTrigger class="w-[280px]">
                                <SelectValue placeholder="Select a team" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem
                                    v-for="team in accessData.availableTeams"
                                    :key="team.id"
                                    :value="team.id"
                                >
                                    {{ team.name }} ({{ team.memberCount }} members)
                                </SelectItem>
                            </SelectContent>
                        </Select>
                        <Button
                            @click="handleAddTeam"
                            :disabled="!selectedTeamId || isAddingTeam"
                        >
                            <Plus class="w-4 h-4 mr-2" />
                            {{ isAddingTeam ? 'Adding...' : 'Add Team' }}
                        </Button>
                    </div>
                    <div v-if="addTeamError" class="text-sm text-destructive mt-2">
                        {{ addTeamError }}
                    </div>
                </CardContent>
            </Card>

            <div v-else-if="project.canEdit && accessData.availableTeams.length === 0" class="text-sm text-muted-foreground">
                All teams in this organization already have access to this project.
            </div>
        </div>
    </div>
</template>
