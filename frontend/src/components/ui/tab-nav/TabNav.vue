<script setup lang="ts">
export interface Tab {
    key: string;
    label: string;
}

defineProps<{
    tabs: Tab[];
    modelValue: string;
}>();

defineEmits<{
    'update:modelValue': [value: string];
}>();
</script>

<template>
    <div class="border-b border-border">
        <div class="flex gap-4">
            <button
                v-for="tab in tabs"
                :key="tab.key"
                class="px-4 py-2 text-sm font-medium transition-colors border-b-2"
                :class="modelValue === tab.key
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'"
                @click="$emit('update:modelValue', tab.key)"
            >
                {{ tab.label }}
            </button>
        </div>
    </div>
</template>
