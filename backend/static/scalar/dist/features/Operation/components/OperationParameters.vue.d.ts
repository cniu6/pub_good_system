import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument, ParameterObject, ReferenceType, RequestBodyObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { OperationProps } from '../../../features/Operation/Operation.vue.js';
type __VLS_Props = {
    breadcrumb?: string[];
    parameters?: ReferenceType<ParameterObject>[];
    requestBody?: RequestBodyObject | undefined;
    eventBus: WorkspaceEventBus | null;
    /** The document the operation belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: Pick<OperationProps['options'], 'hideModels' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties'>;
};
type __VLS_ModelProps = {
    /** Thread the selected request body content type up to the layout */
    'selectedContentType'?: string;
};
type __VLS_PublicProps = __VLS_Props & __VLS_ModelProps;
declare const __VLS_export: import("vue").DefineComponent<__VLS_PublicProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:selectedContentType": (value: string | undefined) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_PublicProps> & Readonly<{
    "onUpdate:selectedContentType"?: ((value: string | undefined) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=OperationParameters.vue.d.ts.map