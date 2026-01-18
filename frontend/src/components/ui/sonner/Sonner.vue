<script setup lang="ts">
import { Toaster as Sonner } from 'vue-sonner'

const props = defineProps<{
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center'
  expand?: boolean
  richColors?: boolean
}>()
</script>

<template>
  <Sonner
    class="toaster group"
    :position="props.position ?? 'bottom-right'"
    :expand="props.expand"
    :rich-colors="props.richColors"
    :toast-options="{
      classes: {
        toast: 'group toast group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg',
        description: 'group-[.toast]:text-muted-foreground',
        actionButton: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
        cancelButton: 'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground',
        error: 'group-[.toaster]:bg-destructive group-[.toaster]:text-destructive-foreground group-[.toaster]:border-destructive',
      },
    }"
  />
</template>
