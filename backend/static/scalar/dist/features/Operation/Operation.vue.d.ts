import type { ClientOptionGroup } from '@scalar/blocks/code-example';
import type { HttpMethod } from '@scalar/helpers/http/http-methods';
import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { WorkspaceStore } from '@scalar/workspace-store/client';
import type { AuthStore } from '@scalar/workspace-store/entities/auth';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import { type MergedSecuritySchemes } from '@scalar/workspace-store/request-example';
import type { OpenApiDocument, PathItemObject, ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * References Operation "block"
 */
declare const _default: typeof __VLS_export;
export default _default;
export type OperationProps = {
    id: string;
    method: HttpMethod;
    /** The subset of the configuration object required for the operation component */
    options: Pick<ApiReferenceConfigurationRaw, 'expandAllResponses' | 'hideModels' | 'hideTestRequestButton' | 'layout' | 'orderRequiredPropertiesFirst' | 'orderSchemaPropertiesBy' | 'expandAllSchemaProperties' | 'showOperationId'>;
    /** Document object */
    document: OpenApiDocument;
    /** Key of the operations path in the document.paths object */
    path: string;
    /** OpenAPI path object that will include the operation */
    pathValue: PathItemObject | undefined;
    /** Currently selected server for the document */
    server: ServerObject | null;
    /** The merged security schemes for the document and the authentication configuration */
    securitySchemes: MergedSecuritySchemes;
    /** The http client options for the dropdown */
    clientOptions: ClientOptionGroup[];
    /** Whether the Classic layout operation is collapsed */
    isCollapsed: boolean;
    /** Whether the operation is a webhook */
    isWebhook: boolean;
    /** The currently selected client for the document */
    selectedClient: WorkspaceStore['workspace']['x-scalar-default-client'];
    /** The currently selected example key, shared across operations for in-sync example pickers */
    selectedExample: WorkspaceStore['workspace']['x-scalar-default-example'];
    /** The event bus */
    eventBus: WorkspaceEventBus;
    /** The auth store */
    authStore: AuthStore;
};
declare const __VLS_export: import("vue").DefineComponent<OperationProps, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<OperationProps> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
//# sourceMappingURL=Operation.vue.d.ts.map