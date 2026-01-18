<script setup lang="ts">
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {AlertTriangle, Check, Edit, Loader2, Trash2, X} from 'lucide-vue-next';
import type {SecretManagerConfig} from '@/services/secret-manager-config.service';
import {ref} from "vue";

defineProps<{
    config: SecretManagerConfig;
    status: 'pending' | 'success' | 'error' | 'unknown';
}>();

const isDeleting = ref(false);

const emit = defineEmits<{
    (e: 'edit'): void;
    (e: 'delete'): void;
}>();
</script>

<template>
    <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
            <div class="flex items-center space-x-2">
                <CardTitle class="text-base font-semibold">{{ config.name }}</CardTitle>
                
                <div v-if="status === 'pending'" class="flex items-center" title="Checking connection...">
                    <Loader2 class="w-3 h-3 text-muted-foreground animate-spin" />
                </div>
                <div v-else-if="status === 'success'" class="w-2.5 h-2.5 rounded-full bg-green-500 border border-green-600" title="Connected"></div>
                <div v-else-if="status === 'error'" class="w-2.5 h-2.5 rounded-full bg-red-500 border border-red-600" title="Connection Error"></div>
            </div>

            <div class="flex space-x-2">
                <template v-if="isDeleting">
                    <span class="text-xs text-muted-foreground mr-2 flex items-center">
                        <AlertTriangle class="w-4 h-4 text-orange-500 mr-1" />
                        Confirm deletion?
                    </span>
                    <Button variant="ghost" size="icon" class="text-green-600 hover:text-green-700 hover:bg-green-50" @click="emit('delete')">
                        <Check class="w-4 h-4" />
                    </Button>
                    <Button variant="ghost" size="icon"  @click="isDeleting = false">
                        <X class="w-4 h-4" />
                    </Button>
                </template>
                <template v-else>
                    <Button variant="ghost" size="icon" @click="emit('edit')">
                        <Edit class="w-4 h-4" />
                    </Button>
                    <Button variant="ghost" size="icon" class="text-muted-foreground hover:text-red-500" @click="isDeleting = true">
                        <Trash2 class="w-4 h-4" />
                    </Button>
                </template>
            </div>
        </CardHeader>
        <CardContent>
            <div class="text-xs text-muted-foreground mt-1">
                Created by {{ config.createdBy?.name || config.createdBy?.email || 'Unknown' }} on {{ new Date(config.createdAt).toLocaleDateString() }}
            </div>
            <div v-if="status === 'error'" class="text-xs text-destructive mt-1">
                Connection check failed. Check if key is valid/active.
            </div>
        </CardContent>
    </Card>
</template>
