import type { SchemaObject, SchemaReferenceType } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Extract schema name from various schema formats
 *
 * Handles $ref, title, name, type, and schema dictionary lookup
 */
export declare const getModelNameFromSchema: (schemaOrRef: SchemaObject | SchemaReferenceType<SchemaObject>) => {
    /** The key in components.schemas (extracted from $ref), used for sidebar navigation. */
    schemaKey: string | null;
    /** The human-readable name to display (schema.title, schema.name, or ref key). */
    label: string;
} | null;
//# sourceMappingURL=schema-name.d.ts.map