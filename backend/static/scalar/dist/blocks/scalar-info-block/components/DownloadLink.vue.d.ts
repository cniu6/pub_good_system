import type { ApiReferenceConfiguration } from '@scalar/types/api-reference';
import { type WorkspaceEventBus } from '@scalar/workspace-store/events';
type __VLS_Props = {
    /** The document download type. */
    documentDownloadType: ApiReferenceConfiguration['documentDownloadType'];
    /** The event bus for the handling all events. */
    eventBus: WorkspaceEventBus;
    /**
     * URL of the OpenAPI document. Required when documentDownloadType is 'direct'
     * so the link can point to the document.
     */
    documentUrl?: string;
    /** The kind of document being rendered. Drives the button label. */
    documentType?: 'openapi' | 'asyncapi';
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=DownloadLink.vue.d.ts.map