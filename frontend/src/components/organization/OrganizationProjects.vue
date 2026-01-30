<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { Plus } from 'lucide-vue-next';
import { SectionHeader } from '@/components/ui/section-header';
import { PageLoader } from '@/components/ui/spinner';
import { ErrorState } from '@/components/ui/error-state';
import { EmptyState } from '@/components/ui/empty-state';
import ProjectListItem from '@/components/project/ProjectListItem.vue';
import CreateProjectDialog from '@/components/project/CreateProjectDialog.vue';
import { useOrganizationStore } from '@/stores/organization';
import { useOrganizationProjects, queryKeys } from '@/queries';
import { useQueryClient } from '@tanstack/vue-query';

export interface Team {
    id: string;
    name: string;
}

const props = defineProps<{
    organizationId: string;
    teams: Team[];
}>();

const emit = defineEmits<{
    'project-created': [];
}>();

const router = useRouter();
const store = useOrganizationStore();
const queryClient = useQueryClient();

// TanStack Query for organization projects
const { data: projects, isLoading, error: queryError, refetch } = useOrganizationProjects(computed(() => props.organizationId));

const isCreateProjectOpen = ref(false);

function openProject(id: string) {
    router.push(`/projects/${id}`);
}

async function handleCreateProject(payload: { name: string; teamId?: string }) {
    if (!payload.name || !payload.teamId) return;

    try {
        await store.createProject(props.organizationId, payload.teamId, payload.name);
        isCreateProjectOpen.value = false;
        queryClient.invalidateQueries({ queryKey: queryKeys.organizationProjects(props.organizationId) });
        emit('project-created');
    } catch (e) {
        console.error('Failed to create project', e);
    }
}
</script>

<template>
    <div class="space-y-4">
        <SectionHeader
            title="Projects"
            action-label="New Project"
            :action-icon="Plus"
            @action="isCreateProjectOpen = true"
        />

        <!-- Error State -->
        <ErrorState
            v-if="queryError"
            title="Failed to load projects"
            :message="queryError instanceof Error ? queryError.message : String(queryError)"
            :retry="refetch"
        />

        <!-- Loading State -->
        <PageLoader v-else-if="isLoading" message="Loading projects..." />

        <div v-else-if="projects && projects.length > 0" class="bg-card rounded-lg border shadow-sm">
            <div class="divide-y divide-border">
                <ProjectListItem
                    v-for="project in projects"
                    :key="project.id"
                    :project="project"
                    @click="openProject"
                />
            </div>
        </div>

        <EmptyState
            v-else-if="!isLoading && !queryError"
            title="No projects found in this organization"
        />

        <CreateProjectDialog
            v-model:open="isCreateProjectOpen"
            :teams="teams"
            @create="handleCreateProject"
        />
    </div>
</template>
