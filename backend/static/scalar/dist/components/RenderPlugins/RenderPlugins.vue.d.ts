import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
type __VLS_Props = {
    viewName: 'content.start' | 'content.end';
    options: Record<string, any>;
    eventBus?: WorkspaceEventBus;
    /** Slug of the active document, used to scope plugin view ids for navigation and deep-linking */
    documentSlug: string;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=RenderPlugins.vue.d.ts.map