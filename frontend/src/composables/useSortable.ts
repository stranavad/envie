import { ref, onMounted, onUnmounted, watch, type Ref } from 'vue';
import Sortable, { type Options, type SortableEvent } from 'sortablejs';

export interface UseSortableOptions<T> {
    /** The list of items to sort */
    list: Ref<T[]>;
    /** Function to get a unique key from each item */
    itemKey: (item: T) => string;
    /** CSS selector for the drag handle (optional - if not provided, entire item is draggable) */
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
    /** Fallback element class (used in fallback mode) */
    fallbackClass?: string;
    /** Filter selector - elements matching won't be draggable */
    filter?: string;
    /** Called when sort order changes within this container */
    onUpdate?: (items: T[], event: SortableEvent) => void;
    /** Called when an item is added from another container */
    onAdd?: (item: T, newIndex: number, event: SortableEvent) => void;
    /** Called when an item is removed to another container */
    onRemove?: (item: T, oldIndex: number, event: SortableEvent) => void;
    /** Called when drag starts */
    onStart?: (event: SortableEvent) => void;
    /** Called when drag ends */
    onEnd?: (event: SortableEvent) => void;
    /** Called on any change (add, remove, update) */
    onChange?: (items: T[], event: SortableEvent) => void;
    /** Additional Sortable.js options */
    sortableOptions?: Partial<Options>;
}

export interface UseSortableReturn {
    /** The Sortable instance (available after mount) */
    sortable: Ref<Sortable | null>;
    /** Destroy the sortable instance */
    destroy: () => void;
    /** Reinitialize the sortable instance */
    reinit: () => void;
}

export function useSortable<T>(
    containerRef: Ref<HTMLElement | null>,
    options: UseSortableOptions<T>
): UseSortableReturn {
    const sortable = ref<Sortable | null>(null);

    const createSortable = () => {
        if (!containerRef.value) return;

        // Destroy existing instance
        if (sortable.value) {
            sortable.value.destroy();
        }

        const sortableOptions: Options = {
            animation: options.animation ?? 200,
            handle: options.handle,
            ghostClass: options.ghostClass ?? 'sortable-ghost',
            chosenClass: options.chosenClass ?? 'sortable-chosen',
            dragClass: options.dragClass ?? 'sortable-drag',
            fallbackClass: options.fallbackClass ?? 'sortable-fallback',
            filter: options.filter,

            // Force fallback mode for better cross-platform compatibility (especially macOS WebView)
            forceFallback: true,
            fallbackOnBody: true,
            fallbackTolerance: 3,

            // Group configuration
            group: options.group,

            onStart: (event: SortableEvent) => {
                options.onStart?.(event);
            },

            onEnd: (event: SortableEvent) => {
                options.onEnd?.(event);
            },

            onUpdate: (event: SortableEvent) => {
                const { oldIndex, newIndex } = event;
                if (oldIndex === undefined || newIndex === undefined) return;

                // Create new array with updated order
                const newList = [...options.list.value];
                const [movedItem] = newList.splice(oldIndex, 1);
                newList.splice(newIndex, 0, movedItem);

                options.onUpdate?.(newList, event);
                options.onChange?.(newList, event);
            },

            onAdd: (event: SortableEvent) => {
                const { newIndex, item } = event;
                if (newIndex === undefined) return;

                // Get the item key from the DOM element
                const itemKey = item.dataset.sortableKey;
                if (!itemKey) {
                    console.warn('useSortable: Added item missing data-sortable-key attribute');
                    return;
                }

                // Find the item in the source list by parsing the event
                // The item data should be passed via a custom attribute or we need to find it
                const addedItem = (event as SortableEvent & { item: HTMLElement }).item;
                const sourceData = addedItem.dataset.sortableItem;

                if (sourceData) {
                    try {
                        const parsedItem = JSON.parse(sourceData) as T;
                        options.onAdd?.(parsedItem, newIndex, event);

                        // Update local list
                        const newList = [...options.list.value];
                        newList.splice(newIndex, 0, parsedItem);
                        options.onChange?.(newList, event);
                    } catch (e) {
                        console.warn('useSortable: Failed to parse item data', e);
                    }
                }
            },

            onRemove: (event: SortableEvent) => {
                const { oldIndex } = event;
                if (oldIndex === undefined) return;

                const removedItem = options.list.value[oldIndex];
                options.onRemove?.(removedItem, oldIndex, event);

                // Update local list
                const newList = [...options.list.value];
                newList.splice(oldIndex, 1);
                options.onChange?.(newList, event);
            },

            ...options.sortableOptions,
        };

        sortable.value = Sortable.create(containerRef.value, sortableOptions);
    };

    const destroy = () => {
        if (sortable.value) {
            sortable.value.destroy();
            sortable.value = null;
        }
    };

    const reinit = () => {
        createSortable();
    };

    onMounted(() => {
        createSortable();
    });

    onUnmounted(() => {
        destroy();
    });

    // Watch for container ref changes
    watch(containerRef, (newContainer) => {
        if (newContainer) {
            createSortable();
        } else {
            destroy();
        }
    });

    return {
        sortable: sortable as Ref<Sortable | null>,
        destroy,
        reinit,
    };
}

/**
 * Helper to generate data attributes for sortable items
 */
export function sortableItemAttrs<T>(item: T, key: string): Record<string, string> {
    return {
        'data-sortable-key': key,
        'data-sortable-item': JSON.stringify(item),
    };
}
