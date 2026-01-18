<script setup lang="ts">
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

export interface OrgUser {
    id: string;
    name: string;
    email: string;
    avatarUrl?: string;
    role: string;
}

defineProps<{
    users: OrgUser[];
}>();
</script>

<template>
    <div class="space-y-4">
        <h3 class="text-lg font-medium">Users</h3>

        <div v-if="users.length > 0" class="border rounded-lg divide-y">
            <div
                v-for="user in users"
                :key="user.id"
                class="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
            >
                <div class="flex items-center gap-4">
                    <Avatar class="h-10 w-10">
                        <AvatarImage :src="user.avatarUrl || ''" />
                        <AvatarFallback>{{ user.name?.[0] }}</AvatarFallback>
                    </Avatar>
                    <div>
                        <div class="font-medium">{{ user.name }}</div>
                        <div class="text-sm text-muted-foreground">{{ user.email }}</div>
                    </div>
                </div>
                <div class="text-sm px-2 py-1 bg-secondary rounded-md text-secondary-foreground font-medium uppercase text-xs">
                    {{ user.role }}
                </div>
            </div>
        </div>

        <div v-else class="p-8 text-center text-muted-foreground border rounded-lg bg-muted/20">
            No users loaded.
        </div>
    </div>
</template>
