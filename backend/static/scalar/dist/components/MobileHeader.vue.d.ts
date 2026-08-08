type __VLS_Props = {
    breadcrumb: string;
    isSidebarOpen: boolean;
    showSidebar: boolean;
};
type __VLS_Slots = {
    actions?(): never;
    sidebar?(props: {
        sidebarClasses: string;
    }): never;
    search?(): never;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {} & {
    toggleSidebar: () => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    onToggleSidebar?: (() => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=MobileHeader.vue.d.ts.map