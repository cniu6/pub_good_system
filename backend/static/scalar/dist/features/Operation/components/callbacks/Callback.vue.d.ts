import type { HttpMethod as HttpMethodType } from '@scalar/helpers/http/http-methods';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import { type OpenApiDocument, type OperationObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { OperationProps } from '../../../../features/Operation/Operation.vue.js';
type __VLS_Props = {
    callback: OperationObject;
    method: HttpMethodType;
    name: string;
    url: string;
    eventBus: WorkspaceEventBus | null;
    /** The document the callback belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: Pick<OperationProps['options'], 'hideModels' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties'>;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Callback.vue.d.ts.map