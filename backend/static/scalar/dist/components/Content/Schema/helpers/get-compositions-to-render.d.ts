import type { OpenApiDocument, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import { type CompositionKeyword } from './schema-composition.js';
type CompositionToRender = {
    composition: CompositionKeyword;
    value: SchemaObject;
};
type DocumentSchemaLookup = Pick<OpenApiDocument, 'components'>;
/**
 * Builds a synthetic `oneOf` composition from a `discriminator.mapping` when the
 * schema declares a mapping but no explicit `oneOf`/`anyOf`. This is the shape
 * NSwag emits for polymorphic types, where the base type is a plain object with
 * a discriminator mapping pointing at the concrete variants.
 *
 * Returns `null` when there is nothing to infer (an explicit composition is
 * already present, no document to resolve refs against, or no resolvable refs).
 */
export declare const inferDiscriminatorMappingComposition: (value: SchemaObject, document?: DocumentSchemaLookup) => SchemaObject | null;
/**
 * Computes which compositions should be rendered and with which values
 *
 * @param value - The schema object to check for compositions
 * @returns Array of compositions to render with their values
 */
export declare const getCompositionsToRender: (value: SchemaObject | undefined, document?: DocumentSchemaLookup) => CompositionToRender[];
export {};
//# sourceMappingURL=get-compositions-to-render.d.ts.map