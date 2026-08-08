import type { AsyncApiMessageObject } from '@scalar/types/asyncapi/3.1';
import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Unwraps an AsyncAPI schema-bearing value into the `SchemaObject` shape the shared OpenAPI
 * `Schema`/`Model` components expect.
 *
 * AsyncAPI reuses JSON Schema, but a value may need extra handling:
 * - The value itself may be a `$ref`; it is resolved to its `$ref-value`.
 * - A Multi Format Schema Object (`schemaFormat` plus a nested `schema`) is unwrapped to its
 *   payload, so we render the inner JSON Schema rather than the wrapper.
 * - Boolean (`true`/`false`) and non-JSON-Schema payloads are skipped.
 */
export declare const unwrapAsyncApiSchema: (value: unknown) => SchemaObject | undefined;
/** Resolves an AsyncAPI message `payload` into the `SchemaObject` the shared Schema component expects. */
export declare const getAsyncApiMessagePayloadSchema: (message: AsyncApiMessageObject) => SchemaObject | undefined;
/** Resolves an AsyncAPI message `headers` into the `SchemaObject` the shared Schema component expects. */
export declare const getAsyncApiMessageHeadersSchema: (message: AsyncApiMessageObject) => SchemaObject | undefined;
//# sourceMappingURL=get-async-api-message-payload-schema.d.ts.map