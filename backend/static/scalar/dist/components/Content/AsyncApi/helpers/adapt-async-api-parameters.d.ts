import type { AsyncApiChannelObject } from '@scalar/types/asyncapi/3.1';
import type { ParameterObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
/**
 * Adapt AsyncAPI channel parameters into the OpenAPI `ParameterObject` shape so we can reuse the
 * existing `ParameterList` rendering instead of building a separate component.
 *
 * AsyncAPI parameters are address placeholders (the `{param}` expressions in a channel address).
 * They are string-only and, because every placeholder in the address must be supplied, always
 * required. We therefore map them to `path` parameters (the closest OpenAPI analog) and fold the
 * `enum`, `default`, and `examples` fields onto a synthetic string schema so the schema renderer
 * can display them. The `location` runtime expression has no display equivalent and is dropped.
 */
export declare const adaptAsyncApiParameters: (parameters: AsyncApiChannelObject["parameters"]) => ParameterObject[];
//# sourceMappingURL=adapt-async-api-parameters.d.ts.map