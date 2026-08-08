import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { DiscriminatorObject, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { SchemaOptions } from '../../../components/Content/Schema/types';
import { type CompositionKeyword } from './helpers/schema-composition.js';
type __VLS_Props = {
    /** The composition keyword (oneOf, anyOf, allOf) */
    composition: CompositionKeyword;
    /** Optional discriminator object for polymorphic schemas */
    discriminator?: DiscriminatorObject;
    /** Optional name for the schema */
    name?: string;
    /** The schema value containing the composition */
    schema: SchemaObject;
    /** Nesting level for proper indentation */
    level: number;
    /** Whether to use compact layout */
    compact?: boolean;
    /** Whether to hide the heading */
    hideHeading?: boolean;
    /** Hide model names in type display */
    hideModelNames?: boolean;
    /** Breadcrumb for navigation */
    breadcrumb?: string[];
    /** Event bus emitting actions */
    eventBus: WorkspaceEventBus | null;
    /** Move the options into  single prop so they are easy to pass around */
    options: SchemaOptions;
    /** When "requestBody", sync selected index with the example snippet */
    schemaContext?: string;
    /** Internal path used to sync nested request body compositions with the code sample */
    compositionPath?: string[];
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {
    compact: boolean;
    hideHeading: boolean;
}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SchemaComposition.vue.d.ts.map