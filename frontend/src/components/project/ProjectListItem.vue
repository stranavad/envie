<script setup lang="ts">
import { Folder, ArrowRight } from 'lucide-vue-next';

export interface ProjectListItemData {
    id: string;
    name: string;
    teamName?: string;
    updatedAt: string;
}

defineProps<{
    project: ProjectListItemData;
}>();

defineEmits<{
    click: [id: string];
}>();

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, { dateStyle: 'medium' });
}
</script>

<template>
    <div
        class="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors cursor-pointer group"
        @click="$emit('click', project.id)"
    >
        <div class="flex items-center gap-4">
            <div class="p-2 bg-primary/10 rounded-full text-primary">
                <Folder class="w-5 h-5" />
            </div>
            <div>
                <h3 class="font-medium leading-none">{{ project.name }}</h3>
                <div class="flex items-center gap-2 mt-1">
                    <span v-if="project.teamName" class="text-sm text-muted-foreground">{{ project.teamName }}</span>
                    <span v-if="project.teamName" class="text-muted-foreground">·</span>
                    <span class="text-sm text-muted-foreground">
                        Updated {{ formatDate(project.updatedAt) }}
                    </span>
                </div>
            </div>
        </div>
        <ArrowRight class="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
    </div>
</template>
