<script setup lang="ts">
import { computed } from 'vue';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Upload, Eye, Plus, RefreshCw, Loader2 } from 'lucide-vue-next';
import type { ConfigItem } from '@/services/project.service';

export interface LocalItem {
    name: string;
    value: string;
}

const props = defineProps<{
    open: boolean;
    localItems: LocalItem[];
    remoteItems: ConfigItem[];
    isPushing?: boolean;
}>();

const emit = defineEmits<{
    'update:open': [value: boolean];
    'push': [];
    'review': [];
}>();

const changes = computed(() => {
    const added: LocalItem[] = [];
    const updated: LocalItem[] = [];
    const unchanged: LocalItem[] = [];

    for (const local of props.localItems) {
        const remote = props.remoteItems.find(r => r.name === local.name);
        if (!remote) {
            added.push(local);
        } else if (remote.value !== local.value) {
            updated.push(local);
        } else {
            unchanged.push(local);
        }
    }

    return { added, updated, unchanged };
});

const hasChanges = computed(() => {
    return changes.value.added.length > 0 || changes.value.updated.length > 0;
});

function handlePush() {
    emit('push');
}

function handleReview() {
    emit('review');
    emit('update:open', false);
}
</script>

<template>
    <Dialog :open="open" @update:open="$emit('update:open', $event)">
        <DialogContent class="sm:max-w-md">
            <DialogHeader>
                <DialogTitle>Push Local Changes</DialogTitle>
                <DialogDescription>
                    Review the changes from your local .env file before pushing to Envie.
                </DialogDescription>
            </DialogHeader>

            <div class="py-4 space-y-4 min-w-0">
                <!-- Summary -->
                <div v-if="hasChanges" class="space-y-2">
                    <div v-if="changes.added.length > 0" class="flex items-center gap-2 text-sm">
                        <Plus class="w-4 h-4 text-green-500" />
                        <span><strong>{{ changes.added.length }}</strong> item{{ changes.added.length !== 1 ? 's' : '' }} to add</span>
                    </div>
                    <div v-if="changes.updated.length > 0" class="flex items-center gap-2 text-sm">
                        <RefreshCw class="w-4 h-4 text-blue-500" />
                        <span><strong>{{ changes.updated.length }}</strong> item{{ changes.updated.length !== 1 ? 's' : '' }} to update</span>
                    </div>
                    <div v-if="changes.unchanged.length > 0" class="flex items-center gap-2 text-sm text-muted-foreground">
                        <span>{{ changes.unchanged.length }} unchanged</span>
                    </div>
                </div>

                <div v-else class="text-sm text-muted-foreground text-center py-4">
                    No changes detected. Your local file matches the remote config.
                </div>

                <!-- Items preview -->
                <div v-if="hasChanges" class="border rounded-lg divide-y max-h-[200px] overflow-y-auto overflow-x-hidden">
                    <div
                        v-for="item in changes.added"
                        :key="'add-' + item.name"
                        class="px-3 py-2 flex items-center gap-2 bg-green-500/5"
                    >
                        <Plus class="w-3 h-3 text-green-500 shrink-0" />
                        <span class="font-mono text-sm truncate flex-1 min-w-0">{{ item.name }}</span>
                    </div>
                    <div
                        v-for="item in changes.updated"
                        :key="'update-' + item.name"
                        class="px-3 py-2 flex items-center gap-2 bg-blue-500/5"
                    >
                        <RefreshCw class="w-3 h-3 text-blue-500 shrink-0" />
                        <span class="font-mono text-sm truncate flex-1 min-w-0">{{ item.name }}</span>
                    </div>
                </div>
            </div>

            <DialogFooter class="flex-col sm:flex-row gap-2">
                <Button variant="outline" @click="$emit('update:open', false)" :disabled="isPushing">
                    Cancel
                </Button>
                <Button variant="outline" @click="handleReview" :disabled="isPushing || !hasChanges">
                    <Eye class="w-4 h-4 mr-2" />
                    Review
                </Button>
                <Button @click="handlePush" :disabled="isPushing || !hasChanges">
                    <Loader2 v-if="isPushing" class="w-4 h-4 mr-2 animate-spin" />
                    <Upload v-else class="w-4 h-4 mr-2" />
                    {{ isPushing ? 'Pushing...' : 'Push' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
