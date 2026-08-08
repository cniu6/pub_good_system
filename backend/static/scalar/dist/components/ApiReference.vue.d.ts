import { type AnyApiReferenceConfiguration } from '@scalar/types/api-reference';
import type { TraversedEntry } from '@scalar/workspace-store/schemas/navigation';
declare const _default: typeof __VLS_export;
export default _default;
declare const __VLS_export: __VLS_WithSlots<import("vue").DefineComponent<{
    /**
     * Configuration for the API reference.
     * Can be a single configuration or an array of configurations for multiple documents.
     */
    configuration?: AnyApiReferenceConfiguration;
}, {
    eventBus: import("@scalar/workspace-store/events").WorkspaceEventBus;
    workspaceStore: import("@scalar/workspace-store/client").WorkspaceStore;
    sidebarItems: import("vue").ComputedRef<TraversedEntry[]>;
}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<{
    /**
     * Configuration for the API reference.
     * Can be a single configuration or an array of configurations for multiple documents.
     */
    configuration?: AnyApiReferenceConfiguration;
}> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>, {
    'content-start'?(): {
        breadcrumb: string;
    };
    'content-end'?(): {
        breadcrumb: string;
    };
    'sidebar-start'?(): {
        breadcrumb: string;
    };
    'sidebar-end'?(): {
        breadcrumb: string;
    };
    'editor-placeholder'?(): {
        breadcrumb: string;
    };
    footer?(): {
        breadcrumb: string;
    };
}>;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=ApiReference.vue.d.ts.map