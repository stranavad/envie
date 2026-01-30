<script setup lang="ts">
import { ref, computed } from 'vue';
import { type ProjectDetail, ProjectService } from '@/services/project.service';
import { Button } from '@/components/ui/button';
import { SectionHeader } from '@/components/ui/section-header';
import { Plus, Trash2, Copy, Key, AlertTriangle } from 'lucide-vue-next';
import { PageLoader } from '@/components/ui/spinner';
import { ErrorState } from '@/components/ui/error-state';
import { EmptyState } from '@/components/ui/empty-state';
import CreateTokenDialog from './dialogs/CreateTokenDialog.vue';
import { toast } from 'vue-sonner';
import { useProjectTokens, queryKeys } from '@/queries';
import { useQueryClient } from '@tanstack/vue-query';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
}>();

const queryClient = useQueryClient();

// TanStack Query for tokens
const { data: tokens, isLoading, error: queryError, refetch } = useProjectTokens(computed(() => props.project.id));

const showCreateDialog = ref(false);
const newlyCreatedToken = ref<string | null>(null);
const copiedTokenId = ref<string | null>(null);

const errorMessage = computed(() => {
    if (queryError.value) {
        return queryError.value instanceof Error ? queryError.value.message : String(queryError.value);
    }
    return '';
});

function handleTokenCreated(token: string) {
    newlyCreatedToken.value = token;
    showCreateDialog.value = false;
    queryClient.invalidateQueries({ queryKey: queryKeys.projectTokens(props.project.id) });
}

async function handleRevoke(tokenId: string) {
    try {
        await ProjectService.deleteToken(props.project.id, tokenId);
        toast.success('Token revoked');
        queryClient.invalidateQueries({ queryKey: queryKeys.projectTokens(props.project.id) });
    } catch (e: any) {
        toast.error('Failed to revoke token: ' + e.message);
    }
}

async function copyToken(token: string, tokenId?: string) {
    await navigator.clipboard.writeText(token);
    if (tokenId) {
        copiedTokenId.value = tokenId;
        setTimeout(() => copiedTokenId.value = null, 2000);
    }
    toast.success('Token copied to clipboard');
}

function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

function formatDateTime(dateString: string): string {
    return new Date(dateString).toLocaleString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

function isExpired(expiresAt: string): boolean {
    return new Date(expiresAt) < new Date();
}

function dismissNewToken() {
    newlyCreatedToken.value = null;
}
</script>

<template>
    <div class="space-y-6">
        <SectionHeader
            title="Access Tokens"
            description="Tokens allow read-only access to this project's secrets from CLI, CI/CD pipelines, and Docker containers."
            action-label="Create Token"
            :action-icon="Plus"
            :action-disabled="!props.project.canEdit"
            @action="showCreateDialog = true"
        />

        <div>
                <!-- New token created banner -->
                <div
                    v-if="newlyCreatedToken"
                    class="mb-6 p-4 rounded-lg border bg-green-500/10 border-green-500/40 space-y-3"
                >
                    <div class="flex items-start gap-3">
                        <Key class="w-5 h-5 text-green-400 mt-0.5" />
                        <div class="flex-1 min-w-0">
                            <p class="font-medium text-green-200">Token created successfully</p>
                            <p class="text-sm text-green-300/70 mt-1">
                                Copy this token now. You won't be able to see it again.
                            </p>
                        </div>
                    </div>
                    <div class="flex items-center gap-2 bg-background/50 rounded p-2 font-mono text-sm">
                        <code class="flex-1 break-all">{{ newlyCreatedToken }}</code>
                        <Button variant="ghost" size="sm" @click="copyToken(newlyCreatedToken!)">
                            <Copy class="w-4 h-4" />
                        </Button>
                    </div>
                    <div class="flex items-start gap-2 text-sm text-orange-300/80">
                        <AlertTriangle class="w-4 h-4 mt-0.5 flex-shrink-0" />
                        <span>Store this token securely. It provides read-only access to all secrets in this project.</span>
                    </div>
                    <Button variant="outline" size="sm" @click="dismissNewToken">
                        I've copied the token
                    </Button>
                </div>

                <ErrorState
                    v-if="errorMessage"
                    title="Failed to load tokens"
                    :message="errorMessage"
                    :retry="refetch"
                />

                <PageLoader v-else-if="isLoading" message="Loading tokens..." />

                <EmptyState
                    v-else-if="!tokens || tokens.length === 0"
                    :icon="Key"
                    title="No access tokens yet"
                    description="Create a token to access this project from the CLI or CI/CD."
                />

                <div v-else class="space-y-3">
                    <div
                        v-for="token in tokens"
                        :key="token.id"
                        class="flex items-center justify-between p-4 rounded-lg border bg-card"
                        :class="{ 'opacity-60': isExpired(token.expiresAt) }"
                    >
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center gap-2">
                                <span class="font-medium">{{ token.name }}</span>
                                <code class="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                                    envie_{{ token.tokenPrefix }}...
                                </code>
                                <span
                                    v-if="isExpired(token.expiresAt)"
                                    class="text-xs bg-destructive/20 text-destructive px-1.5 py-0.5 rounded"
                                >
                                    Expired
                                </span>
                            </div>
                            <div class="text-sm text-muted-foreground mt-1 space-x-4">
                                <span>Created by {{ token.creatorName }} on {{ formatDate(token.createdAt) }}</span>
                                <span>Expires {{ formatDate(token.expiresAt) }}</span>
                                <span v-if="token.lastUsedAt">Last used {{ formatDateTime(token.lastUsedAt) }}</span>
                                <span v-else class="text-muted-foreground/60">Never used</span>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            size="sm"
                            class="text-destructive hover:text-destructive hover:bg-destructive/10"
                            @click="handleRevoke(token.id)"
                            :disabled="!props.project.canEdit"
                        >
                            <Trash2 class="w-4 h-4" />
                        </Button>
                    </div>
                </div>
        </div>

        <CreateTokenDialog
            v-model:open="showCreateDialog"
            :project="props.project"
            :decrypted-key="props.decryptedKey"
            @created="handleTokenCreated"
        />
    </div>
</template>
