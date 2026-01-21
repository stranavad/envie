import { parse as parseDotenv } from 'dotenv';
import type { ConfigItem } from '@/services/project.service';

/**
 * Characters that require quoting in .env files.
 * These can cause issues with shell interpretation or .env parsing.
 */
const SPECIAL_CHARS = /["\n\r\s$`!#&;|<>(){}[\]*?~\\'^]/;

/**
 * Format a value for .env file output.
 * Wraps in double quotes if the value contains special characters.
 */
export function formatEnvValue(val: string): string {
    if (SPECIAL_CHARS.test(val)) {
        // Escape backslashes first, then double quotes, then newlines
        const escaped = val
            .replace(/\\/g, '\\\\')
            .replace(/"/g, '\\"')
            .replace(/\n/g, '\\n')
            .replace(/\r/g, '\\r');
        return `"${escaped}"`;
    }
    return val;
}

/**
 * Build .env format string from config items.
 * Respects category grouping and ordering.
 */
export function buildEnvString(
    items: ConfigItem[],
    categories: string[],
    getCategoryItems: (category: string) => ConfigItem[],
    getUncategorizedItems: () => ConfigItem[]
): string {
    const lines: string[] = [];
    const hasCategories = categories.length > 0;

    if (hasCategories) {
        let isFirstSection = true;

        // Output categories first (in visual order)
        categories.forEach(category => {
            const categoryItems = getCategoryItems(category);
            if (categoryItems.length > 0) {
                // Add empty line before category (except first section)
                if (!isFirstSection) {
                    lines.push('');
                }
                isFirstSection = false;

                lines.push('#');
                lines.push(`# ${category}`);
                lines.push('#');

                categoryItems.forEach(item => {
                    lines.push(`${item.name}=${formatEnvValue(item.value)}`);
                });
            }
        });

        // Then uncategorized items at the bottom
        const uncategorized = getUncategorizedItems();
        if (uncategorized.length > 0) {
            if (!isFirstSection) {
                lines.push('');
            }
            // Only add header if there were categories above
            if (categories.some(c => getCategoryItems(c).length > 0)) {
                lines.push('#');
                lines.push('# Uncategorized');
                lines.push('#');
            }
            uncategorized.forEach(item => {
                lines.push(`${item.name}=${formatEnvValue(item.value)}`);
            });
        }
    } else {
        // No categories - just output all items in order
        items
            .slice()
            .sort((a, b) => a.position - b.position)
            .forEach(item => {
                lines.push(`${item.name}=${formatEnvValue(item.value)}`);
            });
    }

    return lines.join('\n');
}

/**
 * Copy config items to clipboard in .env format.
 * Returns true if successful.
 */
export async function copyEnvToClipboard(
    items: ConfigItem[],
    categories: string[],
    getCategoryItems: (category: string) => ConfigItem[],
    getUncategorizedItems: () => ConfigItem[]
): Promise<boolean> {
    try {
        const content = buildEnvString(items, categories, getCategoryItems, getUncategorizedItems);
        await navigator.clipboard.writeText(content);
        return true;
    } catch (e) {
        console.error('Failed to copy to clipboard', e);
        return false;
    }
}

export interface ParsedEnvItem {
    name: string;
    value: string;
}

/**
 * Parse .env format string into key-value pairs.
 * Uses dotenv library for proper multiline value support.
 */
export function parseEnvString(input: string): ParsedEnvItem[] {
    const parsed = parseDotenv(input);
    return Object.entries(parsed).map(([name, value]) => ({
        name,
        value: value ?? '',
    }));
}
