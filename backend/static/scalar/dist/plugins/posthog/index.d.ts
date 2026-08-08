import { type PostHogConfig } from '@scalar/api-client/plugins/posthog';
import type { ApiReferencePlugin } from '../plugin-manager.js';
export type { PostHogConfig };
/**
 * PostHog analytics plugin for the API Reference.
 *
 * Loading this plugin opts in to analytics for both the API Reference
 * and the embedded API Client (tracked as separate products).
 *
 * Respects the `telemetry` configuration option — when set to `false`,
 * capturing is disabled. Reacts dynamically to config changes at runtime.
 *
 * If the plugin is not loaded, no tracking occurs.
 */
export declare const PostHogPlugin: (config: PostHogConfig) => ApiReferencePlugin;
//# sourceMappingURL=index.d.ts.map