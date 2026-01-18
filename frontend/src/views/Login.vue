<template>
  <div class="h-screen w-screen flex flex-col items-center justify-center bg-background text-foreground animate-in fade-in duration-500">
    <Card class="w-full max-w-sm">
      <CardHeader class="flex flex-col items-center space-y-2 text-center">
        <CardTitle class="text-4xl font-bold tracking-tighter sm:text-5xl">Envie</CardTitle>
        <CardDescription>Your project environment manager.</CardDescription>
      </CardHeader>

      <CardContent class="grid gap-4">
        <div v-if="!showCodeInput" class="grid gap-2">
          <Button
            @click="handleLogin"
            :disabled="isLoading"
            class="w-full"
          >
            <svg v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-6.219-8.56"></path></svg>
            <svg v-else class="mr-2 h-4 w-4" role="img" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg"><title>GitHub</title><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
            Sign in with GitHub
          </Button>
        </div>

        <div v-else class="grid gap-4 animate-in fade-in slide-in-from-bottom-2">
          <div class="grid gap-2">
            <label for="code" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
              Enter Linking Code
            </label>
            <p class="text-sm text-muted-foreground">
              Copy the code from your browser after signing in with GitHub.
            </p>
            <Input
              v-model="codeInput"
              id="code"
              type="text"
              placeholder="XXXX-XXXX-XXXX"
              class="text-center font-mono text-lg tracking-widest"
              maxlength="14"
              @keyup.enter="handleVerify"
              @input="formatCode"
            />
          </div>
          <Button
            @click="handleVerify"
            :disabled="isVerifying || !isValidCode"
            class="w-full"
          >
            <svg v-if="isVerifying" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-6.219-8.56"></path></svg>
            Link Device
          </Button>
          <Button
            variant="ghost"
            @click="showCodeInput = false"
            class="w-full"
          >
            Back
          </Button>
        </div>

        <div v-if="error" class="text-sm text-destructive animate-in fade-in text-center">
          {{ error }}
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import {ref, computed} from 'vue';
import {useAuthStore} from '../stores/auth';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from '@/components/ui/card';

const emit = defineEmits<{
    (e: 'success'): void;
}>();

const auth = useAuthStore();
const isLoading = ref(false);
const showCodeInput = ref(false);
const codeInput = ref('');
const isVerifying = ref(false);
const error = ref('');

// Validate linking code format: XXXX-XXXX-XXXX (12 hex chars with dashes)
const isValidCode = computed(() => {
    const code = codeInput.value.trim();
    return /^[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}$/i.test(code);
});

// Auto-format the code as user types
const formatCode = () => {
    // Remove all non-hex characters
    let value = codeInput.value.replace(/[^a-f0-9]/gi, '').toLowerCase();

    // Insert dashes after every 4 characters
    if (value.length > 4) {
        value = value.slice(0, 4) + '-' + value.slice(4);
    }
    if (value.length > 9) {
        value = value.slice(0, 9) + '-' + value.slice(9);
    }

    // Limit to 14 characters (XXXX-XXXX-XXXX)
    codeInput.value = value.slice(0, 14);
};

const handleLogin = async () => {
    isLoading.value = true;
    error.value = '';

    try {
        // Open OAuth flow (no public key needed - we'll register it after vault setup)
        await auth.login();
        showCodeInput.value = true;
    } catch (err: any) {
        error.value = "Failed to open login page: " + err.toString();
    } finally {
        isLoading.value = false;
    }
};

const handleVerify = async () => {
    if (!isValidCode.value) return;

    isVerifying.value = true;
    error.value = '';

    // Exchange linking code for tokens (no public key yet - we'll register it after vault setup)
    const success = await auth.exchangeLinkingCode(codeInput.value);

    if (success) {
        emit('success');
    } else {
        error.value = "Invalid or expired linking code. Please try again.";
    }
    isVerifying.value = false;
};
</script>
