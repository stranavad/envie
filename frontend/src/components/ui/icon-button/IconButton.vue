<script setup lang="ts">
import {
    TooltipRoot,
    TooltipTrigger,
    TooltipContent,
    TooltipPortal,
    TooltipProvider,
} from 'reka-ui';

defineProps<{
    tooltip: string;
    disabled?: boolean;
}>();

defineEmits<{
    click: [event: MouseEvent];
}>();
</script>

<template>
    <TooltipProvider :delay-duration="150">
        <TooltipRoot>
            <TooltipTrigger as-child>
                <button
                    :disabled="disabled"
                    class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50 transition-colors"
                    @click="$emit('click', $event)"
                >
                    <slot />
                </button>
            </TooltipTrigger>
            <TooltipPortal>
                <TooltipContent
                    side="bottom"
                    :side-offset="6"
                    class="z-50 rounded-md bg-popover border border-border px-3 py-1.5 text-xs text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
                >
                    {{ tooltip }}
                </TooltipContent>
            </TooltipPortal>
        </TooltipRoot>
    </TooltipProvider>
</template>
