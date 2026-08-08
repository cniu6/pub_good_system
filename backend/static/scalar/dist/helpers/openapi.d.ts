import type { OpenApiDocument, OperationObject, ParameterObject, ReferenceType, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Extracts the names of every parameter on an operation.
 *
 * The returned strings contain only parameter names (e.g. `userId`, `limit`) so they can be indexed
 * as a high-signal field for search. Filter-style metadata like `REQUIRED`, `optional`, `query` and
 * the schema type are intentionally excluded — those tokens dilute fuzzy matches and produce false
 * positives for queries like `query` or `integer`.
 */
export declare function extractParameterNames(parameters: ReferenceType<ParameterObject>[]): string[];
/**
 * Extracts the descriptions of every parameter on an operation.
 *
 * Kept separate from parameter names so the search index can weight each independently.
 */
export declare function extractParameterDescriptions(parameters: ReferenceType<ParameterObject>[]): string[];
/**
 * Extracts the names of properties from the request body schema(s) of an operation.
 *
 * Walks every media type and includes both top-level and one level of nested property names so
 * common fields like `email` or `username` surface in search regardless of how the body is shaped.
 */
export declare function extractBodyFieldNames(operation: OperationObject): string[];
/**
 * Extracts the descriptions of properties from the request body schema(s) of an operation.
 */
export declare function extractBodyDescriptions(operation: OperationObject): string[];
/**
 * Extracts the property names of a schema for the search index.
 *
 * Same depth and composition behavior as `extractBodyFieldNames` — descends transparently through
 * `oneOf`/`anyOf`/`allOf`, walks one level into nested object properties, dedupes.
 */
export declare function extractSchemaFieldNames(schema: SchemaObject | undefined): string[];
/**
 * Extracts the property descriptions of a schema for the search index.
 */
export declare function extractSchemaDescriptions(schema: SchemaObject | undefined): string[];
/**
 * Deep merge for objects.
 *
 * OpenAPI documents are untrusted input and can carry keys like `__proto__`, `constructor`, or
 * `prototype` (a schema is free to describe a property named `constructor`). Those keys are merged
 * as ordinary own data, so nothing is silently dropped, but the merge never writes through the
 * prototype chain, which would otherwise pollute `Object.prototype` for every object on the page.
 */
export declare function deepMerge(source: Record<any, any>, target: Record<any, any>): Record<any, any>;
/**
 * Creates an empty specification object.
 * The returning object has the same structure as a valid OpenAPI specification, but everything is empty.
 */
export declare function createEmptySpecification(partialSpecification?: Partial<OpenApiDocument>): OpenApiDocument;
//# sourceMappingURL=openapi.d.ts.map