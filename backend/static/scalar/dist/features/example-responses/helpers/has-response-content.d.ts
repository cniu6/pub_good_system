import type { ResponseObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Checks if a response object has body content (schema, example, or examples).
 * Looks through common media types in priority order.
 */
export declare function hasResponseContent(response: ResponseObject | undefined, responseKey?: string): boolean;
//# sourceMappingURL=has-response-content.d.ts.map