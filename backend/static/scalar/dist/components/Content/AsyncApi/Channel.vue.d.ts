import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { TraversedAsyncApiChannel } from '@scalar/workspace-store/schemas/navigation';
import { type AsyncApiSchemaRenderOptions } from './helpers/async-api-render-options.js';
/** Subset of the configuration the shared `ParameterList` renderer needs. */
type ParameterListOptions = AsyncApiSchemaRenderOptions & Pick<ApiReferenceConfigurationRaw, 'hideModels'>;
type __VLS_Props = {
    channel: TraversedAsyncApiChannel;
    document: AsyncApiDocument;
    layout: 'classic' | 'modern';
    isCollapsed: boolean;
    eventBus: WorkspaceEventBus | null;
    options?: Partial<ParameterListOptions>;
    /** Map of navigation item id to expanded state, shared with the sidebar. */
    expandedItems?: Record<string, boolean>;
    /**
     * Nesting depth in the navigation tree. A channel nested inside a tag
     * (`level !== 0`) inherits the tag's horizontal padding, so it skips its own
     * `SectionContainer` padding to avoid doubling the indentation.
     */
    level?: number;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Channel.vue.d.ts.map