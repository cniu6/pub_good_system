import type { AnyApiReferenceConfiguration } from '@scalar/types/api-reference';
/** Whether any of the passed configurations reference a plugin by URL */
export declare const hasPluginUrls: (configuration: AnyApiReferenceConfiguration) => boolean;
/**
 * Resolve the `pluginUrls` of all passed configurations and append the loaded plugins to the
 * respective configuration's `plugins`.
 *
 * Plugin registration is not reactive — plugins are read once when the API reference renders for
 * the first time — so this must complete before the app is mounted. Each URL is imported only
 * once, even when multiple configurations reference it.
 */
export declare const loadPluginsFromUrls: (configuration: AnyApiReferenceConfiguration) => Promise<void>;
//# sourceMappingURL=load-plugins-from-urls.d.ts.map