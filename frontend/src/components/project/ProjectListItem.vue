<script setup lang="ts">
import { Folder, ArrowRight, Download, Upload, AlertTriangle, Loader2 } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import type { SyncStatus } from '@/services/file-mapping.service';
import { type ProjectShort } from "@/services/project.service";


const props = defineProps<{
    project: ProjectShort;
    syncStatus?: SyncStatus;
    isPulling?: boolean;
}>();

const emit = defineEmits<{
    click: [id: string];
    pull: [id: string];
}>();

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, { dateStyle: 'medium' });
}

function handlePullClick(e: Event) {
    e.stopPropagation();
    emit('pull', props.project.id);
}

function getSyncLabel(): string {
    switch (props.syncStatus) {
        case 'remote_changed':
            return 'Remote changed';
        case 'local_changed':
            return 'Local changed';
        case 'both_changed':
            return 'Conflict';
        case 'file_missing':
            return 'File missing';
        default:
            return '';
    }
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
                    <span class="text-sm text-muted-foreground">
                        Updated {{ formatDate(project.updatedAt) }}
                    </span>
                </div>
            </div>
        </div>

        <div class="flex items-center gap-2">
            <!-- Sync Status Indicator -->
            <div v-if="syncStatus && syncStatus !== 'synced' && syncStatus !== 'not_linked'" class="flex items-center gap-2">
                <div class="flex items-center gap-1.5 text-xs font-medium px-2 py-1 rounded-full"
                    :class="{
                        'bg-blue-500/10 text-blue-400': syncStatus === 'remote_changed',
                        'bg-orange-500/10 text-orange-400': syncStatus === 'local_changed',
                        'bg-red-500/10 text-red-400': syncStatus === 'both_changed' || syncStatus === 'file_missing'
                    }"
                >
                    <Download v-if="syncStatus === 'remote_changed'" class="w-3 h-3" />
                    <Upload v-else-if="syncStatus === 'local_changed'" class="w-3 h-3" />
                    <AlertTriangle v-else class="w-3 h-3" />
                    <span>{{ getSyncLabel() }}</span>
                </div>

                <!-- Quick Pull Button (only for remote_changed) -->
                <Button
                    v-if="syncStatus === 'remote_changed'"
                    size="sm"
                    variant="outline"
                    class="h-7 text-xs"
                    :disabled="isPulling"
                    @click="handlePullClick"
                >
                    <Loader2 v-if="isPulling" class="w-3 h-3 mr-1 animate-spin" />
                    <Download v-else class="w-3 h-3 mr-1" />
                    Pull
                </Button>
            </div>

            <ArrowRight class="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
        </div>
    </div>
</template>
