<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useOrganizationStore } from '@/stores/organization';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Plus, Users, Folder, Loader2 } from 'lucide-vue-next';
import { useRouter } from 'vue-router';

const store = useOrganizationStore();
const router = useRouter();

const showCreateDialog = ref(false);
const newOrgName = ref('');
const isCreating = ref(false);
const isLoading = ref(true);

onMounted(async () => {
    try {
        await store.fetchOrganizations();
    } finally {
        isLoading.value = false;
    }
});

async function handleCreate() {
    if (!newOrgName.value) return;
    isCreating.value = true;
    try {
        await store.createOrganization(newOrgName.value);
        showCreateDialog.value = false;
        newOrgName.value = '';
    } catch (e) {
        console.error(e);
        // Show toaster error?
    } finally {
        isCreating.value = false;
    }
}

function openOrg(id: string) {
    router.push(`/organizations/${id}`);
}
</script>

<template>
    <div class="p-8 max-w-5xl mx-auto space-y-8">
        <div class="flex items-center justify-between">
            <div>
                <h2 class="text-3xl font-bold tracking-tight">Organizations</h2>
                <p class="text-muted-foreground">Manage your organizations and teams.</p>
            </div>
            <Dialog v-model:open="showCreateDialog">
                <DialogTrigger as-child>
                    <Button>
                        <Plus class="mr-2 h-4 w-4" />
                        Create Organization
                    </Button>
                </DialogTrigger>
                <DialogContent class="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>Create Organization</DialogTitle>
                        <DialogDescription>
                            Create a new organization to collaborate with your team.
                        </DialogDescription>
                    </DialogHeader>
                    <div class="grid gap-4 py-4">
                        <div class="space-y-2">
                            <Label for="name">Name</Label>
                            <Input id="name" v-model="newOrgName" placeholder="Acme Inc." />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button type="submit" @click="handleCreate" :disabled="isCreating || !newOrgName">
                            {{ isCreating ? 'Creating...' : 'Create' }}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="flex items-center justify-center py-20">
            <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div
                v-for="org in store.organizations"
                :key="org.id"
                class="rounded-xl border bg-card text-card-foreground shadow cursor-pointer hover:border-primary/50 transition-colors p-6 space-y-4"
                @click="openOrg(org.id)"
            >
                <div class="flex items-center justify-between">
                    <h3 class="font-semibold text-lg">{{ org.name }}</h3>
                    <span class="text-xs px-2 py-1 rounded-full bg-secondary text-secondary-foreground font-medium uppercase">
                        {{ org.role }}
                    </span>
                </div>

                <div class="grid grid-cols-2 gap-4 pt-2">
                    <div class="flex items-center gap-2 text-sm text-muted-foreground">
                        <Folder class="h-4 w-4" />
                        <span>{{ org.projectCount }} Projects</span>
                    </div>
                    <div class="flex items-center gap-2 text-sm text-muted-foreground">
                        <Users class="h-4 w-4" />
                        <span>{{ org.memberCount }} Members</span>
                    </div>
                </div>
            </div>

            <div v-if="store.organizations.length === 0" class="col-span-full text-center py-10 text-muted-foreground">
                No organizations found. Create one to get started.
            </div>
        </div>
    </div>
</template>
