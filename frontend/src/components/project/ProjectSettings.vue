<script setup lang="ts">
import {ref, watch} from 'vue';
import {useRouter} from 'vue-router';
import {type ProjectDetail, ProjectService} from '@/services/project.service';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Card, CardContent, CardDescription, CardHeader, CardTitle,} from '@/components/ui/card';
import KeyRotation from './KeyRotation.vue';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
    teamKey: string;
}>();

const emit = defineEmits<{
    (e: 'projectUpdated', project: ProjectDetail): void;
    (e: 'rotated'): void;
}>();

const router = useRouter();
const error = ref('');
const success = ref('');

// General Settings State
const editName = ref(props.project.name);
const isSaving = ref(false);

watch(() => props.project, (newVal) => {
    editName.value = newVal.name;
});

async function handleUpdateName() {
    if (!editName.value || editName.value === props.project.name) return;

    isSaving.value = true;
    error.value = '';
    success.value = '';
    
    try {
        await ProjectService.updateProject(props.project.id, editName.value);
        // Create updated project object to emit
        const updatedProject = { ...props.project, name: editName.value };
        emit('projectUpdated', updatedProject);
        success.value = "Project name updated.";
    } catch (err: any) {
        error.value = "Failed to update project: " + err.toString();
    } finally {
        isSaving.value = false;
    }
}

async function handleDelete() {
    try {
        await ProjectService.deleteProject(props.project.id);
        await router.push('/');
    } catch (err: any) {
        error.value = "Failed to delete: " + err.toString();
    }
}
</script>

<template>
    <div class="space-y-6">
        <!-- GENERAL SETTINGS -->
        <Card>
            <CardHeader>
                <CardTitle>General Settings</CardTitle>
                <CardDescription>
                    Manage general project information.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <div class="grid gap-2">
                    <label class="text-sm font-medium">Project Name</label>
                    <div class="flex gap-2">
                        <Input v-model="editName" class="max-w-md" @keyup.enter="handleUpdateName"/>
                        <Button @click="handleUpdateName" :disabled="isSaving || editName === project.name">
                            {{ isSaving ? 'Saving...' : 'Save' }}
                        </Button>
                    </div>
                </div>
                <div v-if="success" class="text-sm text-green-600 font-medium">
                    {{ success }}
                </div>
            </CardContent>
        </Card>

        <!-- KEY ROTATION -->
        <KeyRotation
            :project="project"
            :decrypted-key="decryptedKey"
            :team-key="teamKey"
            @rotated="emit('rotated')"
        />

        <div v-if="error" class="text-destructive text-sm">{{ error }}</div>

        <!-- DANGER ZONE -->
        <Card class="border-destructive/50">
            <CardHeader>
                <CardTitle class="text-destructive">Danger Zone</CardTitle>
                <CardDescription>
                    Irreversible actions for this project.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <div class="flex items-center justify-between p-4 border border-destructive/20 rounded-md bg-destructive/5">
                    <div class="space-y-1">
                        <div class="font-medium text-destructive">Delete Project</div>
                        <div class="text-sm text-muted-foreground">Once you delete a project, there is no going back. Please be certain.</div>
                    </div>
                    <Button variant="destructive" @click="handleDelete">Delete Project</Button>
                </div>
            </CardContent>
        </Card>
    </div>
</template>
