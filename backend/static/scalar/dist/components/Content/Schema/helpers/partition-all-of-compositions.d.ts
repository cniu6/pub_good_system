import type { SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { CompositionKeyword } from './schema-composition.js';
type ChoiceKeyword = Extract<CompositionKeyword, 'oneOf' | 'anyOf'>;
/** An ordered piece of an `allOf`: either a merged object or a choice picker. */
type AllOfSegment = {
    kind: 'object';
    schema: SchemaObject;
} | {
    kind: 'choice';
    composition: ChoiceKeyword;
    /** A minimal schema carrying only this group's composition, for SchemaComposition */
    value: SchemaObject;
    /** Ordinal among choice members, matching the request-example composition key */
    choiceIndex: number;
};
/**
 * Splits an `allOf` schema into an ordered list of segments so the renderer can
 * show each `oneOf`/`anyOf` group as its own picker **in the position it was
 * declared**, with the surrounding plain fields around it.
 *
 * `mergeAllOfSchemas` alone keeps only the FIRST `oneOf`/`anyOf` (dropping the
 * 2nd+ groups of an object with several independent mutually-exclusive
 * selections) and, being a merge, also loses the ordering between fields and
 * choices. Walking the members in order fixes both: runs of consecutive object
 * members are merged into one object segment, and each choice member becomes its
 * own picker segment in place.
 */
export declare const partitionAllOfCompositions: (schema: SchemaObject | undefined) => {
    segments: AllOfSegment[];
};
export {};
//# sourceMappingURL=partition-all-of-compositions.d.ts.map