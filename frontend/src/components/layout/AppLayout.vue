<script setup lang="ts">
import Sidebar from './Sidebar.vue'
</script>

<template>
  <div class="dark flex h-screen w-screen overflow-hidden bg-background text-foreground relative">
    <Sidebar class="z-[60] shrink-0" />
    <main class="flex-1 overflow-y-auto overflow-x-hidden bg-background min-w-0">
      <slot />
    </main>
    <div data-tauri-drag-region class="absolute top-0 left-0 right-0 h-10 w-full z-50 bg-transparent select-none pointer-events-auto"></div>
  </div>
</template>
