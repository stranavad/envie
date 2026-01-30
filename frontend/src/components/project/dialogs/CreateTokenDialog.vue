<script setup lang="ts">
import { ref, computed } from 'vue';
import { type ProjectDetail, ProjectService } from '@/services/project.service';
import { EncryptionService } from '@/services/encryption.service';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Calendar } from '@/components/ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Loader2, AlertTriangle, CalendarIcon } from 'lucide-vue-next';
import { type DateValue, getLocalTimeZone, today } from '@internationalized/date';

const props = defineProps<{
    project: ProjectDetail;
    decryptedKey: string;
}>();

const open = defineModel<boolean>('open', { required: true });
const emit = defineEmits<{
    created: [token: string];
}>();

const name = ref('');
const expiresAt = ref<DateValue>();
const isCreating = ref(false);
const error = ref('');

const minDate = today(getLocalTimeZone()).add({ days: 1 });

const isValid = computed(() => {
    return name.value.trim().length > 0 && expiresAt.value !== undefined;
});

const formattedDate = computed(() => {
    if (!expiresAt.value) return '';
    return expiresAt.value.toDate(getLocalTimeZone()).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    });
});

async function handleCreate() {
    if (!isValid.value || !expiresAt.value) return;

    isCreating.value = true;
    error.value = '';

    try {
        const generated = await EncryptionService.generateAccessToken(props.decryptedKey);

        const expiresAtDate = expiresAt.value.toDate(getLocalTimeZone());
        expiresAtDate.setHours(23, 59, 59, 999);

        await ProjectService.createToken(props.project.id, {
            name: name.value.trim(),
            expiresAt: expiresAtDate.toISOString(),
            tokenPrefix: generated.tokenPrefix,
            identityIdHash: generated.identityIdHash,
            encryptedProjectKey: generated.encryptedProjectKey,
        });

        emit('created', generated.token);
        resetForm();
    } catch (e: any) {
        error.value = e.message || 'Failed to create token';
    } finally {
        isCreating.value = false;
    }
}

function resetForm() {
    name.value = '';
    expiresAt.value = undefined;
    error.value = '';
}

function handleOpenChange(value: boolean) {
    if (!value) {
        resetForm();
    }
    open.value = value;
}
</script>

<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent class="sm:max-w-md">
            <DialogHeader>
                <DialogTitle>Create Access Token</DialogTitle>
                <DialogDescription>
                    Create a token to access this project's secrets from CLI or CI/CD.
                </DialogDescription>
            </DialogHeader>

            <div class="space-y-4 py-4">
                <div class="flex items-start gap-2 p-3 rounded-lg bg-orange-500/10 border border-orange-500/30 text-sm">
                    <AlertTriangle class="w-4 h-4 text-orange-400 mt-0.5 flex-shrink-0" />
                    <span class="text-orange-200">
                        This token will have read-only access to all secrets in this project. Store it securely.
                    </span>
                </div>

                <div class="space-y-2">
                    <Label for="name">Token Name</Label>
                    <Input
                        id="name"
                        v-model="name"
                        placeholder="e.g., GitHub Actions, Production Docker"
                        :disabled="isCreating"
                    />
                    <p class="text-xs text-muted-foreground">
                        A descriptive name to identify this token's purpose.
                    </p>
                </div>

                <div class="space-y-2">
                    <Label>Expiration Date</Label>
                    <Popover>
                        <PopoverTrigger as-child>
                            <Button
                                variant="outline"
                                class="w-full justify-start text-left font-normal"
                                :class="{ 'text-muted-foreground': !expiresAt }"
                                :disabled="isCreating"
                            >
                                <CalendarIcon class="mr-2 h-4 w-4" />
                                {{ formattedDate || 'Select expiration date' }}
                            </Button>
                        </PopoverTrigger>
                        <PopoverContent class="w-auto p-0" align="start">
                            <Calendar
                                v-model="expiresAt"
                                :min-value="minDate"
                                initial-focus
                            />
                        </PopoverContent>
                    </Popover>
                    <p class="text-xs text-muted-foreground">
                        Token will stop working after this date.
                    </p>
                </div>

                <div v-if="error" class="text-sm text-destructive">
                    {{ error }}
                </div>
            </div>

            <DialogFooter>
                <Button variant="outline" @click="open = false" :disabled="isCreating">
                    Cancel
                </Button>
                <Button @click="handleCreate" :disabled="!isValid || isCreating">
                    <Loader2 v-if="isCreating" class="w-4 h-4 mr-2 animate-spin" />
                    Create Token
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>
