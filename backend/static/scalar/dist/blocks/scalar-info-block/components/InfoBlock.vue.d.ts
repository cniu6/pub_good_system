import type { ApiReferenceConfiguration } from '@scalar/types/api-reference';
import type { AsyncApiInfoObject } from '@scalar/types/asyncapi/3.1';
import type { Heading } from '@scalar/types/legacy';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { ExternalDocumentationObject, InfoObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    /** Optional unique identifier for the info block. */
    id?: string;
    /** Original specification version of the input document (OpenAPI or AsyncAPI). */
    specificationVersion?: string;
    /** The Info object from the OpenAPI document. */
    info: InfoObject | AsyncApiInfoObject | undefined;
    /** The external documentation object from the OpenAPI document, if present. */
    externalDocs?: ExternalDocumentationObject;
    /** OpenAPI extension fields at the document level. */
    documentExtensions?: Record<string, unknown>;
    /** OpenAPI extension fields at the info object level. */
    infoExtensions?: Record<string, unknown>;
    /** The event bus for the handling all events. */
    eventBus: WorkspaceEventBus;
    /** Heading id generator for Markdown headings */
    headingSlugGenerator: (heading: Heading) => string;
    /** Determines the layout style for the info block ('modern' or 'classic'). */
    layout?: 'modern' | 'classic';
    /** The document download type. */
    documentDownloadType?: ApiReferenceConfiguration['documentDownloadType'];
    /** URL of the OpenAPI document. Used when documentDownloadType is 'direct'. */
    documentUrl?: string;
    /** The kind of document being rendered. Drives download button labels. */
    documentType?: 'openapi' | 'asyncapi';
};
declare var __VLS_15: {};
type __VLS_Slots = {} & {
    selectors?: (props: typeof __VLS_15) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=InfoBlock.vue.d.ts.map