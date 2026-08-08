import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { HeaderObject, OpenApiDocument } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    headers: Record<string, HeaderObject>;
    breadcrumb?: string[];
    eventBus: WorkspaceEventBus | null;
    /** The document the headers belong to, used to resolve schema references for display */
    document?: OpenApiDocument;
    orderRequiredPropertiesFirst: boolean | undefined;
    orderSchemaPropertiesBy: 'alpha' | 'preserve' | undefined;
    expandAllSchemaProperties: boolean | undefined;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Headers.vue.d.ts.map