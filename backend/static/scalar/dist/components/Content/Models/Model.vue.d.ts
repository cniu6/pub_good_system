import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    id: string;
    name: string;
    options: Pick<ApiReferenceConfigurationRaw, 'layout' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties' | 'hideModels'>;
    schema: SchemaObject | undefined;
    isCollapsed: boolean;
    eventBus: WorkspaceEventBus;
    /** The document the model belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Model.vue.d.ts.map