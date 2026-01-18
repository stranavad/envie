<script setup lang="ts">
import { ref, computed, watch } from 'vue';
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
} from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface Team {
  id: string;
  name: string;
}

const props = defineProps<{
  open: boolean;
  teams: Team[]; // Required: teams must be provided
  defaultTeamId?: string;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'create', payload: { name: string; teamId: string }): void;
}>();

const isOpen = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val),
});

const projectName = ref('');
const selectedTeamId = ref(props.defaultTeamId || '');

// Reset form when dialog opens
watch(() => props.open, (newVal) => {
  if (newVal) {
    projectName.value = '';
    selectedTeamId.value = props.defaultTeamId || '';
  }
});

// Update selectedTeamId when defaultTeamId changes
watch(() => props.defaultTeamId, (newVal) => {
  if (newVal) {
    selectedTeamId.value = newVal;
  }
});

const canCreate = computed(() => {
  return projectName.value.trim() && selectedTeamId.value && !props.loading;
});

function handleCreate() {
  if (!canCreate.value) return;
  emit('create', { name: projectName.value.trim(), teamId: selectedTeamId.value });
}
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>Create Project</DialogTitle>
        <DialogDescription>
          Create a new project within a team. Select a team to assign this project to.
        </DialogDescription>
      </DialogHeader>

      <div class="grid gap-4 py-4">
        <div class="space-y-2">
            <Label for="projectName">Project Name</Label>
            <Input
                id="projectName"
                v-model="projectName"
                placeholder="My Top Secret Project"
                @keyup.enter="handleCreate"
                autofocus
            />
        </div>

        <div class="space-y-2">
            <Label for="teamSelect">Team <span class="text-destructive">*</span></Label>
            <Select v-model="selectedTeamId">
            <SelectTrigger>
                <SelectValue placeholder="Select a team" />
            </SelectTrigger>
            <SelectContent>
                <SelectItem v-for="team in teams" :key="team.id" :value="team.id">
                {{ team.name }}
                </SelectItem>
            </SelectContent>
            </Select>
            <p v-if="teams.length === 0" class="text-sm text-muted-foreground">
              No teams available. Create a team first.
            </p>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="isOpen = false">Cancel</Button>
        <Button @click="handleCreate" :disabled="!canCreate">
          {{ loading ? 'Creating...' : 'Create Project' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
