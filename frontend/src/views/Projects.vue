<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { type Project, ProjectService } from '@/services/project.service';
import { KeyRotationService, type PendingRotationWithProject } from '@/services/key-rotation.service';
import { FileMappingService, type SyncStatus } from '@/services/file-mapping.service';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Building2, ArrowRight, KeyRound, AlertTriangle, Loader2 } from 'lucide-vue-next';
import ProjectListItem from '@/components/project/ProjectListItem.vue';
import { useProjectDecryption } from '@/composables/useProjectDecryption';
import { useFileSync } from '@/composables/useFileSync';

const router = useRouter();
const { decryptProjectKeys } = useProjectDecryption();
const { pullToLocal } = useFileSync();
const projects = ref<Project[]>([]);
const pendingRotations = ref<PendingRotationWithProject[]>([]);
const syncStatusMap = ref<Record<string, SyncStatus>>({});
const pullingMap = ref<Record<string, boolean>>({});
const isLoading = ref(false);
const error = ref('');

// Group projects by organization
const projectsByOrg = computed(() => {
    const grouped: Record<string, { orgName: string; orgId: string; projects: Project[] }> = {};

    for (const project of projects.value) {
        const orgId = project.organizationId;
        if (!grouped[orgId]) {
            grouped[orgId] = {
                orgId,
                orgName: project.organizationName,
                projects: []
            };
        }
        grouped[orgId].projects.push(project);
    }

    return Object.values(grouped);
});

async function loadProjects() {
    isLoading.value = true;
    error.value = '';
    try {
        const [projectsData, rotationsData] = await Promise.all([
            ProjectService.getProjects(),
            KeyRotationService.getUserPendingRotations().catch(() => [])
        ]);
        projects.value = projectsData;
        pendingRotations.value = rotationsData;

        // Load sync statuses for linked projects
        await loadSyncStatuses(projectsData);
    } catch (err: any) {
        error.value = 'Failed to load projects: ' + err.toString();
    } finally {
        isLoading.value = false;
    }
}

async function loadSyncStatuses(projectsList: Project[]) {
    try {
        const mappings = await FileMappingService.getAllMappings();
        const projectMap = new Map(projectsList.map(p => [p.id, p]));
        const statusMap: Record<string, SyncStatus> = {};

        for (const mapping of mappings) {
            const project = projectMap.get(mapping.projectId);
            if (!project) continue;

            const result = await FileMappingService.checkSyncStatus(
                mapping.projectId,
                project.configChecksum || ''
            );

            statusMap[mapping.projectId] = result.status;
        }

        syncStatusMap.value = statusMap;
    } catch (e) {
        console.error('Failed to load sync statuses', e);
    }
}

async function handlePull(projectId: string) {
    const project = projects.value.find(p => p.id === projectId);
    if (!project) return;

    pullingMap.value[projectId] = true;

    try {
        // Get project details to access encrypted keys
        const projectDetail = await ProjectService.getProject(projectId);

        // Decrypt project key
        const { projectKey } = await decryptProjectKeys({
            teamId: projectDetail.teamId,
            organizationId: projectDetail.organizationId,
            encryptedTeamKey: projectDetail.encryptedTeamKey,
            encryptedProjectKey: projectDetail.encryptedProjectKey,
        });

        // Pull using composable
        await pullToLocal(projectId, projectKey, projectDetail.configChecksum || '');

        syncStatusMap.value[projectId] = 'synced';
    } catch (e: any) {
        console.error('Pull failed', e);
        error.value = `Pull failed for ${project.name}: ${e.message || e.toString()}`;
    } finally {
        pullingMap.value[projectId] = false;
    }
}

function navigateToProject(id: string) {
    router.push(`/projects/${id}`);
}

function navigateToOrganization(orgId: string) {
    router.push(`/organizations/${orgId}`);
}

function navigateToOrganizations() {
    router.push('/organizations');
}

loadProjects();
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8 relative">
        <div class="flex items-center justify-between">
            <div>
                <h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
                <p class="text-muted-foreground">Your recent projects across all organizations.</p>
            </div>

            <Button variant="outline" @click="navigateToOrganizations">
                <Building2 class="w-4 h-4 mr-2" />
                View Organizations
            </Button>
        </div>

        <!-- Pending Rotations Alert -->
        <Card v-if="pendingRotations.length > 0" class="border-orange-500/50 bg-orange-500/5">
            <CardHeader class="pb-3">
                <CardTitle class="flex items-center gap-2 text-orange-700">
                    <KeyRound class="w-5 h-5" />
                    Key Rotations Awaiting Your Approval
                </CardTitle>
                <CardDescription>
                    The following projects have pending key rotations that need your verification and approval.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div class="space-y-3">
                    <div
                        v-for="rotation in pendingRotations"
                        :key="rotation.id"
                        class="flex items-center justify-between p-3 bg-background rounded-lg border"
                    >
                        <div class="flex items-center gap-3">
                            <AlertTriangle class="w-5 h-5 text-orange-500" />
                            <div>
                                <p class="font-medium">{{ rotation.project?.name || 'Unknown Project' }}</p>
                                <p class="text-sm text-muted-foreground">
                                    Initiated by {{ rotation.initiator.name }} · Expires {{ new Date(rotation.expiresAt).toLocaleString() }}
                                </p>
                            </div>
                        </div>
                        <Button size="sm" @click="navigateToProject(rotation.projectId)">
                            Review
                            <ArrowRight class="w-4 h-4 ml-2" />
                        </Button>
                    </div>
                </div>
            </CardContent>
        </Card>

        <div v-if="error" class="bg-destructive/15 text-destructive p-4 rounded-md">
            {{ error }}
        </div>

        <div v-if="isLoading" class="flex flex-col items-center py-12 text-muted-foreground">
            <Loader2 class="h-8 w-8 animate-spin mb-4" />
            <p>Loading projects...</p>
        </div>

        <div v-else-if="projects.length === 0" class="bg-card rounded-lg border shadow-sm p-8 text-center">
            <div class="flex flex-col items-center gap-4">
                <div class="p-4 bg-primary/10 rounded-full">
                    <Building2 class="w-8 h-8 text-primary" />
                </div>
                <div class="space-y-2">
                    <h3 class="text-lg font-medium">No projects yet</h3>
                    <p class="text-muted-foreground max-w-md">
                        Projects are created within organizations. Create or join an organization to get started.
                    </p>
                </div>
                <Button @click="navigateToOrganizations">
                    <Building2 class="w-4 h-4 mr-2" />
                    Go to Organizations
                </Button>
            </div>
        </div>

        <div v-else class="space-y-8">
            <!-- Projects grouped by organization -->
            <div v-for="org in projectsByOrg" :key="org.orgId" class="space-y-4">
                <div class="flex items-center justify-between">
                    <div class="flex items-center gap-3">
                        <div class="p-2 bg-primary/10 rounded-full">
                            <Building2 class="w-4 h-4 text-primary" />
                        </div>
                        <h2 class="text-lg font-semibold">{{ org.orgName }}</h2>
                        <span class="text-sm text-muted-foreground">
                            ({{ org.projects.length }} project{{ org.projects.length !== 1 ? 's' : '' }})
                        </span>
                    </div>
                    <Button variant="ghost" size="sm" @click="navigateToOrganization(org.orgId)">
                        View All
                        <ArrowRight class="w-4 h-4 ml-2" />
                    </Button>
                </div>

                <div class="bg-card rounded-lg border shadow-sm">
                    <div class="divide-y divide-border">
                        <ProjectListItem
                            v-for="project in org.projects"
                            :key="project.id"
                            :project="project"
                            :sync-status="syncStatusMap[project.id]"
                            :is-pulling="pullingMap[project.id]"
                            @click="navigateToProject"
                            @pull="handlePull"
                        />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
