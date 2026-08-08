import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Determine if property heading should be displayed
 *
 * @param schema - The schema object to check
 * @param name - Optional property name
 * @param required - Whether the property is required
 * @returns true if heading should be displayed, false otherwise
 */
export declare const shouldDisplayHeading: (schema: SchemaObject | undefined, name?: string, required?: boolean) => boolean;
//# sourceMappingURL=should-display-heading.d.ts.map