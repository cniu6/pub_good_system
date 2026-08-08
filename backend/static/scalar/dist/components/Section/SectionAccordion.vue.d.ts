type __VLS_Props = {
    transparent?: boolean;
    modelValue: boolean;
};
declare var __VLS_18: {}, __VLS_20: {
    active: boolean;
}, __VLS_33: {}, __VLS_35: {};
type __VLS_Slots = {} & {
    title?: (props: typeof __VLS_18) => any;
} & {
    actions?: (props: typeof __VLS_20) => any;
} & {
    description?: (props: typeof __VLS_33) => any;
} & {
    default?: (props: typeof __VLS_35) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {} & {
    "update:modelValue": (value: boolean) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:modelValue"?: ((value: boolean) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=SectionAccordion.vue.d.ts.map