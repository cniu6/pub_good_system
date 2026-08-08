import type { Heading } from '@scalar/types/legacy';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
type __VLS_Props = {
    eventBus: WorkspaceEventBus | null;
    headingSlugGenerator: (heading: Heading) => string;
    /** Markdown document */
    description?: string;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=InfoDescription.vue.d.ts.map