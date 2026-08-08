import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type SelectorProps = {
    /** The event bus to use for emitting events */
    eventBus: WorkspaceEventBus;
    /** The selected server  */
    selectedServer: ServerObject | null;
    /** Available servers */
    servers: ServerObject[];
};
/**
 * ServerSelector
 *
 * Core component for rendering a server selector block.
 * Handles server selection and emits a 'server:update:selected' event when the selected server changes.
 *
 * @event server:update:selected - Emitted when the selected server changes
 * @event server:update:variables - Emitted when a server variable changes
 */
declare const _default: typeof __VLS_export;
export default _default;
declare const __VLS_export: import("vue").DefineComponent<SelectorProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<SelectorProps> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
//# sourceMappingURL=ServerSelector.vue.d.ts.map