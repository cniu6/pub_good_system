import type { ParameterObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Deep object query parameters serialize as name[prop] pairs in URLs.
 * Rendering the same shape keeps docs aligned with the request UI.
 */
export declare const flattenDeepObjectQueryParameter: (parameter: ParameterObject) => ParameterObject[];
//# sourceMappingURL=flatten-deep-object-query-parameter.d.ts.map