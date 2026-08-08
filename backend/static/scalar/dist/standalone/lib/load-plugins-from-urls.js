//#region src/standalone/lib/load-plugins-from-urls.ts
/**
* Import a plugin module with a browser-native dynamic `import()`.
*
* The specifier is fully dynamic, so bundlers can't analyze it and leave the `import()` call
* as-is in the output (including the UMD standalone bundle) — the `@vite-ignore` comment just
* silences the warning about that.
*/
var importModule = (url) => import(
	/* @vite-ignore */
	url
);
/**
* Load a single plugin module and return its default export.
*
* Failures (network errors, modules without a function default export) are logged and swallowed
* so a broken plugin URL never prevents the API reference itself from mounting.
*/
var loadPluginFromUrl = async (url) => {
	try {
		const plugin = (await importModule(url))?.default;
		if (typeof plugin !== "function") {
			console.error(`[@scalar/api-reference] The module at ${url} does not export an API Reference plugin as its default export.`);
			return;
		}
		return plugin;
	} catch (error) {
		console.error(`[@scalar/api-reference] Failed to load the plugin module at ${url}:`, error);
		return;
	}
};
/** Normalize the configuration input to a list of configuration objects */
var getConfigurations = (configuration) => Array.isArray(configuration) ? configuration : [configuration];
/** Whether any of the passed configurations reference a plugin by URL */
var hasPluginUrls = (configuration) => getConfigurations(configuration).some((config) => Boolean(config.pluginUrls?.length));
/**
* Resolve the `pluginUrls` of all passed configurations and append the loaded plugins to the
* respective configuration's `plugins`.
*
* Plugin registration is not reactive — plugins are read once when the API reference renders for
* the first time — so this must complete before the app is mounted. Each URL is imported only
* once, even when multiple configurations reference it.
*/
var loadPluginsFromUrls = async (configuration) => {
	const pendingImports = /* @__PURE__ */ new Map();
	const importOnce = (url) => {
		const pending = pendingImports.get(url) ?? loadPluginFromUrl(url);
		pendingImports.set(url, pending);
		return pending;
	};
	await Promise.all(getConfigurations(configuration).map(async (config) => {
		if (!config.pluginUrls?.length) return;
		const plugins = (await Promise.all(config.pluginUrls.map(importOnce))).filter((plugin) => plugin !== void 0);
		if (plugins.length) config.plugins = [...config.plugins ?? [], ...plugins];
	}));
};
//#endregion
export { hasPluginUrls, loadPluginsFromUrls };

//# sourceMappingURL=load-plugins-from-urls.js.map