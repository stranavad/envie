<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Building2 } from 'lucide-vue-next';
import { OrganizationService } from '@/services/organization.service';

const props = defineProps<{
    organizationId: string;
    organizationName: string;
    canEdit: boolean;
}>();

const emit = defineEmits<{
    'name-updated': [name: string];
}>();

const editName = ref(props.organizationName);
const isSaving = ref(false);
const successMessage = ref('');
const errorMessage = ref('');

watch(() => props.organizationName, (newName) => {
    editName.value = newName;
});

async function handleUpdateName() {
    if (!editName.value || editName.value === props.organizationName) return;

    isSaving.value = true;
    successMessage.value = '';
    errorMessage.value = '';

    try {
        await OrganizationService.updateOrganization(props.organizationId, { name: editName.value });
        successMessage.value = 'Organization name updated successfully.';
        emit('name-updated', editName.value);
    } catch (e: any) {
        errorMessage.value = 'Failed to update organization name: ' + (e.message || e);
    } finally {
        isSaving.value = false;
    }
}
</script>

<template>
    <div class="space-y-6">
        <Card>
            <CardHeader>
                <CardTitle class="flex items-center gap-2">
                    <Building2 class="w-5 h-5" />
                    Basic Information
                </CardTitle>
                <CardDescription>
                    Manage your organization's basic settings.
                </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
                <div class="grid gap-2">
                    <Label for="orgName">Organization Name</Label>
                    <div class="flex gap-2">
                        <Input
                            id="orgName"
                            v-model="editName"
                            class="max-w-md"
                            :disabled="!canEdit"
                            @keyup.enter="handleUpdateName"
                        />
                        <Button
                            @click="handleUpdateName"
                            :disabled="!canEdit || isSaving || editName === organizationName"
                        >
                            {{ isSaving ? 'Saving...' : 'Save' }}
                        </Button>
                    </div>
                    <p v-if="!canEdit" class="text-sm text-muted-foreground">
                        Only organization owners and admins can update the organization name.
                    </p>
                </div>

                <div v-if="successMessage" class="text-sm text-green-600 font-medium">
                    {{ successMessage }}
                </div>
                <div v-if="errorMessage" class="text-sm text-destructive">
                    {{ errorMessage }}
                </div>
            </CardContent>
        </Card>
    </div>
</template>
