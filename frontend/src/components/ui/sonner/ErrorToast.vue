<script setup lang="ts">
import { ref } from 'vue'
import { Copy, Check, AlertCircle } from 'lucide-vue-next'

const props = defineProps<{
    message: string
}>()

const copied = ref(false)

async function copyToClipboard() {
    try {
        await navigator.clipboard.writeText(props.message)
        copied.value = true
        setTimeout(() => {
            copied.value = false
        }, 2000)
    } catch (err) {
        console.error('Failed to copy:', err)
    }
}
</script>

<template>
    <div class="flex items-start gap-3 w-full max-w-[356px] rounded-lg border border-destructive bg-destructive p-4 text-destructive-foreground shadow-lg">
        <AlertCircle class="h-5 w-5 shrink-0 mt-0.5" />
        <div class="flex-1 min-w-0">
            <p class="text-sm font-medium">Error</p>
            <p class="text-sm opacity-90 mt-1 break-words">{{ message }}</p>
        </div>
        <button
            @click="copyToClipboard"
            class="shrink-0 p-1.5 rounded hover:bg-destructive-foreground/10 transition-colors"
            title="Copy error message"
        >
            <Check v-if="copied" class="h-4 w-4" />
            <Copy v-else class="h-4 w-4" />
        </button>
    </div>
</template>
