import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument, ParameterObject, ResponseObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { OperationProps } from '../../../features/Operation/Operation.vue.js';
type __VLS_Props = {
    parameter: ParameterObject | ResponseObject;
    name: string;
    breadcrumb?: string[];
    eventBus: WorkspaceEventBus | null;
    collapsableItems?: boolean;
    /** The document the operation belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: Pick<OperationProps['options'], 'hideModels' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties'>;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ParameterListItem.vue.d.ts.map