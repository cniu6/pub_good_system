import type { AsyncApiServerEntry } from '@scalar/workspace-store/channel-example';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
type SelectorProps = {
    /** The event bus to use for emitting events */
    eventBus: WorkspaceEventBus;
    /** The currently selected server */
    selectedServer: AsyncApiServerEntry | null;
    /** Available servers */
    servers: AsyncApiServerEntry[];
};
/**
 * AsyncApiServerSelector
 *
 * Core component for rendering an AsyncAPI server selector block. It mirrors the
 * OpenAPI ServerSelector, but works with the AsyncAPI server shape (a named map
 * of `host`/`protocol`/`pathname` rather than an array of `url`).
 *
 * @event asyncapi-server:update:selected - Emitted when the selected server changes
 * @event asyncapi-server:update:variables - Emitted when a server variable changes
 */
declare const _default: typeof __VLS_export;
export default _default;
declare const __VLS_export: import("vue").DefineComponent<SelectorProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<SelectorProps> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
//# sourceMappingURL=AsyncApiServerSelector.vue.d.ts.map