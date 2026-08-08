import { type ModalState } from '@scalar/components/modal';
import type { ModelsSectionLabel } from '@scalar/types/api-reference';
import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    modalState: ModalState;
    document: OpenApiDocument | AsyncApiDocument | undefined;
    eventBus: WorkspaceEventBus;
    modelsSectionLabel?: ModelsSectionLabel;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SearchModal.vue.d.ts.map