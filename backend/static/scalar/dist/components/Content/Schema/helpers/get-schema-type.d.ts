import type { ReferenceType, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Computes the structural type for a schema.
 * This helper always returns type information, never schema titles or ref names.
 *
 * Priority order:
 * 1. const values
 * 2. Array types (with special handling for items)
 * 3. type with contentEncoding
 * 4. raw type
 */
export declare const getSchemaType: (valueOrRef: SchemaObject | ReferenceType<SchemaObject>) => string;
//# sourceMappingURL=get-schema-type.d.ts.map