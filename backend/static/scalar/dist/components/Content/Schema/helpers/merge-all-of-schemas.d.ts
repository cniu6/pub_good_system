import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Merges multiple OpenAPI schema objects into a single schema object.
 * Handles nested allOf compositions and merges properties recursively.
 *
 * @param schemas - Array of OpenAPI schema objects to merge
 * @param rootSchema - Optional root schema to merge with the result
 * @param seenRefs - `$ref` strings already being merged higher in the call stack
 * @returns Merged schema object
 */
export declare const mergeAllOfSchemas: (schemas: SchemaObject | undefined, rootSchema?: SchemaObject, seenRefs?: Set<string>) => SchemaObject;
//# sourceMappingURL=merge-all-of-schemas.d.ts.map