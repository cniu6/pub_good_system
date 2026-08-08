import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { TraversedAsyncApiOperation } from '@scalar/workspace-store/schemas/navigation';
import type { AsyncApiSchemaRenderOptions } from './helpers/async-api-render-options.js';
/** Subset of the configuration the nested `Message`/`Schema` renderers need. */
type OperationOptions = AsyncApiSchemaRenderOptions;
type __VLS_Props = {
    operation: TraversedAsyncApiOperation;
    document: AsyncApiDocument;
    eventBus: WorkspaceEventBus | null;
    options?: Partial<OperationOptions>;
    /** Map of navigation item id to expanded state, shared with the sidebar. */
    expandedItems?: Record<string, boolean>;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Operation.vue.d.ts.map