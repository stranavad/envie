<script setup lang="ts">
import { cn } from '@/lib/utils';
import { Loader2 } from 'lucide-vue-next';
import type { HTMLAttributes } from 'vue';

interface Props {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  class?: HTMLAttributes['class'];
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
});

const sizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
  xl: 'h-12 w-12',
};
</script>

<template>
  <Loader2 :class="cn('animate-spin text-muted-foreground', sizeClasses[size], props.class)" />
</template>
