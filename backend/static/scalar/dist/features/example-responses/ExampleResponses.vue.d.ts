import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { ResponsesObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * TODO: copyToClipboard isn't using the right content if there are multiple examples
 */
type __VLS_Props = {
    responses: ResponsesObject;
    /**
     * The document-wide selected example key. Honored only when the current response defines an
     * example with the same key, so response example pickers stay in sync between operations without
     * blanking out responses that do not share that key.
     */
    selectedExample?: string;
    /** Event bus, used to broadcast the selected example so other operations can follow */
    eventBus?: WorkspaceEventBus;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ExampleResponses.vue.d.ts.map