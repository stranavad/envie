import { ref, computed, watch, type Ref } from 'vue';
import type { ConfigItem } from '@/services/project.service';

export function useCategoryManagement(configItems: Ref<ConfigItem[]>) {
    // Category state
    const collapsedCategories = ref<Set<string>>(new Set());
    const emptyCategories = ref<string[]>([]);
    const categoryOrder = ref<string[]>([]);
    const categoryItemOrders = ref<Map<string, string[]>>(new Map());
    const uncategorizedItemOrder = ref<string[]>([]);
    const initialCategoriesLoaded = ref(false);

    // Get all unique category names
    const allCategoryNames = computed(() => {
        const categorySet = new Set<string>();

        configItems.value.forEach(item => {
            if (item.category) {
                categorySet.add(item.category);
            }
        });

        // Add empty categories
        emptyCategories.value.forEach(cat => {
            categorySet.add(cat);
        });

        return categorySet;
    });

    // Get categories in order (respecting user-defined order)
    const categories = computed(() => {
        const allNames = allCategoryNames.value;

        // If we have an explicit order, use it (filtering out deleted categories, adding new ones at end)
        if (categoryOrder.value.length > 0) {
            const ordered: string[] = [];

            // First add categories that are in the order list
            categoryOrder.value.forEach(cat => {
                if (allNames.has(cat)) {
                    ordered.push(cat);
                }
            });

            // Then add any new categories not in the order list
            allNames.forEach(cat => {
                if (!ordered.includes(cat)) {
                    ordered.push(cat);
                }
            });

            return ordered;
        }

        // Default: sort by minimum position of items
        const categoryMap = new Map<string, number>();

        configItems.value.forEach(item => {
            if (item.category) {
                const existing = categoryMap.get(item.category);
                if (existing === undefined || item.position < existing) {
                    categoryMap.set(item.category, item.position);
                }
            }
        });

        // Add empty categories at the end
        emptyCategories.value.forEach(cat => {
            if (!categoryMap.has(cat)) {
                categoryMap.set(cat, Infinity);
            }
        });

        return Array.from(categoryMap.entries())
            .sort((a, b) => a[1] - b[1])
            .map(([name]) => name);
    });

    // Check if we have any categories
    const hasCategories = computed(() => categories.value.length > 0);

    // Wrapper for draggable (needs objects with keys)
    const categoriesForDrag = computed(() =>
        categories.value.map(name => ({ name }))
    );

    // Get uncategorized items
    const uncategorizedItems = computed(() => {
        const items = configItems.value.filter(item => !item.category);

        // Check if we have a custom order
        if (uncategorizedItemOrder.value.length > 0) {
            return items.sort((a, b) => {
                const aIdx = uncategorizedItemOrder.value.indexOf(a.id);
                const bIdx = uncategorizedItemOrder.value.indexOf(b.id);
                if (aIdx !== -1 && bIdx !== -1) return aIdx - bIdx;
                if (aIdx !== -1) return -1;
                if (bIdx !== -1) return 1;
                return a.position - b.position;
            });
        }

        return items.sort((a, b) => a.position - b.position);
    });

    // Get items for a specific category
    function getCategoryItems(category: string): ConfigItem[] {
        const items = configItems.value.filter(item => item.category === category);

        // Check if we have a custom order for this category
        const customOrder = categoryItemOrders.value.get(category);
        if (customOrder && customOrder.length > 0) {
            // Sort by custom order, fallback to position for items not in custom order
            return items.sort((a, b) => {
                const aIdx = customOrder.indexOf(a.id);
                const bIdx = customOrder.indexOf(b.id);
                if (aIdx !== -1 && bIdx !== -1) return aIdx - bIdx;
                if (aIdx !== -1) return -1;
                if (bIdx !== -1) return 1;
                return a.position - b.position;
            });
        }

        return items.sort((a, b) => a.position - b.position);
    }

    // Toggle category collapse
    function toggleCategory(category: string) {
        if (collapsedCategories.value.has(category)) {
            collapsedCategories.value.delete(category);
        } else {
            collapsedCategories.value.add(category);
        }
        // Trigger reactivity
        collapsedCategories.value = new Set(collapsedCategories.value);
    }

    // Check if category is collapsed
    function isCategoryCollapsed(category: string): boolean {
        return collapsedCategories.value.has(category);
    }

    // Add a new category
    function addCategory(name: string) {
        if (!categories.value.includes(name)) {
            emptyCategories.value.push(name);
        }
    }

    // Rename a category
    function renameCategory(oldName: string, newName: string) {
        // Update all items with this category
        configItems.value.forEach(item => {
            if (item.category === oldName) {
                item.category = newName;
            }
        });

        // Update emptyCategories if it was an empty category
        const emptyIdx = emptyCategories.value.indexOf(oldName);
        if (emptyIdx !== -1) {
            emptyCategories.value[emptyIdx] = newName;
        }

        // Update category order
        const orderIdx = categoryOrder.value.indexOf(oldName);
        if (orderIdx !== -1) {
            categoryOrder.value[orderIdx] = newName;
        }

        // Update categoryItemOrders
        if (categoryItemOrders.value.has(oldName)) {
            const order = categoryItemOrders.value.get(oldName)!;
            categoryItemOrders.value.delete(oldName);
            categoryItemOrders.value.set(newName, order);
            categoryItemOrders.value = new Map(categoryItemOrders.value);
        }
    }

    // Delete a category (moves items to uncategorized)
    function deleteCategory(category: string) {
        // Move all items in this category to uncategorized
        configItems.value.forEach(item => {
            if (item.category === category) {
                item.category = undefined;
            }
        });

        // Remove from emptyCategories if present
        emptyCategories.value = emptyCategories.value.filter(c => c !== category);

        // Remove from collapsed
        collapsedCategories.value.delete(category);
        collapsedCategories.value = new Set(collapsedCategories.value);

        // Remove from order tracking
        categoryItemOrders.value.delete(category);
        categoryItemOrders.value = new Map(categoryItemOrders.value);
    }

    // Recalculate all positions based on current visual order
    function recalculatePositions() {
        let position = 0;

        // First each category's items (categories at top)
        categories.value.forEach(category => {
            const items = getCategoryItems(category);
            items.forEach(item => {
                const idx = configItems.value.findIndex(i => i.id === item.id);
                if (idx !== -1) {
                    configItems.value[idx].position = position++;
                }
            });

            // Update the categoryItemOrders to match current positions
            categoryItemOrders.value.set(category, items.map(i => i.id));
        });

        // Then uncategorized items (at bottom)
        const uncatItems = uncategorizedItems.value;
        uncatItems.forEach(item => {
            const idx = configItems.value.findIndex(i => i.id === item.id);
            if (idx !== -1) {
                configItems.value[idx].position = position++;
            }
        });

        // Update uncategorized order
        uncategorizedItemOrder.value = uncatItems.map(i => i.id);

        // Trigger reactivity for categoryItemOrders
        categoryItemOrders.value = new Map(categoryItemOrders.value);
    }

    // Handle category order change (from drag)
    function onCategoryOrderChange(newOrder: string[]) {
        categoryOrder.value = newOrder;
        recalculatePositions();
    }

    // Handle category drag update (from draggable component)
    function onCategoryDragUpdate(newList: { name: string }[]) {
        onCategoryOrderChange(newList.map(c => c.name));
    }

    // Handle items change within a category
    function onCategoryItemsChange(category: string, newItems: ConfigItem[]) {
        // Store the new order for this category
        const newOrder = newItems.map(item => item.id);
        categoryItemOrders.value.set(category, newOrder);
        categoryItemOrders.value = new Map(categoryItemOrders.value);

        // Update category for any items that were moved into this category
        newItems.forEach(item => {
            const idx = configItems.value.findIndex(i => i.id === item.id);
            if (idx !== -1) {
                configItems.value[idx].category = category;
            }
        });

        recalculatePositions();
    }

    // Handle uncategorized items change
    function onUncategorizedChange(newItems: ConfigItem[]) {
        // Store the new order for uncategorized items
        uncategorizedItemOrder.value = newItems.map(item => item.id);

        // Only update category for items that were moved FROM a category to uncategorized
        newItems.forEach(item => {
            const idx = configItems.value.findIndex(i => i.id === item.id);
            if (idx !== -1 && configItems.value[idx].category) {
                configItems.value[idx].category = undefined;
            }
        });

        recalculatePositions();
    }

    // Move item to a specific category
    function moveItemToCategory(itemId: string, category: string | undefined) {
        const idx = configItems.value.findIndex(i => i.id === itemId);
        if (idx !== -1) {
            configItems.value[idx].category = category;
            recalculatePositions();
        }
    }

    // Auto-group uncategorized items by their prefix (first segment before underscore)
    function autoGroupByPrefix() {
        // Get uncategorized items
        const uncategorized = configItems.value.filter(item => !item.category);

        // Group by prefix (first segment before underscore, uppercase)
        const prefixGroups = new Map<string, ConfigItem[]>();

        uncategorized.forEach(item => {
            const underscoreIndex = item.name.indexOf('_');
            if (underscoreIndex === -1) {
                // No underscore, skip this item
                return;
            }

            const prefix = item.name.substring(0, underscoreIndex).toUpperCase();
            if (!prefixGroups.has(prefix)) {
                prefixGroups.set(prefix, []);
            }
            prefixGroups.get(prefix)!.push(item);
        });

        // Process groups with 2+ items
        prefixGroups.forEach((items, prefix) => {
            if (items.length < 2) {
                return;
            }

            // Check if category already exists (case-insensitive match)
            const existingCategory = categories.value.find(
                cat => cat.toUpperCase() === prefix
            );

            const targetCategory = existingCategory || prefix;

            // Create category if it doesn't exist
            if (!existingCategory) {
                addCategory(targetCategory);
            }

            // Assign items to the category
            items.forEach(item => {
                const idx = configItems.value.findIndex(i => i.id === item.id);
                if (idx !== -1) {
                    configItems.value[idx].category = targetCategory;
                }
            });
        });

        recalculatePositions();
    }

    // Collapse categories by default when loading from server
    watch(categories, (newCategories) => {
        if (newCategories.length > 0 && !initialCategoriesLoaded.value) {
            collapsedCategories.value = new Set(newCategories);
            initialCategoriesLoaded.value = true;
        }
    }, { immediate: true });

    return {
        // State
        collapsedCategories,
        emptyCategories,
        categoryOrder,

        // Computed
        categories,
        hasCategories,
        categoriesForDrag,
        uncategorizedItems,

        // Functions
        getCategoryItems,
        toggleCategory,
        isCategoryCollapsed,
        addCategory,
        renameCategory,
        deleteCategory,
        recalculatePositions,
        onCategoryOrderChange,
        onCategoryDragUpdate,
        onCategoryItemsChange,
        onUncategorizedChange,
        moveItemToCategory,
        autoGroupByPrefix,
    };
}
