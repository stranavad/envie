<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { TabNav } from '@/components/ui/tab-nav';
import { ArrowLeft } from 'lucide-vue-next';
import { PageLoader } from '@/components/ui/spinner';
import { ErrorState } from '@/components/ui/error-state';
import OrganizationProjects from '@/components/organization/OrganizationProjects.vue';
import OrganizationTeams from '@/components/organization/OrganizationTeams.vue';
import OrganizationUsers from '@/components/organization/OrganizationUsers.vue';
import OrganizationSettings from '@/components/organization/OrganizationSettings.vue';
import { useOrganization, useOrganizationUsers, useTeams, queryKeys } from '@/queries';
import { useQueryClient } from '@tanstack/vue-query';

const router = useRouter();
const route = useRoute();
const queryClient = useQueryClient();
const orgId = route.params.id as string;

const activeTab = ref('projects');
const tabs = [
    { key: 'projects', label: 'Projects' },
    { key: 'teams', label: 'Teams' },
    { key: 'users', label: 'Users' },
    { key: 'settings', label: 'Settings' }
];

// TanStack Queries
const { data: organization, isLoading: orgLoading, error: orgError, refetch: refetchOrg } = useOrganization(orgId);
const { data: users, isLoading: usersLoading } = useOrganizationUsers(orgId);
const { data: teams, isLoading: teamsLoading } = useTeams(orgId);

const isLoading = computed(() => orgLoading.value || usersLoading.value || teamsLoading.value);

const canEditOrg = computed(() => {
    const role = organization.value?.role;
    return role === 'owner' || role === 'Owner' || role === 'admin';
});

function handleTeamsUpdated() {
    queryClient.invalidateQueries({ queryKey: queryKeys.teams(orgId) });
}

function handleNameUpdated() {
    queryClient.invalidateQueries({ queryKey: queryKeys.organization(orgId) });
}

function handleUsersUpdated() {
    queryClient.invalidateQueries({ queryKey: queryKeys.organizationUsers(orgId) });
}

const errorMessage = computed(() => {
    if (orgError.value) {
        return orgError.value instanceof Error ? orgError.value.message : String(orgError.value);
    }
    return '';
});
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

        <!-- Loading State -->
        <PageLoader v-if="isLoading" message="Loading organization..." />

        <ErrorState
            v-else-if="errorMessage"
            title="Failed to load organization"
            :message="errorMessage"
            :retry="refetchOrg"
        />

        <div v-else-if="organization" class="space-y-6">
            <!-- Header -->
            <div class="flex flex-col gap-1">
                <h1 class="text-3xl font-bold tracking-tight">{{ organization.name }}</h1>
                <div class="flex gap-4 text-sm text-muted-foreground font-mono">
                    <span>ID: {{ orgId }}</span>
                    <span>Role: {{ organization.role }}</span>
                </div>
            </div>

            <!-- Tabs -->
            <TabNav v-model="activeTab" :tabs="tabs" />

            <!-- Tab Content -->
            <div v-show="activeTab === 'projects'">
                <OrganizationProjects
                    :organization-id="orgId"
                    :teams="teams || []"
                />
            </div>

            <div v-show="activeTab === 'teams'">
                <OrganizationTeams
                    :organization-id="orgId"
                    :teams="teams || []"
                    :users="users || []"
                    :can-edit="canEditOrg"
                    @teams-updated="handleTeamsUpdated"
                />
            </div>

            <div v-show="activeTab === 'users'">
                <OrganizationUsers
                    :organization-id="orgId"
                    :users="users || []"
                    :can-edit="canEditOrg"
                    @users-updated="handleUsersUpdated"
                />
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

        <div v-else class="text-center py-20 text-muted-foreground">
            Organization not found.
        </div>
    </div>
</template>
