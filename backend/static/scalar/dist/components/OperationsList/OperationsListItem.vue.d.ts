import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { TraversedOperation, TraversedWebhook } from '@scalar/workspace-store/schemas/navigation';
type __VLS_Props = {
    operation: TraversedOperation | TraversedWebhook;
    isCollapsed?: boolean;
    eventBus: WorkspaceEventBus | null;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=OperationsListItem.vue.d.ts.map