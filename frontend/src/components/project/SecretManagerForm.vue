<script setup lang="ts">
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Textarea} from '@/components/ui/textarea';
import {Label} from '@/components/ui/label';
import {Input} from '@/components/ui/input';
import {onMounted, ref} from 'vue';
import {SecretManagerService, serviceAccountKeySchema} from '@/services/secret-manager.service';
import {type SecretManagerConfig, SecretManagerConfigService} from '@/services/secret-manager-config.service';
import {AlertCircle, CheckCircle2, Plug, Save} from 'lucide-vue-next';
import { useConfigEncryption } from '@/composables/useConfigEncryption';

const props = defineProps<{
    project: any;
    decryptedKey: string | null;
    initialConfig?: SecretManagerConfig | null;
}>();

const emit = defineEmits<{
    (e: 'saved'): void;
    (e: 'cancel'): void;
}>();

const { decryptSecretManagerConfig, encryptConfigValue } = useConfigEncryption();

const formName = ref('');
const formKey = ref('');
const formError = ref('');
const isTestingConnection = ref(false);
const testConnectionSuccess = ref(false);
const availableSecrets = ref<string[]>([]);
const isSaving = ref(false);

onMounted(async () => {
  if(!props.initialConfig){
    return
  }

  formName.value = props.initialConfig.name;

  try {
    if (!props.decryptedKey) {
      formError.value = "Project key not available";
      return;
    }
    formKey.value = await decryptSecretManagerConfig(props.decryptedKey, props.initialConfig.encryptedKey)
  } catch (e) {
      console.error(e);
      formError.value = "Failed to decrypt key: " + e;
  }
});

async function handleTestConnection() {
    isTestingConnection.value = true;
    formError.value = '';
    availableSecrets.value = [];
    testConnectionSuccess.value = false;

    try {
         // Validate JSON format
        let parsedJson;
        try {
            parsedJson = JSON.parse(formKey.value);
        } catch (e) {
             throw new Error("Invalid JSON format");
        }

        // Validate Schema
        const result = serviceAccountKeySchema.safeParse(parsedJson);
        if (!result.success) {
             const issues = result.error.issues.map(i => `${i.path.join('.')}: ${i.message}`).join(', ');
             throw new Error(`Invalid Service Account Key: ${issues}`);
        }

        // Test with fast check first
        const success = await SecretManagerService.testConnection(formKey.value);
        if (!success) {
            throw new Error("Connection test failed. Check permissions or key validity.");
        }
        
        availableSecrets.value = await SecretManagerService.listSecrets(formKey.value);
        
        testConnectionSuccess.value = true;
    } catch (err: any) {
        formError.value = err.message || err.toString();
        testConnectionSuccess.value = false;
    } finally {
        isTestingConnection.value = false;
    }
}

async function handleSave() {
    if (!testConnectionSuccess.value) {
        formError.value = "You must successfully test the connection before saving.";
        return;
    }
    if (!formName.value || !formKey.value) {
        formError.value = "Name and Key are required.";
        return;
    }
    if (!props.decryptedKey) {
        formError.value = "Project key missing.";
        return;
    }

    isSaving.value = true;
    try {
        // Encrypt Key
        const encryptedKey = await encryptConfigValue(props.decryptedKey!, formKey.value);


        if (props.initialConfig) {
             await SecretManagerConfigService.updateConfig(props.project.id, props.initialConfig.id, formName.value, encryptedKey);
        } else {
             await SecretManagerConfigService.createConfig(props.project.id, formName.value, encryptedKey);
        }
        
        emit('saved');
    } catch (e: any) {
        formError.value = "Failed to save: " + e.toString();
    } finally {
        isSaving.value = false;
    }
}
</script>

<template>
    <Card class="shadow-none">
        <CardHeader>
            <CardTitle class="text-base">{{ initialConfig ? 'Edit Configuration' : 'Add New Configuration' }}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-4">
                <div class="space-y-2">
                <Label>Configuration Name</Label>
                <Input v-model="formName" placeholder="e.g., Production Secrets" maxlength="50" />
            </div>

            <div class="space-y-2">
                <Label>Service Account Key (JSON)</Label>
                <Textarea 
                    v-model="formKey" 
                    placeholder='{ "type": "service_account", ... }' 
                    class="font-mono text-xs h-32"
                />
                    <p class="text-xs text-muted-foreground">
                        Paste your Google Cloud Service Account JSON key here. It will be encrypted before storage.
                </p>
            </div>
            
            <div v-if="formError" class="flex items-center p-4 text-sm text-destructive bg-destructive/15 rounded-md">
                <AlertCircle class="w-4 h-4 mr-2" />
                {{ formError }}
            </div>

                <div v-if="testConnectionSuccess" class="flex items-center p-4 text-sm text-green-700 bg-green-500/15 rounded-md">
                <CheckCircle2 class="w-4 h-4 mr-2" />
                Connection Successful! Found {{ availableSecrets.length }} secrets.
            </div>

            <div class="flex justify-between pt-2">
                <Button type="button" variant="secondary" @click="handleTestConnection" :disabled="isTestingConnection || !formKey">
                    <Plug class="w-4 h-4 mr-2" />
                    {{ isTestingConnection ? 'Testing...' : 'Test Connection' }}
                </Button>
                <div class="flex space-x-2">
                    <Button type="button" variant="ghost" @click="emit('cancel')">Cancel</Button>
                    <Button type="button" @click="handleSave" :disabled="!testConnectionSuccess || isTestingConnection || isSaving">
                        <Save class="w-4 h-4 mr-2" /> {{ isSaving ? 'Saving...' : 'Save' }}
                    </Button>
                </div>
            </div>
        </CardContent>
    </Card>
</template>
