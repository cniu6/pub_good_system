import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { TraversedEntry } from '@scalar/workspace-store/schemas/navigation';
type __VLS_Props = {
    entries: TraversedEntry[];
    document: AsyncApiDocument;
    expandedItems: Record<string, boolean>;
    options: Pick<ApiReferenceConfigurationRaw, 'layout' | 'expandAllSchemaProperties' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'hideModels' | 'modelsSectionLabel'>;
    eventBus: WorkspaceEventBus;
    level?: number;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=AsyncApiTraversedEntry.vue.d.ts.map