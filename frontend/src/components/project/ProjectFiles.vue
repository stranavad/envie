<script setup lang="ts">
import { ref, onMounted } from 'vue';
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
import { Upload, Download, Trash2, File, Loader2, AlertCircle, Check } from 'lucide-vue-next';
import { ProjectService, type ProjectDetail, type ProjectFile } from '@/services/project.service';
import { EncryptionService } from '@/services/encryption.service';
import { save } from '@tauri-apps/plugin-dialog';
import { writeFile } from '@tauri-apps/plugin-fs';

const props = defineProps<{
    project: ProjectDetail;
    projectKey: string;
}>();

const files = ref<ProjectFile[]>([]);
const isLoading = ref(false);
const error = ref('');

// Upload state
const isUploadOpen = ref(false);
const uploadFile = ref<File | null>(null);
const isUploading = ref(false);
const uploadProgress = ref('');
const uploadError = ref('');

// Download state
const downloadingFileId = ref<string | null>(null);
const downloadSuccess = ref('');

const MAX_FILE_SIZE = 1 * 1024 * 1024; // 1MB

onMounted(() => {
    loadFiles();
});

async function loadFiles() {
    isLoading.value = true;
    error.value = '';
    try {
        files.value = await ProjectService.getFiles(props.project.id);
    } catch (e: any) {
        console.error('Failed to load files', e);
        error.value = e.message || 'Failed to load files';
    } finally {
        isLoading.value = false;
    }
}

function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, {
        dateStyle: 'medium'
    });
}

function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files[0]) {
        const file = input.files[0];
        if (file.size > MAX_FILE_SIZE) {
            uploadError.value = `File too large. Maximum size is ${formatFileSize(MAX_FILE_SIZE)}`;
            uploadFile.value = null;
            return;
        }
        uploadFile.value = file;
        uploadError.value = '';
    }
}

async function handleUpload() {
    if (!uploadFile.value || !props.projectKey) return;

    isUploading.value = true;
    uploadError.value = '';
    uploadProgress.value = 'Reading file...';

    try {
        const file = uploadFile.value;

        // Read file as ArrayBuffer
        const arrayBuffer = await file.arrayBuffer();
        const fileData = new Uint8Array(arrayBuffer);

        // Calculate checksum of original file
        uploadProgress.value = 'Calculating checksum...';
        const checksum = await calculateChecksum(fileData);

        // Generate File Encryption Key (FEK)
        uploadProgress.value = 'Generating encryption key...';
        const fek = EncryptionService.generateAesKey();

        // Encrypt file content with FEK
        uploadProgress.value = 'Encrypting file...';
        const encryptedData = await encryptFile(fileData, fek);

        // Encrypt FEK with project key
        uploadProgress.value = 'Securing encryption key...';
        const encryptedFek = await EncryptionService.encryptValue(props.projectKey, fek);

        // Create form data
        uploadProgress.value = 'Uploading...';
        const formData = new FormData();
        formData.append('file', new Blob([encryptedData]), file.name);
        formData.append('name', file.name);
        formData.append('encryptedFek', encryptedFek);
        formData.append('checksum', checksum);
        formData.append('mimeType', file.type || 'application/octet-stream');
        formData.append('originalSize', file.size.toString());

        await ProjectService.uploadFile(props.project.id, {
            file: new Blob([encryptedData]) as File,
            name: file.name,
            encryptedFek,
            checksum,
            mimeType: file.type || 'application/octet-stream',
            originalSize: file.size
        });

        uploadProgress.value = 'Done!';
        isUploadOpen.value = false;
        uploadFile.value = null;
        await loadFiles();

    } catch (e: any) {
        console.error('Upload failed', e);
        uploadError.value = e.response?.data?.error || e.message || 'Upload failed';
    } finally {
        isUploading.value = false;
        uploadProgress.value = '';
    }
}

async function handleDownload(file: ProjectFile) {
    if (!props.projectKey) return;

    downloadingFileId.value = file.id;
    downloadSuccess.value = '';

    try {
        // Get encrypted file from server
        const downloadRes = await ProjectService.downloadFile(props.project.id, file.id);
        const { data: base64Data, encryptedFek, checksum, name } = downloadRes;

        // Decrypt FEK with project key
        const fek = await EncryptionService.decryptValue(props.projectKey, encryptedFek);

        // Decode base64 to encrypted bytes
        const encryptedData = Uint8Array.from(atob(base64Data), c => c.charCodeAt(0));

        // Decrypt file content
        const decryptedData = await decryptFile(encryptedData, fek);

        // Verify checksum
        const actualChecksum = await calculateChecksum(decryptedData);
        if (checksum && actualChecksum !== checksum) {
            throw new Error('File integrity check failed');
        }

        // Get file extension from name
        const ext = name.includes('.') ? name.split('.').pop() : undefined;

        // Show save dialog
        const savePath = await save({
            defaultPath: name,
            filters: ext ? [{ name: 'File', extensions: [ext] }] : undefined
        });

        if (!savePath) {
            // User cancelled
            return;
        }

        // Write file to selected location
        await writeFile(savePath, decryptedData);

        downloadSuccess.value = `File saved to ${savePath}`;

        // Clear success message after 5 seconds
        setTimeout(() => {
            downloadSuccess.value = '';
        }, 5000);

    } catch (e: any) {
        console.error('Download failed', e);
        error.value = e.message || 'Download failed';
    } finally {
        downloadingFileId.value = null;
    }
}

async function handleDelete(file: ProjectFile) {
    try {
        await ProjectService.deleteFile(props.project.id, file.id);
        await loadFiles();
    } catch (e: any) {
        console.error('Delete failed', e);
        alert(e.message || 'Delete failed');
    }
}

// Encrypt file using AES-GCM
async function encryptFile(data: Uint8Array, keyBase64: string): Promise<Uint8Array> {
    const key = await crypto.subtle.importKey(
        'raw',
        Uint8Array.from(atob(keyBase64), c => c.charCodeAt(0)),
        { name: 'AES-GCM' },
        false,
        ['encrypt']
    );

    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await crypto.subtle.encrypt(
        { name: 'AES-GCM', iv },
        key,
        data
    );

    // Prepend IV to encrypted data
    const result = new Uint8Array(iv.length + encrypted.byteLength);
    result.set(iv);
    result.set(new Uint8Array(encrypted), iv.length);
    return result;
}

// Decrypt file using AES-GCM
async function decryptFile(data: Uint8Array, keyBase64: string): Promise<Uint8Array> {
    const key = await crypto.subtle.importKey(
        'raw',
        Uint8Array.from(atob(keyBase64), c => c.charCodeAt(0)),
        { name: 'AES-GCM' },
        false,
        ['decrypt']
    );

    // Extract IV from start of data
    const iv = data.slice(0, 12);
    const encryptedContent = data.slice(12);

    const decrypted = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv },
        key,
        encryptedContent
    );

    return new Uint8Array(decrypted);
}

// Calculate SHA-256 checksum
async function calculateChecksum(data: Uint8Array): Promise<string> {
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}
</script>

<template>
    <div class="space-y-6">
        <div v-if="error" class="bg-destructive/15 text-destructive p-4 rounded-md flex items-center gap-2">
            <AlertCircle class="w-4 h-4" />
            {{ error }}
        </div>

        <div v-if="downloadSuccess" class="bg-green-500/15 text-green-700 p-4 rounded-md flex items-center gap-2">
            <Check class="w-4 h-4" />
            {{ downloadSuccess }}
        </div>

        <!-- Header -->
        <div class="flex justify-between items-center">
            <div>
                <h3 class="text-lg font-medium">Files</h3>
                <p class="text-sm text-muted-foreground">
                    Encrypted files shared with project members. Max 1MB per file.
                </p>
            </div>
            <Button v-if="project.canEdit" @click="isUploadOpen = true">
                <Upload class="w-4 h-4 mr-2" />
                Upload File
            </Button>
        </div>

        <!-- Loading state -->
        <div v-if="isLoading" class="flex flex-col items-center py-12 text-muted-foreground">
            <Loader2 class="h-8 w-8 animate-spin mb-4" />
            <p>Loading files...</p>
        </div>

        <!-- Files list -->
        <div v-else-if="files.length > 0" class="bg-card rounded-lg border shadow-sm">
            <div class="divide-y divide-border">
                <div
                    v-for="file in files"
                    :key="file.id"
                    class="flex items-center justify-between p-4 hover:bg-muted/50 transition-colors"
                >
                    <div class="flex items-center gap-4">
                        <div class="p-2 bg-primary/10 rounded-lg text-primary">
                            <File class="w-5 h-5" />
                        </div>
                        <div>
                            <div class="font-medium">{{ file.name }}</div>
                            <div class="flex items-center gap-2 text-sm text-muted-foreground">
                                <span>{{ formatFileSize(file.sizeBytes) }}</span>
                                <span>·</span>
                                <span>{{ file.uploadedBy.name }}</span>
                                <span>·</span>
                                <span>{{ formatDate(file.createdAt) }}</span>
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <Button
                            variant="ghost"
                            size="sm"
                            @click="handleDownload(file)"
                            :disabled="downloadingFileId === file.id"
                        >
                            <Loader2 v-if="downloadingFileId === file.id" class="w-4 h-4 animate-spin" />
                            <Download v-else class="w-4 h-4" />
                        </Button>
                        <Button
                            v-if="project.canEdit"
                            variant="ghost"
                            size="sm"
                            class="text-muted-foreground hover:text-destructive"
                            @click="handleDelete(file)"
                        >
                            <Trash2 class="w-4 h-4" />
                        </Button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Empty state -->
        <div v-else class="text-center py-8 text-muted-foreground border rounded-lg bg-muted/20">
            <File class="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No files uploaded yet.</p>
            <p class="text-sm mt-1">Upload files to share with project members.</p>
        </div>

        <!-- Upload Dialog -->
        <Dialog v-model:open="isUploadOpen">
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Upload File</DialogTitle>
                    <DialogDescription>
                        File will be encrypted before upload. Maximum size: 1MB.
                    </DialogDescription>
                </DialogHeader>
                <div class="py-4 space-y-4">
                    <div class="space-y-2">
                        <Label for="file">Select File</Label>
                        <Input
                            id="file"
                            type="file"
                            @change="handleFileSelect"
                            :disabled="isUploading"
                        />
                    </div>

                    <div v-if="uploadFile" class="text-sm text-muted-foreground">
                        Selected: {{ uploadFile.name }} ({{ formatFileSize(uploadFile.size) }})
                    </div>

                    <div v-if="uploadProgress" class="text-sm text-muted-foreground flex items-center gap-2">
                        <Loader2 class="w-4 h-4 animate-spin" />
                        {{ uploadProgress }}
                    </div>

                    <div v-if="uploadError" class="text-sm text-destructive">
                        {{ uploadError }}
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" @click="isUploadOpen = false" :disabled="isUploading">
                        Cancel
                    </Button>
                    <Button @click="handleUpload" :disabled="isUploading || !uploadFile">
                        {{ isUploading ? 'Uploading...' : 'Upload' }}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    </div>
</template>
