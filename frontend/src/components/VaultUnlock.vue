<script setup lang="ts">
import {ref} from 'vue';
import {useVaultStore} from '../stores/vault';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card';
import {Lock} from 'lucide-vue-next';
import {invoke} from '@tauri-apps/api/core';

const props = defineProps<{
    userName?: string;
}>();

const password = ref('');
const vaultStore = useVaultStore();
const isLoading = ref(false);

async function handleUnlock() {
    isLoading.value = true;
    try {
        await vaultStore.unlockVault(password.value);
    } catch (e) {
        // error handling is done in store but we can clear loading
    } finally {
        isLoading.value = false;
    }
}

async function handleReset() {
    isLoading.value = true;
    try {
        await invoke('nuke_vault', { userId: vaultStore.userId || "" });
        window.location.reload();
    } catch(e) {
        alert("Failed to reset: " + e);
        isLoading.value = false;
    }
}
</script>

<template>
    <div class="fixed inset-0 bg-background flex items-center justify-center p-4 z-50">
        <Card class="w-full max-w-sm shadow-lg">
            <CardHeader class="text-center">
                <div class="mx-auto p-3 bg-primary/10 rounded-full w-fit mb-4">
                    <Lock class="w-6 h-6 text-primary" />
                </div>
                <CardTitle v-if="props.userName">Welcome back, {{ props.userName }}</CardTitle>
                <CardTitle v-else>Unlock Vault</CardTitle>
                <CardDescription>
                    Enter your master password to continue.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <div class="space-y-2">
                    <Input 
                        v-model="password" 
                        type="password" 
                        placeholder="Master Password" 
                        autofocus
                        @keyup.enter="handleUnlock" 
                    />
                </div>
                <div v-if="vaultStore.error" class="text-sm text-destructive font-medium text-center">
                    {{ vaultStore.error }}
                </div>
            </CardContent>
            <CardFooter class="flex flex-col gap-4">
                <Button class="w-full" @click="handleUnlock" :disabled="isLoading || !password">
                    {{ isLoading ? 'Unlocking...' : 'Unlock' }}
                </Button>
                
                <button @click="handleReset" class="text-xs text-muted-foreground hover:text-destructive underline">
                    Forgot Password? Reset Vault
                </button>
            </CardFooter>
        </Card>
    </div>
</template>
