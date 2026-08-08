import type { AsyncApiServerEntry } from '@scalar/workspace-store/channel-example';
type __VLS_Props = {
    /** The selected server */
    selectedServer: AsyncApiServerEntry | null;
    /** Available servers */
    servers: AsyncApiServerEntry[];
    /** The id of the target to use for the popover (e.g. address bar) */
    target: string;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {
    servers: AsyncApiServerEntry[];
    serverUrlWithoutTrailingSlash: import("vue").ComputedRef<string>;
    serverOptions: import("vue").ComputedRef<{
        id: string;
        label: string;
    }[]>;
    selectedServer: AsyncApiServerEntry | null;
}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {} & {
    "update:modelValue": (value: string) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:modelValue"?: ((value: string) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Selector.vue.d.ts.map