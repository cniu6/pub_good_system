import type { ApiReferenceConfigurationRaw } from '@scalar/types/api-reference';
import type { AuthStore } from '@scalar/workspace-store/entities/auth';
import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import { type MergedSecuritySchemes } from '@scalar/workspace-store/request-example';
import type { XScalarEnvironment } from '@scalar/workspace-store/schemas/extensions/document/x-scalar-environments';
import type { ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { WorkspaceDocument } from '@scalar/workspace-store/schemas/workspace';
type __VLS_Props = {
    options: Pick<ApiReferenceConfigurationRaw, 'authentication' | 'oauth2RedirectUri' | 'persistAuth' | 'proxyUrl'>;
    authStore: AuthStore;
    document: WorkspaceDocument | undefined;
    eventBus: WorkspaceEventBus;
    securitySchemes: MergedSecuritySchemes;
    selectedServer: ServerObject | null;
    environment: XScalarEnvironment;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=Auth.vue.d.ts.map