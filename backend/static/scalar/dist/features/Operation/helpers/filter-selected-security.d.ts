import type { SelectedSecurity } from '@scalar/workspace-store/entities/auth';
import type { SecuritySchemeObjectSecret } from '@scalar/workspace-store/request-example';
import { type MergedSecuritySchemes } from '@scalar/workspace-store/request-example';
import type { OpenApiDocument, OperationObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Find the intersection between which security is selected on the document and what this operation requires
 *
 * If there is no overlap, we return the first requirement
 */
export declare const filterSelectedSecurity: (document: OpenApiDocument, operation: OperationObject | null, selectedSecurityDocument?: SelectedSecurity, selectedSecurityOperation?: SelectedSecurity, securitySchemes?: MergedSecuritySchemes) => SecuritySchemeObjectSecret[];
//# sourceMappingURL=filter-selected-security.d.ts.map