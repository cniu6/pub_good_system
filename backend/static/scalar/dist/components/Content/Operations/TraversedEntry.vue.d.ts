import type { ClientOptionGroup } from '@scalar/blocks/code-example';
import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { WorkspaceStore } from '@scalar/workspace-store/client';
import type { AuthStore } from '@scalar/workspace-store/entities/auth';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { MergedSecuritySchemes } from '@scalar/workspace-store/request-example';
import type { TraversedEntry } from '@scalar/workspace-store/schemas/navigation';
import type { OpenApiDocument, ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type __VLS_Props = {
    /** The auth store */
    authStore: AuthStore;
    /** The level of depth */
    level?: number;
    /** Traversed entries to render */
    entries: TraversedEntry[];
    /** The document object */
    document: OpenApiDocument;
    /** The http client options for the dropdown */
    clientOptions: ClientOptionGroup[];
    /** The subset of the configuration object required for the operation component */
    options: Pick<ApiReferenceConfigurationRaw, 'expandAllResponses' | 'hideTestRequestButton' | 'layout' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties' | 'showOperationId' | 'hideModels' | 'modelsSectionLabel'>;
    /** Currently selected server for the document */
    selectedServer: ServerObject | null;
    /** The merged security schemes for the document and the authentication configuration */
    securitySchemes: MergedSecuritySchemes;
    /** Currently selected http client for the document */
    selectedClient: WorkspaceStore['workspace']['x-scalar-default-client'];
    /** Currently selected example key, shared across operations for in-sync example pickers */
    selectedExample: WorkspaceStore['workspace']['x-scalar-default-example'];
    /** Used to determine if an entry is collapsed */
    expandedItems: Record<string, boolean>;
    /** The event bus for the handling all events. */
    eventBus: WorkspaceEventBus;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=TraversedEntry.vue.d.ts.map