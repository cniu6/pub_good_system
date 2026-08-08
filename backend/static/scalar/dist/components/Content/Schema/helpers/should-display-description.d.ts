import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Determine if description should be displayed
 *
 * @param schema - The schema object to check
 * @param propDescription - Optional description from props
 * @returns Description string to display, or null if should not be displayed
 */
export declare const shouldDisplayDescription: (schema: SchemaObject | undefined, propDescription?: string) => string | null;
//# sourceMappingURL=should-display-description.d.ts.map