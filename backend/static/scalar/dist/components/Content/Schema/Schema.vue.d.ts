import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { DiscriminatorObject, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { SchemaOptions } from '../../../components/Content/Schema/types';
type __VLS_Props = {
    schema?: SchemaObject;
    /** Track how deep we've gone */
    level?: number;
    name?: string;
    /** A tighter layout with less borders and without a heading */
    compact?: boolean;
    /** Shows a toggle to hide/show children */
    noncollapsible?: boolean;
    /** Hide the heading */
    hideHeading?: boolean;
    /** Hide the schema description */
    hideDescription?: boolean;
    /** Show a special one way toggle for additional properties, also has a top border when open */
    additionalProperties?: boolean;
    /** Hide model names in type display */
    hideModelNames?: boolean;
    /** Discriminator object */
    discriminator?: DiscriminatorObject;
    /** Breadcrumb for the schema */
    breadcrumb?: string[];
    /** Event bus emitting actions */
    eventBus: WorkspaceEventBus | null;
    /** Move the options into a single prop so they are easy to pass around */
    options: SchemaOptions;
    /** When "requestBody", composition dropdown selection is synced with the example snippet */
    schemaContext?: string;
    /** Internal path used to sync nested request body compositions with the code sample */
    compositionPath?: string[];
    /**
     * Stable identity of this schema node, derived from its raw (unresolved)
     * value by the parent. Used to detect self-referential cycles. See
     * {@link getCycleKey}.
     */
    cycleKey?: unknown;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Schema.vue.d.ts.map