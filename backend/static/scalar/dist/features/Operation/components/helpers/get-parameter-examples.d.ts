import type { ParameterObject, ResponseObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
type GetParameterExamplesArgs = {
    parameter: ParameterObject | ResponseObject;
    schemaExamples?: unknown[];
    contentExamples?: unknown;
};
/**
 * Build a normalized examples array from parameter/content/schema examples.
 * Undefined values are removed so the UI does not render "undefined" entries.
 */
export declare const getParameterExamples: ({ parameter, schemaExamples, contentExamples, }: GetParameterExamplesArgs) => unknown[];
export {};
//# sourceMappingURL=get-parameter-examples.d.ts.map