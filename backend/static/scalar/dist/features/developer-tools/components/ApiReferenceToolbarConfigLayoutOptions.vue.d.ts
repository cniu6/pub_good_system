import { type ApiReferenceConfiguration } from '@scalar/types/api-reference';
type LayoutOptions = {
    showSidebar?: boolean;
    defaultOpenFirstTag?: boolean;
    defaultOpenAllTags?: boolean;
    expandAllModelSections?: boolean;
    expandAllResponses?: boolean;
    hideClientButton?: boolean;
    hideDarkModeToggle?: boolean;
    hideModels?: boolean;
    hideSearch?: boolean;
    hideTestRequestButton?: boolean;
    showOperationId?: boolean;
};
type __VLS_Props = {
    configuration?: Partial<ApiReferenceConfiguration>;
};
type __VLS_ModelProps = {
    modelValue?: LayoutOptions;
};
type __VLS_PublicProps = __VLS_Props & __VLS_ModelProps;
declare const __VLS_export: import("vue").DefineComponent<__VLS_PublicProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:modelValue": (value: LayoutOptions) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_PublicProps> & Readonly<{
    "onUpdate:modelValue"?: ((value: LayoutOptions) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ApiReferenceToolbarConfigLayoutOptions.vue.d.ts.map