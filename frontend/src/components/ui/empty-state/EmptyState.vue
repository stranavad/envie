<script setup lang="ts">
import type { Component } from 'vue';

interface Props {
    icon?: Component;
    title: string;
    description?: string;
}

defineProps<Props>();
</script>

<template>
    <div class="text-center py-10 text-muted-foreground border rounded-lg bg-muted/20">
        <component
            v-if="icon"
            :is="icon"
            class="w-12 h-12 mx-auto mb-4 opacity-50"
        />
        <p>{{ title }}</p>
        <p v-if="description" class="text-sm mt-1">{{ description }}</p>
    </div>
</template>
