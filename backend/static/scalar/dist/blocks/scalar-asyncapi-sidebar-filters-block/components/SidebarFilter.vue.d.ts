/**
 * SidebarFilter
 *
 * A compact picker inside the AsyncAPI sidebar filters section. Reused for the
 * stacked protocol and server filters.
 *
 * `options` is expected to lead with an "All …" entry, which is also used as the
 * fallback selection — so the first option doubles as the cleared state.
 */
type __VLS_Props = {
    /** Row label shown to the left of the dropdown (e.g. "Protocol"). */
    label: string;
    /** Filter options, leading with the "All …" entry that clears the filter. */
    options: {
        id: string;
        label: string;
    }[];
    /** The currently selected option id. */
    modelValue?: string;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {} & {
    "update:modelValue": (id: string) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:modelValue"?: ((id: string) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SidebarFilter.vue.d.ts.map