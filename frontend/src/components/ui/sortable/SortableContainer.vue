<script setup lang="ts" generic="T">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import Sortable, { type Options, type SortableEvent } from 'sortablejs';

const props = withDefaults(defineProps<{
    /** The list of items */
    modelValue: T[];
    /** Function to get a unique key from each item */
    itemKey: keyof T | ((item: T) => string);
    /** CSS selector for the drag handle */
    handle?: string;
    /** Group name for drag between containers */
    group?: string | Options['group'];
    /** Animation duration in ms */
    animation?: number;
    /** Ghost element class */
    ghostClass?: string;
    /** Chosen element class */
    chosenClass?: string;
    /** Dragging element class */
    dragClass?: string;
    /** Filter selector - elements matching won't be draggable */
    filter?: string;
    /** Container tag */
    tag?: string;
    /** Container class */
    class?: string;
}>(), {
    animation: 200,
    ghostClass: 'sortable-ghost',
    chosenClass: 'sortable-chosen',
    dragClass: 'sortable-drag',
    tag: 'div',
});

const emit = defineEmits<{
    'update:modelValue': [items: T[]];
    'start': [event: SortableEvent];
    'end': [event: SortableEvent];
    'add': [item: T, newIndex: number, event: SortableEvent];
    'remove': [item: T, oldIndex: number, event: SortableEvent];
    'change': [items: T[]];
}>();

const containerRef = ref<HTMLElement | null>(null);
let sortable: Sortable | null = null;

const getKey = (item: T): string => {
    if (typeof props.itemKey === 'function') {
        return props.itemKey(item);
    }
    return String(item[props.itemKey]);
};

const createSortable = () => {
    if (!containerRef.value) return;

    if (sortable) {
        sortable.destroy();
    }

    const options: Options = {
        animation: props.animation,
        handle: props.handle,
        ghostClass: props.ghostClass,
        chosenClass: props.chosenClass,
        dragClass: props.dragClass,
        filter: props.filter,
        group: props.group,

        // Force fallback mode for macOS WebView compatibility
        forceFallback: true,
        fallbackOnBody: true,
        fallbackTolerance: 3,
        fallbackClass: 'sortable-fallback',

        onStart: (event: SortableEvent) => {
            emit('start', event);
        },

        onEnd: (event: SortableEvent) => {
            emit('end', event);
        },

        onUpdate: (event: SortableEvent) => {
            const { oldIndex, newIndex } = event;
            if (oldIndex === undefined || newIndex === undefined) return;

            const newList = [...props.modelValue];
            const [movedItem] = newList.splice(oldIndex, 1);
            newList.splice(newIndex, 0, movedItem);

            emit('update:modelValue', newList);
            emit('change', newList);
        },

        onAdd: (event: SortableEvent) => {
            const { newIndex, item } = event;
            if (newIndex === undefined) return;

            // Get the serialized item data from the element
            const itemData = item.dataset.sortableItem;
            if (!itemData) {
                console.warn('SortableContainer: Added item missing data-sortable-item');
                return;
            }

            try {
                const addedItem = JSON.parse(itemData) as T;
                emit('add', addedItem, newIndex, event);

                const newList = [...props.modelValue];
                newList.splice(newIndex, 0, addedItem);
                emit('update:modelValue', newList);
                emit('change', newList);
            } catch (e) {
                console.warn('SortableContainer: Failed to parse item data', e);
            }
        },

        onRemove: (event: SortableEvent) => {
            const { oldIndex } = event;
            if (oldIndex === undefined) return;

            const removedItem = props.modelValue[oldIndex];
            emit('remove', removedItem, oldIndex, event);

            const newList = [...props.modelValue];
            newList.splice(oldIndex, 1);
            emit('update:modelValue', newList);
            emit('change', newList);
        },
    };

    sortable = Sortable.create(containerRef.value, options);
};

onMounted(() => {
    createSortable();
});

onUnmounted(() => {
    if (sortable) {
        sortable.destroy();
        sortable = null;
    }
});

// Recreate sortable when group changes (for dynamic groups)
watch(() => props.group, () => {
    createSortable();
});

defineExpose({
    getKey,
});
</script>

<template>
    <component :is="tag" ref="containerRef" :class="props.class">
        <div
            v-for="item in modelValue"
            :key="getKey(item)"
            :data-sortable-key="getKey(item)"
            :data-sortable-item="JSON.stringify(item)"
        >
            <slot name="item" :element="item" />
        </div>
    </component>
</template>
