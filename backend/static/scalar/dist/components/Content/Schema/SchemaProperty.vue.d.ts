import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { DiscriminatorObject, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import { type Component } from 'vue';
import type { SchemaOptions } from '../../../components/Content/Schema/types';
type __VLS_Props = {
    is?: string | Component;
    schema: SchemaObject | undefined;
    noncollapsible?: boolean;
    level?: number;
    name?: string;
    required?: boolean;
    compact?: boolean;
    discriminator?: DiscriminatorObject;
    description?: string;
    hideModelNames?: boolean;
    hideHeading?: boolean;
    /** When the root schema was resolved from a $ref, pass the ref name for display (e.g. "Data"). */
    modelName?: string | null;
    variant?: 'additionalProperties' | 'patternProperties';
    breadcrumb?: string[];
    eventBus: WorkspaceEventBus | null;
    options: SchemaOptions;
    /** Enum values for property names (from JSON Schema propertyNames keyword). */
    propertyNamesEnum?: string[];
    /** Resolved propertyNames schema, used to show key constraints like `format`. */
    propertyNamesSchema?: SchemaObject;
    /** When "requestBody", composition selection is synced with the example snippet */
    schemaContext?: string;
    /** Internal path used to sync nested request body compositions with the code sample */
    compositionPath?: string[];
    /** Internal path segment for this property when building nested composition keys */
    compositionPathSegment?: string;
    /** Stable identity of this property's schema, used for cycle detection. */
    cycleKey?: unknown;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {
    required: boolean;
    level: number;
    compact: boolean;
    hideModelNames: boolean;
}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SchemaProperty.vue.d.ts.map