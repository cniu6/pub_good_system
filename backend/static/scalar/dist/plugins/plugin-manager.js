//#region src/plugins/plugin-manager.ts
/**
* A no-op auth state used when no accessor is provided (e.g. in tests or a standalone manager).
* It reports an empty authentication state so plugins can call the read methods unconditionally.
*/
var createEmptyAuthState = () => ({
	export: () => ({}),
	getAuthSecrets: () => void 0,
	getAuthSelectedSchemas: () => void 0
});
/** Plugin view slots that can render custom components in the content area */
var PLUGIN_VIEW_NAMES = ["content.start", "content.end"];
/**
* Build a stable, unique id for a plugin view component.
*
* The same id is used as the DOM element id (so scroll navigation can find it) and as the
* sidebar navigation entry id (so clicking it scrolls to the element). Keeping both in sync
* is what lets plugin views participate in the existing scroll-spy and navigation logic.
*
* The id is prefixed with the document slug so it matches the navigation id convention. That is
* what lets URL deep-linking work: `getIdFromUrl` re-prepends the document slug on initial load,
* so a hash like `plugin-view/<plugin>/<view>/<index>` resolves back to this exact id.
*/
var getPluginViewId = (documentSlug, pluginName, viewName, index) => `${documentSlug}/plugin-view/${pluginName}/${viewName}/${index}`;
/**
* Create the plugin manager store
*
* This store manages all plugins registered with the API client
*/
var createPluginManager = ({ plugins = [], auth }) => {
	const registeredPlugins = /* @__PURE__ */ new Map();
	const authState = auth ?? createEmptyAuthState();
	plugins.forEach((plugin) => {
		const pluginInstance = plugin();
		registeredPlugins.set(pluginInstance.name, pluginInstance);
	});
	return {
		/**
		* Get all extensions with the given name from registered plugins
		*/
		getSpecificationExtensions: (name) => {
			const extensions = [];
			for (const plugin of registeredPlugins.values()) for (const extension of plugin.extensions) if (extension.name === name) extensions.push(extension);
			return extensions;
		},
		/**
		* Get all components for a specific view from registered plugins.
		*
		* Each component carries a stable `id` (scoped to the active document slug) so the rendered
		* DOM element and the sidebar entry (see `getSidebarEntries`) share the same id and stay in
		* sync for scroll navigation and deep-linking.
		*/
		getViewComponents: (viewName, documentSlug) => {
			const components = [];
			for (const plugin of registeredPlugins.values()) {
				const viewComponents = plugin.views?.[viewName];
				if (viewComponents) viewComponents.forEach((component, index) => {
					components.push({
						...component,
						id: getPluginViewId(documentSlug, plugin.name, viewName, index)
					});
				});
			}
			return components;
		},
		/**
		* Notify all plugins that the API Reference has been initialized
		*/
		notifyInit: (config) => {
			for (const plugin of registeredPlugins.values()) plugin.hooks?.onInit?.({
				config,
				auth: authState
			});
		},
		/**
		* Notify all plugins that the configuration has changed
		*/
		notifyConfigChange: (config) => {
			for (const plugin of registeredPlugins.values()) plugin.hooks?.onConfigChange?.({
				config,
				auth: authState
			});
		},
		/**
		* Get the read-only accessor for the global authentication state.
		*
		* Plugin view components can call this (via `usePluginManager`) to read stored secrets and the
		* selected security schemes. Returns an empty auth state when no accessor was provided.
		*/
		getAuthState: () => authState,
		/**
		* Notify all plugins that the API Reference is being destroyed
		*/
		notifyDestroy: () => {
			for (const plugin of registeredPlugins.values()) plugin.hooks?.onDestroy?.();
		},
		/**
		* Get all client plugins provided by registered plugins
		*/
		getApiClientPlugins: () => {
			const apiClientPlugins = [];
			for (const plugin of registeredPlugins.values()) if (plugin.apiClientPlugins) apiClientPlugins.push(...plugin.apiClientPlugins);
			return apiClientPlugins;
		},
		/**
		* Get all sidebar entries contributed by plugin views.
		*
		* Only views that opt in via `sidebar.show` are returned. Each entry's `id` matches the
		* id of the rendered component (see `getViewComponents`), so the API Reference can add it
		* to the sidebar navigation and scrolling/active-tracking work out of the box.
		*/
		getSidebarEntries: (documentSlug) => {
			const entries = [];
			for (const plugin of registeredPlugins.values()) for (const viewName of PLUGIN_VIEW_NAMES) (plugin.views?.[viewName])?.forEach((component, index) => {
				if (component.sidebar?.show && component.sidebar.label) entries.push({
					id: getPluginViewId(documentSlug, plugin.name, viewName, index),
					label: component.sidebar.label,
					viewName
				});
			});
			return entries;
		}
	};
};
//#endregion
export { createPluginManager };

//# sourceMappingURL=plugin-manager.js.map