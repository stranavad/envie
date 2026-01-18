<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useOrganizationStore } from '@/stores/organization';
import { Button } from '@/components/ui/button';
import { TabNav } from '@/components/ui/tab-nav';
import { Users, ArrowLeft } from 'lucide-vue-next';
import OrganizationProjects from '@/components/organization/OrganizationProjects.vue';
import OrganizationTeams from '@/components/organization/OrganizationTeams.vue';
import OrganizationUsers from '@/components/organization/OrganizationUsers.vue';
import OrganizationSettings from '@/components/organization/OrganizationSettings.vue';
import { OrganizationService } from '@/services/organization.service';

const router = useRouter();
const route = useRoute();
const store = useOrganizationStore();
const orgId = route.params.id as string;

// Tab state
const activeTab = ref('projects');
const tabs = [
    { key: 'projects', label: 'Projects' },
    { key: 'teams', label: 'Teams' },
    { key: 'users', label: 'Users' },
    { key: 'settings', label: 'Settings' }
];

// Data
const teams = ref<any[]>([]);
const users = ref<any[]>([]);

const organization = computed(() => store.currentOrganization);

const canEditOrg = computed(() => {
    const role = organization.value?.role;
    return role === 'owner' || role === 'Owner' || role === 'admin';
});

onMounted(async () => {
    await store.getOrganization(orgId);
    await Promise.all([
        loadTeams(),
        loadUsers()
    ]);
});

async function loadTeams() {
    try {
        teams.value = await store.fetchTeams(orgId) || [];
    } catch (e) {
        console.error('Failed to load teams', e);
    }
}

async function loadUsers() {
    try {
        users.value = await OrganizationService.getOrganizationUsers(orgId);
    } catch (e) {
        console.error('Failed to load users', e);
    }
}

async function handleTeamsUpdated() {
    await loadTeams();
}

async function handleNameUpdated() {
    await store.getOrganization(orgId);
}
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8">
        <!-- Back Button -->
        <div class="flex items-center gap-4">
            <Button
                variant="ghost"
                class="-ml-2 px-2 text-muted-foreground hover:text-foreground"
                @click="router.push('/organizations')"
            >
                <ArrowLeft class="w-4 h-4 mr-2" />
                Back
            </Button>
        </div>

        <div v-if="organization" class="space-y-6">
            <!-- Header -->
            <div class="flex items-center justify-between">
                <div class="flex flex-col gap-1">
                    <h1 class="text-3xl font-bold tracking-tight">{{ organization.name }}</h1>
                    <div class="flex gap-4 text-sm text-muted-foreground font-mono">
                        <span>ID: {{ orgId }}</span>
                        <span>Role: {{ organization.role }}</span>
                    </div>
                </div>
                <Button variant="outline" size="sm">
                    <Users class="mr-2 h-4 w-4" />
                    Invite Member
                </Button>
            </div>

            <!-- Tabs -->
            <TabNav v-model="activeTab" :tabs="tabs" />

            <!-- Tab Content -->
            <div v-show="activeTab === 'projects'">
                <OrganizationProjects
                    :organization-id="orgId"
                    :teams="teams"
                />
            </div>

            <div v-show="activeTab === 'teams'">
                <OrganizationTeams
                    :organization-id="orgId"
                    :teams="teams"
                    :users="users"
                    :can-edit="canEditOrg"
                    @teams-updated="handleTeamsUpdated"
                />
            </div>

            <div v-show="activeTab === 'users'">
                <OrganizationUsers :users="users" />
            </div>

            <div v-show="activeTab === 'settings'">
                <OrganizationSettings
                    :organization-id="orgId"
                    :organization-name="organization.name"
                    :can-edit="canEditOrg"
                    @name-updated="handleNameUpdated"
                />
            </div>
        </div>

        <div v-else class="h-full flex items-center justify-center">
            <div class="loading loading-spinner loading-lg"></div>
        </div>
    </div>
</template>
