import type { ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    /** The selected server URL */
    selectedServer: ServerObject | null;
    /** Available servers */
    servers: ServerObject[];
    /** The id of the target to use for the popover (e.g. address bar) */
    target: string;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {
    servers: ServerObject[];
    serverUrlWithoutTrailingSlash: import("vue").ComputedRef<string>;
    serverOptions: import("vue").ComputedRef<{
        id: string;
        label: string;
    }[]>;
    selectedServer: ServerObject | null;
}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {} & {
    "update:modelValue": (value: string) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:modelValue"?: ((value: string) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Selector.vue.d.ts.map