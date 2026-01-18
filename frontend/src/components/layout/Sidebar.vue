<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { Home, Settings, ChevronLeft, Menu, ShieldCheck, Building2 } from 'lucide-vue-next';

const router = useRouter();
const auth = useAuthStore();
const isCollapsed = ref(false);
const isAutoCollapsed = ref(false);

const COLLAPSE_BREAKPOINT = 1200; // px - collapse when width is below this

function checkScreenSize() {
    const shouldAutoCollapse = window.innerWidth < COLLAPSE_BREAKPOINT;

    if (shouldAutoCollapse && !isAutoCollapsed.value) {
        // Screen became small - auto collapse
        isCollapsed.value = true;
        isAutoCollapsed.value = true;
    } else if (!shouldAutoCollapse && isAutoCollapsed.value) {
        // Screen became large again - auto expand only if it was auto-collapsed
        isCollapsed.value = false;
        isAutoCollapsed.value = false;
    }
}

function toggleCollapse() {
    isCollapsed.value = !isCollapsed.value;
    // If user manually toggles, clear auto-collapse state
    isAutoCollapsed.value = false;
}

function navigate(path: string) {
    router.push(path);
}

onMounted(() => {
    checkScreenSize();
    window.addEventListener('resize', checkScreenSize);
});

onUnmounted(() => {
    window.removeEventListener('resize', checkScreenSize);
});
</script>

<template>
  <aside
    class="flex h-full flex-col border-r bg-sidebar text-sidebar-foreground transition-all duration-300 ease-in-out"
    :class="isCollapsed ? 'w-16' : 'w-64'"
  >
    <div class="flex h-16 items-center border-b px-4" :class="isCollapsed ? 'justify-center' : 'justify-between'">
      <h1 v-if="!isCollapsed" class="text-2xl font-bold tracking-tight truncate">Envie</h1>
      <Button variant="ghost" size="icon" @click="toggleCollapse" class="h-8 w-8">
        <Menu v-if="isCollapsed" class="h-4 w-4" />
        <ChevronLeft v-else class="h-4 w-4" />
      </Button>
    </div>

    <nav class="flex-1 space-y-2 p-2">
        <!-- Dashboard Item -->
        <Button
            variant="ghost"
            :class="[
                'w-full justify-start cursor-pointer hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed ? 'justify-center px-2' : 'px-4'
            ]"
            @click="navigate('/')"
            :title="isCollapsed ? 'Dashboard' : ''"
        >
            <Home class="h-5 w-5" :class="{ 'mr-2': !isCollapsed }" />
            <span v-if="!isCollapsed">Dashboard</span>
        </Button>

        <Button
            variant="ghost"
            :class="[
                'w-full justify-start cursor-pointer hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed ? 'justify-center px-2' : 'px-4'
            ]"
            @click="navigate('/identities')"
            :title="isCollapsed ? 'Identities' : ''"
        >
            <ShieldCheck class="h-5 w-5" :class="{ 'mr-2': !isCollapsed }" />
            <span v-if="!isCollapsed">Identities</span>
        </Button>

        <!-- Organizations Item -->
        <Button
            variant="ghost"
            :class="[
                'w-full justify-start cursor-pointer hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed ? 'justify-center px-2' : 'px-4'
            ]"
            @click="navigate('/organizations')"
            :title="isCollapsed ? 'Organizations' : ''"
        >
            <Building2 class="h-5 w-5" :class="{ 'mr-2': !isCollapsed }" />
            <span v-if="!isCollapsed">Organizations</span>
        </Button>

        <!-- Settings Item -->
        <Button
            variant="ghost"
            :class="[
                'w-full justify-start cursor-pointer hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                isCollapsed ? 'justify-center px-2' : 'px-4'
            ]"
            @click="navigate('/settings')"
            :title="isCollapsed ? 'Settings' : ''"
        >
            <Settings class="h-5 w-5" :class="{ 'mr-2': !isCollapsed }" />
            <span v-if="!isCollapsed">Settings</span>
        </Button>
    </nav>

    <div class="border-t border-sidebar-border p-4" v-if="auth.user">
      <div class="flex items-center gap-3" :class="{ 'justify-center': isCollapsed }">
        <Avatar>
          <AvatarImage :src="auth.user.avatarUrl" :alt="auth.user.name" />
          <AvatarFallback>{{ auth.user.name.charAt(0).toUpperCase() }}</AvatarFallback>
        </Avatar>
        <div class="text-sm overflow-hidden" v-if="!isCollapsed">
          <p class="font-medium truncate">{{ auth.user.name }}</p>
          <p class="text-xs text-muted-foreground truncate">{{ auth.user.email }}</p>
        </div>
      </div>
    </div>
  </aside>
</template>
