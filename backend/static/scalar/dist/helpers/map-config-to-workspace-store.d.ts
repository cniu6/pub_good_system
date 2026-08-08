import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { WorkspaceStore } from '@scalar/workspace-store/client';
import { type MaybeRefOrGetter, type Ref } from 'vue';
export declare const mapConfigToWorkspaceStore: ({ config, store, isDarkMode, }: {
    config: MaybeRefOrGetter<ApiReferenceConfigurationRaw>;
    store: WorkspaceStore;
    isDarkMode: Ref<boolean>;
}) => void;
//# sourceMappingURL=map-config-to-workspace-store.d.ts.map