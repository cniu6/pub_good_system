import type { SecuritySchemeObjectSecret } from '@scalar/workspace-store/request-example';
import type { OpenApiDocument, OperationObject, ServerObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import { type RequiredSecurity } from '../../../features/Operation/helpers/get-required-security.js';
import type { OperationProps } from '../../../features/Operation/Operation.vue.js';
type __VLS_Props = Omit<OperationProps, 'document' | 'pathValue' | 'server' | 'securitySchemes' | 'authStore'> & {
    /** Operation object with path params */
    operation: OperationObject;
    /** The selected server for the operation */
    selectedServer: ServerObject | null;
    /** The selected security schemes for the operation */
    selectedSecuritySchemes: SecuritySchemeObjectSecret[];
    /** Required/optional security state for the badge next to the path */
    requiredSecurity: RequiredSecurity;
    /** The document the operation belongs to, used to resolve schema references for display */
    document?: OpenApiDocument;
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=ClassicLayout.vue.d.ts.map