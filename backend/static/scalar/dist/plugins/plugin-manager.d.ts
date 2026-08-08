import type { ClientPlugin } from '@scalar/oas-utils/helpers';
import type { ApiReferencePlugin as OriginalApiReferencePlugin, PluginAuthState, SpecificationExtension, ViewComponent } from '@scalar/types/api-reference';
export type ApiReferencePlugin = OriginalApiReferencePlugin;
export type { PluginAuthState };
type CreatePluginManagerParams = {
    plugins?: ApiReferencePlugin[];
    /**
     * Read-only accessor for the global authentication state.
     *
     * Passed through to plugin lifecycle hooks and exposed via `getAuthState`, letting plugins read
     * stored secrets and the selected security schemes without being able to mutate them.
     */
    auth?: PluginAuthState;
};
/** Plugin view slots that can render custom components in the content area */
declare const PLUGIN_VIEW_NAMES: readonly ["content.start", "content.end"];
type PluginViewName = (typeof PLUGIN_VIEW_NAMES)[number];
/** A plugin view component paired with the stable id used for the DOM and sidebar navigation */
export type PluginViewComponent = ViewComponent & {
    id: string;
};
/** A sidebar entry contributed by a plugin view */
type PluginSidebarEntry = {
    id: string;
    label: string;
    viewName: PluginViewName;
};
/**
 * Create the plugin manager store
 *
 * This store manages all plugins registered with the API client
 */
export declare const createPluginManager: ({ plugins, auth }: CreatePluginManagerParams) => {
    /**
     * Get all extensions with the given name from registered plugins
     */
    getSpecificationExtensions: (name: `x-${string}`) => SpecificationExtension[];
    /**
     * Get all components for a specific view from registered plugins.
     *
     * Each component carries a stable `id` (scoped to the active document slug) so the rendered
     * DOM element and the sidebar entry (see `getSidebarEntries`) share the same id and stay in
     * sync for scroll navigation and deep-linking.
     */
    getViewComponents: (viewName: PluginViewName, documentSlug: string) => PluginViewComponent[];
    /**
     * Notify all plugins that the API Reference has been initialized
     */
    notifyInit: (config: Record<string, unknown>) => void;
    /**
     * Notify all plugins that the configuration has changed
     */
    notifyConfigChange: (config: Record<string, unknown>) => void;
    /**
     * Get the read-only accessor for the global authentication state.
     *
     * Plugin view components can call this (via `usePluginManager`) to read stored secrets and the
     * selected security schemes. Returns an empty auth state when no accessor was provided.
     */
    getAuthState: () => PluginAuthState;
    /**
     * Notify all plugins that the API Reference is being destroyed
     */
    notifyDestroy: () => void;
    /**
     * Get all client plugins provided by registered plugins
     */
    getApiClientPlugins: () => ClientPlugin[];
    /**
     * Get all sidebar entries contributed by plugin views.
     *
     * Only views that opt in via `sidebar.show` are returned. Each entry's `id` matches the
     * id of the rendered component (see `getViewComponents`), so the API Reference can add it
     * to the sidebar navigation and scrolling/active-tracking work out of the box.
     */
    getSidebarEntries: (documentSlug: string) => PluginSidebarEntry[];
};
export type PluginManager = ReturnType<typeof createPluginManager>;
//# sourceMappingURL=plugin-manager.d.ts.map