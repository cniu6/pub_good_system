import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { OpenApiDocument, RequestBodyObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    breadcrumb?: string[];
    requestBody?: RequestBodyObject;
    eventBus: WorkspaceEventBus | null;
    /** The document the request body belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
    options: {
        orderRequiredPropertiesFirst: boolean | undefined;
        orderSchemaPropertiesBy: 'alpha' | 'preserve' | undefined;
        hideModels: boolean | undefined;
        expandAllSchemaProperties: boolean | undefined;
    };
};
type __VLS_ModelProps = {
    'selectedContentType'?: string;
};
type __VLS_PublicProps = __VLS_Props & __VLS_ModelProps;
declare var __VLS_1: {};
type __VLS_Slots = {} & {
    title?: (props: typeof __VLS_1) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_PublicProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:selectedContentType": (value: string) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_PublicProps> & Readonly<{
    "onUpdate:selectedContentType"?: ((value: string) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=RequestBody.vue.d.ts.map