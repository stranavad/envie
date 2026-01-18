<script setup lang="ts">
import { ref } from 'vue';
import { useVaultStore } from '../stores/vault';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Lock } from 'lucide-vue-next';

const password = ref('');
const confirmPassword = ref('');
const error = ref('');
const vaultStore = useVaultStore();
const isLoading = ref(false);

async function handleSetup() {
    error.value = '';
    
    if (password.value.length < 8) {
        error.value = "Password must be at least 8 characters.";
        return;
    }

    if (password.value !== confirmPassword.value) {
        error.value = "Passwords do not match.";
        return;
    }

    isLoading.value = true;
    try {
        await vaultStore.initVault(password.value);
    } catch (e: any) {
        error.value = e.toString();
    } finally {
        isLoading.value = false;
    }
}
</script>

<template>
    <div class="fixed inset-0 bg-background flex items-center justify-center p-4">
        <Card class="w-full max-w-md">
            <CardHeader>
                <div class="flex items-center gap-2 mb-2">
                    <div class="p-2 bg-primary/10 rounded-full">
                        <Lock class="w-6 h-6 text-primary" />
                    </div>
                </div>
                <CardTitle>Welcome to Envie</CardTitle>
                <CardDescription>
                    Create a secure master password to protect your encryption keys. This password will be required every time you open the app.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <div class="space-y-2">
                    <label class="text-sm font-medium">Master Password</label>
                    <Input v-model="password" type="password" placeholder="••••••••" @keyup.enter="handleSetup" />
                </div>
                <div class="space-y-2">
                    <label class="text-sm font-medium">Confirm Password</label>
                    <Input v-model="confirmPassword" type="password" placeholder="••••••••" @keyup.enter="handleSetup" />
                </div>
                <div v-if="error" class="text-sm text-destructive">{{ error }}</div>
                <div v-if="vaultStore.error" class="text-sm text-destructive">{{ vaultStore.error }}</div>
            </CardContent>
            <CardFooter>
                <Button class="w-full" @click="handleSetup" :disabled="isLoading">
                    {{ isLoading ? 'Initializing Vault...' : 'Create Vault' }}
                </Button>
            </CardFooter>
        </Card>
    </div>
</template>
