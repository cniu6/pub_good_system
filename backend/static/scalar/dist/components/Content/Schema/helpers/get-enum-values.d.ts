import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Extract enum values from schema or array items
 *
 * @param value - The schema object to extract enum values from
 * @returns Array of enum values, or empty array if no enum found
 */
export declare const getEnumValues: (value: SchemaObject | undefined) => unknown[];
//# sourceMappingURL=get-enum-values.d.ts.map