import type { WorkspacePlugin } from '@scalar/workspace-store/workspace-plugin';
/**
 * Plugin to persist workspace state changes with debounced writes.
 */
export declare const persistencePlugin: ({ debounceDelay, maxWait, persistAuth, }: {
    debounceDelay?: number;
    /** Maximum time in milliseconds to wait before forcing execution, even with continuous calls. */
    maxWait?: number;
    /**
     * Determines whether authentication details should be persisted.
     * Accepts a boolean or a function that returns a boolean.
     * Allows for conditional persistence logic, e.g., based on environment or user settings.
     */
    persistAuth?: boolean | (() => boolean);
}) => WorkspacePlugin;
//# sourceMappingURL=persistence-plugin.d.ts.map