<script setup lang="ts">
import { AlertCircle } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';

interface Props {
  title?: string;
  message?: string;
  retry?: () => void;
}

withDefaults(defineProps<Props>(), {
  title: 'Something went wrong',
});
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 space-y-4">
    <div class="rounded-full bg-destructive/10 p-3">
      <AlertCircle class="h-6 w-6 text-destructive" />
    </div>
    <div class="text-center space-y-1">
      <h3 class="font-medium text-foreground">{{ title }}</h3>
      <p v-if="message" class="text-sm text-muted-foreground max-w-md">{{ message }}</p>
    </div>
    <Button v-if="retry" variant="outline" size="sm" @click="retry">
      Try again
    </Button>
  </div>
</template>
