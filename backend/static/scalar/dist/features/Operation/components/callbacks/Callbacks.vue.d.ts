import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { CallbackObject, OpenApiDocument } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { OperationProps } from '../../../../features/Operation/Operation.vue.js';
type __VLS_Props = {
    path: string;
    callbacks: CallbackObject;
    eventBus: WorkspaceEventBus | null;
    /** The document the callbacks belong to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: Pick<OperationProps['options'], 'hideModels' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties'>;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Callbacks.vue.d.ts.map