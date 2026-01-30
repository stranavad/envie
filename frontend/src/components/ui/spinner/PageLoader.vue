<script setup lang="ts">
import Spinner from './Spinner.vue';

interface Props {
  message?: string;
}

defineProps<Props>();
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 space-y-4">
    <Spinner size="lg" class="text-primary" />
    <p v-if="message" class="text-sm text-muted-foreground animate-pulse">{{ message }}</p>
  </div>
</template>
