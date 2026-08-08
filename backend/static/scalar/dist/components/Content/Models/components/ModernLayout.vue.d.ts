import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    id: string;
    name: string;
    schema: SchemaObject;
    isCollapsed: boolean;
    eventBus: WorkspaceEventBus;
    /** The document the model belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: {
        orderRequiredPropertiesFirst: boolean | undefined;
        orderSchemaPropertiesBy: 'alpha' | 'preserve' | undefined;
        hideModels: boolean | undefined;
        expandAllSchemaProperties: boolean | undefined;
    };
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ModernLayout.vue.d.ts.map