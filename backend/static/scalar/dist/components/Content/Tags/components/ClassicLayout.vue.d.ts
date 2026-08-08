import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { TraversedTag } from '@scalar/workspace-store/schemas/navigation';
type __VLS_Props = {
    tag: TraversedTag;
    isCollapsed: boolean;
    eventBus: WorkspaceEventBus | null;
};
declare var __VLS_36: {};
type __VLS_Slots = {} & {
    default?: (props: typeof __VLS_36) => any;
};
declare const __VLS_base: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const __VLS_export: __VLS_WithSlots<typeof __VLS_base, __VLS_Slots>;
declare const _default: typeof __VLS_export;
export default _default;
type __VLS_WithSlots<T, S> = T & {
    new (): {
        $slots: S;
    };
};
//# sourceMappingURL=ClassicLayout.vue.d.ts.map