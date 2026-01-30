<script setup lang="ts">
import { Button } from '@/components/ui/button';
import type { Component } from 'vue';

interface Props {
    title: string;
    description?: string;
    actionLabel?: string;
    actionIcon?: Component;
    actionDisabled?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    action: [];
}>();
</script>

<template>
    <div class="flex justify-between items-start gap-4">
        <div class="space-y-1">
            <h3 class="text-lg font-medium">{{ title }}</h3>
            <p v-if="description" class="text-sm text-muted-foreground">
                {{ description }}
            </p>
        </div>
        <div class="flex items-center gap-2 shrink-0">
            <slot name="actions" />
            <Button
                v-if="actionLabel"
                size="sm"
                variant="outline"
                :disabled="actionDisabled"
                @click="emit('action')"
            >
                <component v-if="actionIcon" :is="actionIcon" class="w-4 h-4 mr-2" />
                {{ actionLabel }}
            </Button>
        </div>
    </div>
</template>
