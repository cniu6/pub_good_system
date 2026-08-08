import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Checks if array items have complex structure
 * like: objects, references, discriminators, compositions, or nested arrays with complex items
 *
 * @param value - The schema object to check
 * @returns true if the array has complex items, false otherwise
 */
export declare const hasComplexArrayItems: (value: SchemaObject | undefined) => boolean;
//# sourceMappingURL=has-complex-array-items.d.ts.map