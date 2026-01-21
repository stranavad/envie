<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Button } from '@/components/ui/button';
import { Plus, Loader2 } from 'lucide-vue-next';
import ProjectListItem, { type ProjectListItemData } from '@/components/project/ProjectListItem.vue';
import CreateProjectDialog from '@/components/project/CreateProjectDialog.vue';
import { useOrganizationStore } from '@/stores/organization';
import { ProjectService } from '@/services/project.service';

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

const projects = ref<ProjectListItemData[]>([]);
const isCreateProjectOpen = ref(false);
const isLoading = ref(true);

onMounted(async () => {
    try {
        await loadProjects();
    } finally {
        isLoading.value = false;
    }
});

async function loadProjects() {
    try {
        const allProjects = await ProjectService.getProjects();
        projects.value = allProjects.filter((p) => p.organizationId === props.organizationId);
    } catch (e) {
        console.error('Failed to load projects', e);
    }
}

function openProject(id: string) {
    router.push(`/projects/${id}`);
}

async function handleCreateProject(payload: { name: string; teamId?: string }) {
    if (!payload.name || !payload.teamId) return;

    try {
        await store.createProject(props.organizationId, payload.teamId, payload.name);
        isCreateProjectOpen.value = false;
        await loadProjects();
        emit('project-created');
    } catch (e) {
        console.error('Failed to create project', e);
    }
}

defineExpose({ loadProjects });
</script>

<template>
    <div class="space-y-4">
        <div class="flex justify-between items-center">
            <h3 class="text-lg font-medium">Projects</h3>
            <Button size="sm" @click="isCreateProjectOpen = true">
                <Plus class="mr-2 h-4 w-4" />
                New Project
            </Button>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="flex items-center justify-center py-12">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>

        <div v-else-if="projects.length > 0" class="bg-card rounded-lg border shadow-sm">
            <div class="divide-y divide-border">
                <ProjectListItem
                    v-for="project in projects"
                    :key="project.id"
                    :project="project"
                    @click="openProject"
                />
            </div>
        </div>

        <div v-else class="text-center py-8 text-muted-foreground border rounded-lg bg-muted/20">
            No projects found in this organization.
        </div>

        <CreateProjectDialog
            v-model:open="isCreateProjectOpen"
            :teams="teams"
            @create="handleCreateProject"
        />
    </div>
</template>
