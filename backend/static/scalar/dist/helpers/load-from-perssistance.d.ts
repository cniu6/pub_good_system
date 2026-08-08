import type { WorkspaceStore } from '@scalar/workspace-store/client';
/**
 * Loads the default HTTP client from storage and applies it to the workspace.
 * Only updates if no default client is already set. Accepts both built-in client
 * ids and custom sample ids (e.g. `custom/python`) so a custom selection persists.
 */
export declare const loadClientFromStorage: (store: WorkspaceStore) => void;
/**
 * Loads the authentication data from storage and applies it to the workspace.
 * Only updates if no authentication data is already set.
 */
export declare const loadAuthFromStorage: (store: WorkspaceStore, slug: string) => void;
//# sourceMappingURL=load-from-perssistance.d.ts.map